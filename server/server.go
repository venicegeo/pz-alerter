// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	_ "io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"bytes"
	_ "io"
	"mime/multipart"
	_ "net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/Shopify/sarama"
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	loggerPkg "github.com/venicegeo/pz-logger/lib"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
)

type LockedAdminStats struct {
	sync.Mutex
	WorkflowAdminStats
}

var stats LockedAdminStats

func init() {
	stats.Date = time.Now()
}

func StatusOK(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusOK, obj)
}

func StatusCreated(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusCreated, obj)
}

func StatusNotFound(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusNotFound, obj)
}

func StatusBadRequest(c *gin.Context, err error) {

	type errret struct {
		Status    int    `json:"status"`
		Error     string `json:"error"`
		Message   string `json:"message"`
		Timestamp int64  `json:"timestamp"`
		Path      string `json:"path"`
	}
	e := errret{
		Status:    400,
		Error:     "Bad Request",
		Message:   err.Error(),
		Timestamp: time.Now().Unix(),
		Path:      c.Request.URL.Path,
	}

	c.JSON(http.StatusBadRequest, e)

}

//---------------------------------------------------------------------------

type Server struct {
	eventTypeDB *EventTypeDB
	eventDB     *EventDB
	triggerDB   *TriggerDB
	alertDB     *AlertDB

	uuidgen uuidgenPkg.IUuidGenService
	logger  loggerPkg.IClient
}

var server *Server
var sysConfig *piazza.SystemConfig

func NewServer(
	eventtypesIndex elasticsearch.IIndex,
	eventsIndex elasticsearch.IIndex,
	triggersIndex elasticsearch.IIndex,
	alertsIndex elasticsearch.IIndex,
	uuidgen uuidgenPkg.IUuidGenService,
	logger loggerPkg.IClient) (*Server, error) {
	var s Server
	var err error

	s.uuidgen = uuidgen
	s.logger = logger

	s.eventTypeDB, err = NewEventTypeDB(&s, eventtypesIndex)
	if err != nil {
		return nil, err
	}

	s.eventDB, err = NewEventDB(&s, eventsIndex)
	if err != nil {
		return nil, err
	}

	s.triggerDB, err = NewTriggerDB(&s, triggersIndex)
	if err != nil {
		return nil, err
	}

	s.alertDB, err = NewAlertDB(&s, alertsIndex)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Server) NewIdent() Ident {
	var debugIds = false

	var uuid string
	var err error

	if debugIds {
		uuid, err = s.uuidgen.GetDebugUuid("W")
		if err != nil {
			panic("uuidgen failed")
		}
	} else {
		uuid, err = s.uuidgen.GetUuid()
		if err != nil {
			panic("uuidgen failed")
		}
		// log.Printf("uuid: %s", uuid)
	}
	return Ident(uuid)
}

//---------------------------------------------------------------------------

func handleGetAdminStats(c *gin.Context) {
	stats.Lock()
	t := stats.WorkflowAdminStats
	stats.Unlock()
	StatusOK(c, t)
}

func handleGetEvents(c *gin.Context) {
	eventTypeId := c.Query("eventTypeId")
	if eventTypeId == "" {
		format := elasticsearch.GetFormatParams(c, 10, 0, "id", elasticsearch.SortAscending)

		m, err := server.eventDB.GetAll("", format)
		if err != nil {
			StatusBadRequest(c, err)
			return
		}

		StatusOK(c, m)
	} else {
		handleGetEventsByEventType(c)
	}
}

func handleGetEventsV2(c *gin.Context) {
	format := elasticsearch.GetFormatParamsV2(c, 10, 0, "id", elasticsearch.SortAscending)

	m, count, err := server.eventDB.GetAllWithCount("", format)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	bar := make([]interface{}, len(*m))

	for i, e := range *m {
		bar[i] = e
	}

	var order string

	if format.Order {
		order = "desc"
	} else {
		order = "asc"
	}

	foo := &piazza.Common18FListResponse{
		Data: bar,
		Pagination: piazza.Pagination{
			Page:    format.From,
			PerPage: format.Size,
			Count:   count,
			SortBy:  format.Key,
			Order:   order,
		},
	}

	StatusOK(c, foo)
}

