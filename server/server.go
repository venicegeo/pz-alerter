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
	_ "os"

	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
)

type LockedAdminSettings struct {
	sync.Mutex
	WorkflowAdminSettings
}

var settings LockedAdminSettings

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
		Message string
	}
	e := errret{Message: err.Error()}
	c.JSON(http.StatusBadRequest, e)
}

//---------------------------------------------------------------------------

type Server struct {
	eventTypeDB *EventTypeDB
	eventDB     *EventDB
	triggerDB   *TriggerDB
	alertDB     *AlertDB

	uuidgen uuidgenPkg.IUuidGenService
}

var server *Server
var sysConfig *piazza.SystemConfig

func NewServer(
	eventtypesIndex elasticsearch.IIndex,
	eventsIndex elasticsearch.IIndex,
	triggersIndex elasticsearch.IIndex,
	alertsIndex elasticsearch.IIndex,
	uuidgen uuidgenPkg.IUuidGenService) (*Server, error) {
	var s Server
	var err error

	s.uuidgen = uuidgen

	s.eventTypeDB, err = NewEventTypeDB(&s, eventtypesIndex)
	if err != nil {
		return nil, err
	}
	err = s.eventTypeDB.Esi.Flush()
	if err != nil {
		return nil, err
	}

	s.eventDB, err = NewEventDB(&s, eventsIndex)
	if err != nil {
		return nil, err
	}
	err = s.eventDB.Esi.Flush()
	if err != nil {
		return nil, err
	}

	s.triggerDB, err = NewTriggerDB(&s, triggersIndex)
	if err != nil {
		return nil, err
	}
	err = s.triggerDB.Esi.Flush()
	if err != nil {
		return nil, err
	}

	s.alertDB, err = NewAlertDB(&s, alertsIndex)
	if err != nil {
		return nil, err
	}
	err = s.alertDB.Esi.Flush()
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Server) NewIdent() Ident {
	var debugIds = true

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

func handleGetAdminSettings(c *gin.Context) {
	settings.Lock()
	t := settings
	settings.Unlock()
	StatusOK(c, t)
}

func handlePostAdminSettings(c *gin.Context) {
	var s WorkflowAdminSettings
	err := c.BindJSON(&s)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	settings.Lock()
	settings.WorkflowAdminSettings = s
	settings.Unlock()
	StatusOK(c, s)
}

func handlePostAdminShutdown(c *gin.Context) {
	piazza.HandlePostAdminShutdown(c)
}

func handeGetEventByID(c *gin.Context) {
	eventType := c.Param("eventType")
	s := c.Param("id")

	id := Ident(s)
	event, err := server.eventDB.GetOne(eventType, id)
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

	err = server.alertDB.Flush()
	if err != nil {
		StatusBadRequest(c, err)
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

	err = server.alertDB.Flush()
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

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

	all, err := server.alertDB.GetAll()
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	StatusOK(c, all)
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

	err = server.triggerDB.Flush()
	if err != nil {
		StatusBadRequest(c, err)
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
	m, err := server.triggerDB.GetAll()
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	StatusOK(c, m)
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

	err = server.triggerDB.Flush()
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

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

	err = server.eventTypeDB.Flush()
	if err != nil {
		StatusBadRequest(c, err)
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
	ets, err := server.eventTypeDB.GetAll()
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	StatusOK(c, ets)
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

	err = server.eventTypeDB.Flush()
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	StatusCreated(c, retID)
}

func handeDeleteEventByID(c *gin.Context) {
	id := c.Param("id")
	eventType := c.Param("eventType")

	ok, err := server.eventDB.DeleteByID(eventType, Ident(id))
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	if !ok {
		StatusNotFound(c, gin.H{"id": id})
		return
	}

	err = server.eventDB.Flush()
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	StatusOK(c, nil)
}

func handleGetEventsByEventType(c *gin.Context) {
	eventType := c.Param("eventType")

	m, err := server.eventDB.GetAll(eventType)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}
	StatusOK(c, m)
}

func handlePostEvent(c *gin.Context) {

	eventType := c.Param("eventType")

	var event Event
	err := c.BindJSON(&event)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	event.ID = server.NewIdent()
	_, err = server.eventDB.PostData(eventType, event, event.ID)
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	retID := WorkflowIDResponse{ID: event.ID}

	err = server.eventDB.Flush()
	if err != nil {
		StatusBadRequest(c, err)
		return
	}

	{
		// log.Printf("event:\n")
		// log.Printf("\tID: %v\n", event.ID)
		// log.Printf("\tType: %v\n", eventType)
		// log.Printf("\tData: %v\n", event.Data)

		// Find triggers associated with event
		triggerIDs, err := server.eventDB.PercolateEventData(eventType, event.Data, event.ID)
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

				// log.Printf("\ntriggerID: %v\n", triggerID)
				trigger, err := server.triggerDB.GetOne(triggerID)
				if err != nil {
					StatusBadRequest(c, err)
					return
				}
				if trigger == nil {
					StatusNotFound(c, gin.H{"id": triggerID})
					return
				}
				// log.Printf("trigger: %v\n", trigger)
				// log.Printf("\tJob: %v\n\n", trigger.Job.Task)

				var jobInstance = trigger.Job.Task

				//  Not very robust,  need to find a better way
				for key, value := range event.Data {
					jobInstance = strings.Replace(jobInstance, "$"+key, fmt.Sprintf("%v", value), 1)
				}

				log.Printf("jobInstance: %s\n\n", jobInstance)

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

			}(triggerID)
		}

		waitGroup.Wait()
	}

	StatusCreated(c, retID)
}

func handleHealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "Hi. I'm pz-workflow.")
}