func lookupEventTypeNameByEventID(id Ident) (string, error) {
	var mapping string = ""

	types, err := server.eventDB.Esi.GetTypes()
	// log.Printf("types: %v", types)
	if err == nil {
		for _, typ := range types {
			// log.Printf("trying %s\n", typ)
			if server.eventDB.Esi.ItemExists(typ, id.String()) {
				mapping = typ
				break
			}
		}
	} else {
		return "", err
	}

	return mapping, nil
}
func handleGetEventByID(c *gin.Context) {
	// eventType := c.Param("eventType")
	s := c.Param("id")
	id := Ident(s)
	// event, err := server.eventDB.GetOne(eventType, id)
	mapping, err := lookupEventTypeNameByEventID(id)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	log.Printf("The Mapping is:  %s\n", mapping)

	event, err := server.eventDB.GetOne(mapping, id)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	if event == nil {
		StatusNotFound(c, gin.H{"id": id})
		return
	}
	StatusOK(c, event)
}

func handleDeleteAlertByID(c *gin.Context) {
	id := c.Param("id")
	ok, err := server.alertDB.DeleteByID(Ident(id))
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	if !ok {
		StatusNotFound(c, gin.H{"id": id})
		return
	}

	StatusOK(c, nil)
}

func handlePostAlert(c *gin.Context) {
	var alert Alert
	err := c.BindJSON(&alert)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	alert.ID = server.NewIdent()

	id, err := server.alertDB.PostData(&alert, alert.ID)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	retID := WorkflowIDResponse{ID: id}

	StatusCreated(c, retID)
}

func handleGetAlertByID(c *gin.Context) {
	id := c.Param("id")

	alert, err := server.alertDB.GetOne(Ident(id))
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	if alert == nil {
		StatusNotFound(c, gin.H{"id": id})
		return
	}

	StatusOK(c, alert)
}

func handleGetAlerts(c *gin.Context) {
	// TODO: conditionID := c.Query("condition")

	format := elasticsearch.GetFormatParams(c, 10, 0, "id", elasticsearch.SortAscending)
	all, err := server.alertDB.GetAll(format)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	StatusOK(c, all)
}

func handleGetAlertsV2(c *gin.Context) {
	// TODO: conditionID := c.Query("condition")

	format := elasticsearch.GetFormatParamsV2(c, 10, 0, "id", elasticsearch.SortAscending)
	all, count, err := server.alertDB.GetAllWithCount(format)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	bar := make([]interface{}, len(*all))

	for i, e := range *all {
		bar[i] = e
	}

	var order string

	if format.Order {
		order = "desc"
	} else {
		order = "asc"
	}

	foo := &piazza.Common18FListResponse{
		Data: bar,
		Pagination: piazza.Pagination{
			Page:    format.From,
			PerPage: format.Size,
			Count:   count,
			SortBy:  format.Key,
			Order:   order,
		},
	}

	StatusOK(c, foo)
}

func handleDeleteTriggerByID(c *gin.Context) {
	id := c.Param("id")
	ok, err := server.triggerDB.DeleteTrigger(Ident(id))
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	if !ok {
		StatusNotFound(c, gin.H{"id": id})
		return
	}

	StatusOK(c, nil)
}

func handleGetTriggerByID(c *gin.Context) {
	id := c.Param("id")

	trigger, err := server.triggerDB.GetOne(Ident(id))
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	if trigger == nil {
		StatusNotFound(c, gin.H{"id": id})
		return
	}
	StatusOK(c, trigger)
}

func handleGetTriggers(c *gin.Context) {
	format := elasticsearch.GetFormatParams(c, 10, 0, "id", elasticsearch.SortAscending)

	m, err := server.triggerDB.GetAll(format)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	StatusOK(c, m)
}

func handleGetTriggersV2(c *gin.Context) {
	format := elasticsearch.GetFormatParamsV2(c, 10, 0, "id", elasticsearch.SortAscending)

	m, count, err := server.triggerDB.GetAllWithCount(format)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	bar := make([]interface{}, len(*m))

	for i, e := range *m {
		bar[i] = e
	}

	var order string

	if format.Order {
		order = "desc"
	} else {
		order = "asc"
	}

	foo := &piazza.Common18FListResponse{
		Data: bar,
		Pagination: piazza.Pagination{
			Page:    format.From,
			PerPage: format.Size,
			Count:   count,
			SortBy:  format.Key,
			Order:   order,
		},
	}

	StatusOK(c, foo)
}

func handlePostTrigger(c *gin.Context) {
	trigger := &Trigger{}
	err := c.BindJSON(trigger)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	trigger.ID = server.NewIdent()

	_, err = server.triggerDB.PostTrigger(trigger, trigger.ID)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	a := WorkflowIDResponse{ID: trigger.ID}

	StatusCreated(c, a)
}

func handleDeleteEventTypeByID(c *gin.Context) {
	id := c.Param("id")
	ok, err := server.eventTypeDB.DeleteByID(Ident(id))
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	if !ok {
		StatusNotFound(c, gin.H{"id": id})
		return
	}

	StatusOK(c, nil)
}

func handleGetEventTypeByID(c *gin.Context) {
	id := c.Param("id")

	event, err := server.eventTypeDB.GetOne(Ident(id))
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	if event == nil {
		StatusNotFound(c, gin.H{"id": id})
		return
	}
	StatusOK(c, event)
}

func handleGetEventTypes(c *gin.Context) {
	format := elasticsearch.GetFormatParams(c, 10, 0, "id", elasticsearch.SortAscending)

	ets, err := server.eventTypeDB.GetAll(format)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	StatusOK(c, ets)
}

func handleGetEventTypesV2(c *gin.Context) {
	format := elasticsearch.GetFormatParamsV2(c, 10, 0, "id", elasticsearch.SortAscending)

	ets, count, err := server.eventTypeDB.GetAllWithCount(format)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	bar := make([]interface{}, len(*ets))

	for i, e := range *ets {
		bar[i] = e
	}

	var order string

	if format.Order {
		order = "desc"
	} else {
		order = "asc"
	}

	foo := &piazza.Common18FListResponse{
		Data: bar,
		Pagination: piazza.Pagination{
			Page:    format.From,
			PerPage: format.Size,
			Count:   count,
			SortBy:  format.Key,
			Order:   order,
		},
	}

	StatusOK(c, foo)
}

func handlePostEventType(c *gin.Context) {
	eventType := &EventType{}
	err := c.BindJSON(eventType)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	eventType.ID = server.NewIdent()
	id, err := server.eventTypeDB.PostData(eventType, eventType.ID)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	err = server.eventDB.AddMapping(eventType.Name, eventType.Mapping)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	retID := WorkflowIDResponse{ID: id}

	StatusCreated(c, retID)
}

func handleDeleteEventByID(c *gin.Context) {
	s := c.Param("id")
	id := Ident(s)
	// eventType := c.Param("eventType")
	mapping, err := lookupEventTypeNameByEventID(id)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	log.Printf("The Mapping is:  %s\n", mapping)

	ok, err := server.eventDB.DeleteByID(mapping, Ident(id))
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	if !ok {
		StatusNotFound(c, gin.H{"id": id})
		return
	}

	StatusOK(c, nil)
}

func handleGetEventsByEventType(c *gin.Context) {

	format := elasticsearch.GetFormatParams(c, 10, 0, "id", elasticsearch.SortAscending)

	// eventType := c.Param("eventType")
	eventTypeId := c.Query("eventTypeId")

	m, err := server.eventDB.GetAll(eventTypeId, format)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	StatusOK(c, m)
}

func handleGetEventsByEventTypeV2(c *gin.Context) {
	format := elasticsearch.GetFormatParamsV2(c, 10, 0, "id", elasticsearch.SortAscending)

	eventType := c.Param("eventType")

	m, count, err := server.eventDB.GetAllWithCount(eventType, format)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	bar := make([]interface{}, len(*m))

	for i, e := range *m {
		bar[i] = e
	}

	var order string

	if format.Order {
		order = "desc"
	} else {
		order = "asc"
	}

	foo := &piazza.Common18FListResponse{
		Data: bar,
		Pagination: piazza.Pagination{
			Page:    format.From,
			PerPage: format.Size,
			Count:   count,
			SortBy:  format.Key,
			Order:   order,
		},
	}

	StatusOK(c, foo)
}