func CreateHandlers(sys *piazza.SystemConfig,
	logger *loggerPkg.CustomLogger,
	uuidgen uuidgenPkg.IUuidGenService,
	eventtypesIndex elasticsearch.IIndex,
	eventsIndex elasticsearch.IIndex,
	triggersIndex elasticsearch.IIndex,
	alertsIndex elasticsearch.IIndex) (http.Handler, error) {

	var err error

	sysConfig = sys

	server, err = NewServer(eventtypesIndex, eventsIndex,
		triggersIndex, alertsIndex, uuidgen)
	if err != nil {
		return nil, errors.New("internal error: server context failed to initialize")
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//router.Use(gin.Logger())
	//router.Use(gin.Recovery())

	//---------------------------------------------------------------

	router.GET("/", handleHealthCheck)

	router.POST("/v1/events/:eventType", handlePostEvent)
	///////////////////////	router.GET("/v1/events", handleGetEvents)
	router.GET("/v1/events/:eventType", handleGetEventsByEventType)
	router.GET("/v1/events/:eventType/:id", handeGetEventByID)
	router.DELETE("/v1/events/:eventType/:id", handeDeleteEventByID)

	router.POST("/v1/eventtypes", handlePostEventType)
	router.GET("/v1/eventtypes", handleGetEventTypes)
	router.GET("/v1/eventtypes/:id", handleGetEventTypeByID)
	router.DELETE("/v1/eventtypes/:id", handleDeleteEventTypeByID)

	router.POST("/v1/triggers", handlePostTrigger)
	router.GET("/v1/triggers", handleGetTriggers)
	router.GET("/v1/triggers/:id", handleGetTriggerByID)
	router.DELETE("/v1/triggers/:id", handleDeleteTriggerByID)

	router.POST("/v1/alerts", handlePostAlert)
	router.GET("/v1/alerts", handleGetAlerts)
	router.GET("/v1/alerts/:id", handleGetAlertByID)
	router.DELETE("/v1/alerts/:id", handleDeleteAlertByID)

	router.POST("/v1/admin/settings", handlePostAdminSettings)
	router.POST("/v1/admin/shutdown", handlePostAdminShutdown)
	router.GET("/v1/admin/stats", handleGetAdminStats)
	router.GET("/v1/admin/settings", handleGetAdminSettings)

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