func handlePostEvent(c *gin.Context) {
	// log.Printf("---------------------\n")

	// eventType := c.Param("eventType")

	var event Event
	err := c.BindJSON(&event)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	eventTypeId := event.EventTypeID
	eventType, err := server.eventTypeDB.GetOne(eventTypeId)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	event.ID = server.NewIdent()
	_, err = server.eventDB.PostData(eventType.Name, event, event.ID)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	retID := WorkflowIDResponse{ID: event.ID}

	{
		// log.Printf("event:\n")
		// log.Printf("\tID: %v\n", event.ID)
		// log.Printf("\tType: %v\n", eventType)
		// log.Printf("\tData: %v\n", event.Data)

		// Find triggers associated with event
		triggerIDs, err := server.eventDB.PercolateEventData(eventType.Name, event.Data, event.ID)
		if err != nil {
			StatusBadRequest(c, err)
			return
		}

		// For each trigger,  apply the event data and submit job
		var waitGroup sync.WaitGroup

		for _, triggerID := range *triggerIDs {
			waitGroup.Add(1)
			go func(triggerID Ident) {
				defer waitGroup.Done()

				log.Printf("\ntriggerID: %v\n", triggerID)
				trigger, err := server.triggerDB.GetOne(triggerID)
				if err != nil {
					StatusBadRequest(c, err)
					return
				}
				if trigger == nil {
					StatusNotFound(c, gin.H{"id": triggerID})
					return
				}

				// Give the job the Ident
				Job := trigger.Job
				Job.JobID = server.NewIdent()

				jobInstance, err := json.Marshal(Job)
				jobString := string(jobInstance)

				log.Printf("trigger: %v\n", trigger)
				log.Printf("\tJob: %v\n\n", jobString)

				// Not very robust,  need to find a better way
				for key, value := range event.Data {
					jobString = strings.Replace(jobString, "$"+key, fmt.Sprintf("%v", value), 1)
				}

				log.Printf("jobInstance: %s\n\n", jobString)

				server.logger.Info("job submission: %s\n", jobString)

				sendToKafka(jobString)

				// Send alert
				var alert Alert
				err := c.BindJSON(&alert)
				if err != nil {
					StatusBadRequest(c, err)
					return
				}

				alert.ID = server.NewIdent()

				id, err := server.alertDB.PostData(&alert, alert.ID)
				if err != nil {
					StatusBadRequest(c, err)
					return
				}

				alert.TriggerID = triggerID
				alert.EventID = event.ID
				alert.JobID = Job.JobID

				retID := WorkflowIDResponse{ID: id}

				StatusCreated(c, retID)

				/**
				// Figure out how to post the jobInstance to job manager server.
				url, err := sysConfig.GetURL(piazza.PzGateway)
				if err != nil {
					StatusBadRequest(c, err)
					return
				}
				gatewayURL := fmt.Sprintf("%s/job", url)
				extraParams := map[string]string{
					"body": jobInstance,
				}

				request, err := postToPzGatewayJobService(gatewayURL, extraParams)
				if err != nil {
					log.Fatal(err)
				}

				client := &http.Client{}
				log.Printf(request.URL.String())
				resp, err := client.Do(request)
				if err != nil {
					log.Fatal(err)
				} else {
					body := &bytes.Buffer{}
					_, err := body.ReadFrom(resp.Body)
					if err != nil {
						log.Fatal(err)
					}
					resp.Body.Close()
					log.Println(resp.StatusCode)
					//    log.Println(resp.Header)
					log.Println(body)
				}
				**/

			}(triggerID)
		}

		waitGroup.Wait()
	}

	StatusCreated(c, retID)
	// log.Printf("---------------------\n")
}

func sendToKafka(jobInstance string)  {
	log.Printf("***********************\n")
	log.Printf("%s\n", jobInstance)

	kafkaAddress, err := sysConfig.GetAddress(piazza.PzKafka)
	if err != nil {
		// return err
		log.Printf("%v\n", err)
	}

	// Get Space we are running in.   Default to int
	space := os.Getenv("SPACE")
	if space == "" {
		space = "int"
	}

	topic := fmt.Sprintf("Request-Job-%s", space)
	message := jobInstance

	log.Printf("%s\n", kafkaAddress)
	log.Printf("%s\n", topic)

	producer, err := sarama.NewSyncProducer([]string{kafkaAddress}, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(message)}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Printf("FAILED to send message: %s\n", err)
	} else {
		log.Printf("> message sent to partition %d at offset %d\n", partition, offset)
	}

	log.Printf("***********************\n")
}

func handleHealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "Hi. I'm pz-workflow.")
}

func CreateHandlers(sys *piazza.SystemConfig,
	logger loggerPkg.IClient,
	uuidgen uuidgenPkg.IUuidGenService,
	eventtypesIndex elasticsearch.IIndex,
	eventsIndex elasticsearch.IIndex,
	triggersIndex elasticsearch.IIndex,
	alertsIndex elasticsearch.IIndex) (http.Handler, error) {

	var err error

	sysConfig = sys

	server, err = NewServer(eventtypesIndex, eventsIndex,
		triggersIndex, alertsIndex, uuidgen, logger)
	if err != nil {
		return nil, errors.New("internal error: server context failed to initialize")
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//router.Use(gin.Logger())
	//router.Use(gin.Recovery())

	//---------------------------------------------------------------

	router.GET("/", handleHealthCheck)

	router.POST("/v1/events", handlePostEvent)
	router.POST("/v2/event", handlePostEvent)

	router.GET("/v1/events", handleGetEvents)
	router.GET("/v2/event", handleGetEventsV2)

/*
	router.GET("/v1/events/:eventType", handleGetEventsByEventType)
	router.GET("/v2/event/:eventType", handleGetEventsByEventTypeV2)
*/

	router.GET("/v1/events/:id", handleGetEventByID)
	router.GET("/v2/event/:id", handleGetEventByID)

/*
	router.DELETE("/v1/events/:eventType/:id", handeDeleteEventByID)
	router.DELETE("/v2/event/:eventType/:id", handeDeleteEventByID)
*/
	router.DELETE("/v1/events/:id", handleDeleteEventByID)
	router.DELETE("/v2/event/:id", handleDeleteEventByID)

	router.POST("/v1/eventtypes", handlePostEventType)
	router.POST("/v2/eventType", handlePostEventType)

	router.GET("/v1/eventtypes", handleGetEventTypes)
	router.GET("/v2/eventType", handleGetEventTypesV2)

	router.GET("/v1/eventtypes/:id", handleGetEventTypeByID)
	router.GET("/v2/eventType/:id", handleGetEventTypeByID)

	router.DELETE("/v1/eventtypes/:id", handleDeleteEventTypeByID)
	router.DELETE("/v2/eventType/:id", handleDeleteEventTypeByID)

	router.POST("/v1/triggers", handlePostTrigger)
	router.POST("/v2/trigger", handlePostTrigger)

	router.GET("/v1/triggers", handleGetTriggers)
	router.GET("/v2/trigger", handleGetTriggersV2)

	router.GET("/v1/triggers/:id", handleGetTriggerByID)
	router.GET("/v2/trigger/:id", handleGetTriggerByID)

	router.DELETE("/v1/triggers/:id", handleDeleteTriggerByID)
	router.DELETE("/v2/trigger/:id", handleDeleteTriggerByID)

	router.POST("/v1/alerts", handlePostAlert)
	router.POST("/v2/alert", handlePostAlert)

	router.GET("/v1/alerts", handleGetAlerts)
	router.GET("/v2/alert", handleGetAlertsV2)

	router.GET("/v1/alerts/:id", handleGetAlertByID)
	router.GET("/v2/alert/:id", handleGetAlertByID)

	router.DELETE("/v1/alerts/:id", handleDeleteAlertByID)
	router.DELETE("/v2/alert/:id", handleDeleteAlertByID)

	router.GET("/v1/admin/stats", handleGetAdminStats)
	router.GET("/v2/admin/stats", handleGetAdminStats)

	logger.Info("handlers set")

	return router, nil
}

func postToPzGatewayJobService(uri string, params map[string]string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req, error := http.NewRequest("POST", uri, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	return req, error
}
