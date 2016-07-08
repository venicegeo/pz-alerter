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

package workflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	_ "io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"

	_ "io"
	"net/http"
	_ "net/url"

	pzlogger "github.com/venicegeo/pz-logger/logger"
	pzuuidgen "github.com/venicegeo/pz-uuidgen/uuidgen"
)

//------------------------------------------

type LockedAdminStats struct {
	sync.Mutex
	WorkflowAdminStats
}

type WorkflowService struct {
	eventTypeDB *EventTypeDB
	eventDB     *EventDB
	triggerDB   *TriggerDB
	alertDB     *AlertDB

	stats LockedAdminStats

	logger  pzlogger.IClient
	uuidgen pzuuidgen.IClient

	sys *piazza.SystemConfig
}

var defaultPagination = &piazza.JsonPagination{
	PerPage: 10,
	Page:    0,
	SortBy:  "id",
	Order:   piazza.PaginationOrderAscending,
}

//------------------------------------------

func (service *WorkflowService) Init(
	sys *piazza.SystemConfig,
	logger pzlogger.IClient,
	uuidgen pzuuidgen.IClient,
	eventtypesIndex elasticsearch.IIndex,
	eventsIndex elasticsearch.IIndex,
	triggersIndex elasticsearch.IIndex,
	alertsIndex elasticsearch.IIndex) error {

	service.sys = sys

	service.stats.CreatedOn = time.Now()

	var err error

	service.logger = logger
	service.uuidgen = uuidgen

	service.eventTypeDB, err = NewEventTypeDB(service, eventtypesIndex)
	if err != nil {
		return err
	}

	service.eventDB, err = NewEventDB(service, eventsIndex)
	if err != nil {
		return err
	}

	service.triggerDB, err = NewTriggerDB(service, triggersIndex)
	if err != nil {
		return err
	}

	service.alertDB, err = NewAlertDB(service, alertsIndex)
	if err != nil {
		return err
	}

	return nil
}

func (s *WorkflowService) newIdent() (Ident, error) {
	log.Printf("Service:newIdent()")

	uuid, err := s.uuidgen.GetUuid()
	if err != nil {
		log.Printf("==> err %#v", err)
		return NoIdent, err
	}
	log.Printf("==> %s", uuid)

	return Ident(uuid), nil
}

func (service *WorkflowService) lookupEventTypeNameByEventID(id Ident) (string, error) {
	var mapping string = ""

	types, err := service.eventDB.Esi.GetTypes()
	// log.Printf("types: %v", types)
	if err == nil {
		for _, typ := range types {
			// log.Printf("trying %s\n", typ)
			if service.eventDB.Esi.ItemExists(typ, id.String()) {
				mapping = typ
				break
			}
		}
	} else {
		return "", err
	}

	return mapping, nil
}

func (service *WorkflowService) sendToKafka(jobInstance string, JobID Ident) {
	//log.Printf("***********************\n")
	//log.Printf("%s\n", jobInstance)

	kafkaAddress, err := service.sys.GetAddress(piazza.PzKafka)
	if err != nil {
		log.Fatalf("Kafka-related failure (1): %s", err)
	}

	// Get Space we are running in.   Default to int
	space := os.Getenv("SPACE")
	if space == "" {
		space = "int"
	}

	topic := fmt.Sprintf("Request-Job-%s", space)
	message := jobInstance

	//log.Printf("%s\n", kafkaAddress)
	//log.Printf("%s\n", topic)

	producer, err := sarama.NewSyncProducer([]string{kafkaAddress}, nil)
	if err != nil {
		log.Fatalf("Kafka-related failure (2): %s", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalf("Kafka-related failure (3): %s", err)
		}
	}()

	msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(message), Key: sarama.StringEncoder(JobID)}
	partition, offset, err := producer.SendMessage(msg)
	_ = partition
	_ = offset
	if err != nil {
		//log.Printf("FAILED to send message: %s\n", err)
	} else {
		//log.Printf("> message sent to partition %d at offset %d\n", partition, offset)
	}

	//log.Printf("***********************\n")
}

func (service *WorkflowService) postToPzGatewayJobService(uri string, params map[string]string) (*http.Request, error) {
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

//------------------------------------------

func statusOK(obj interface{}) *piazza.JsonResponse {
	return &piazza.JsonResponse{StatusCode: http.StatusOK, Data: obj}
}

func statusCreated(obj interface{}) *piazza.JsonResponse {
	return &piazza.JsonResponse{StatusCode: http.StatusCreated, Data: obj}
}

func statusBadRequest(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{StatusCode: http.StatusBadRequest, Message: err.Error()}
}

func statusInternalServerError(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{StatusCode: http.StatusInternalServerError, Message: err.Error()}
}

func statusNotFound(id Ident) *piazza.JsonResponse {
	return &piazza.JsonResponse{StatusCode: http.StatusNotFound, Message: string(id)}
}

//------------------------------------------

func (service *WorkflowService) GetAdminStats() *piazza.JsonResponse {
	service.stats.Lock()
	t := service.stats.WorkflowAdminStats
	service.stats.Unlock()
	return statusOK(t)
}

//------------------------------------------

func (service *WorkflowService) GetEvents(c *gin.Context) *piazza.JsonResponse {
	params := piazza.NewQueryParams(c.Request)

	format, err := piazza.NewJsonPagination(params, defaultPagination)
	if err != nil {
		return statusBadRequest(err)
	}

	m, count, err := service.eventDB.GetAll("", format)
	if err != nil {
		return statusBadRequest(err)
	}

	bar := make([]interface{}, len(*m))

	for i, e := range *m {
		bar[i] = e
	}

	format.Count = int(count)

	resp := statusOK(bar)
	resp.Pagination = format
	return resp
}

func (service *WorkflowService) DeleteEventType(c *gin.Context) *piazza.JsonResponse {
	id := Ident(c.Param("id"))
	ok, err := service.eventTypeDB.DeleteByID(Ident(id))
	if err != nil {
		return statusBadRequest(err)
	}
	if !ok {
		return statusNotFound(id)
	}

	return statusOK(nil)
}

func (service *WorkflowService) GetEventType(c *gin.Context) *piazza.JsonResponse {
	id := Ident(c.Param("id"))

	event, err := service.eventTypeDB.GetOne(Ident(id))
	if err != nil {
		return statusBadRequest(err)
	}
	if event == nil {
		return statusNotFound(id)
	}
	return statusOK(event)
}

func (service *WorkflowService) GetAllEventTypes(c *gin.Context) *piazza.JsonResponse {
	params := piazza.NewQueryParams(c.Request)

	format, err := piazza.NewJsonPagination(params, defaultPagination)
	if err != nil {
		return statusBadRequest(err)
	}

	ets, count, err := service.eventTypeDB.GetAll(format)
	if err != nil {
		return statusBadRequest(err)
	}

	bar := make([]interface{}, len(*ets))

	for i, e := range *ets {
		bar[i] = e
	}

	format.Count = int(count)

	resp := statusOK(bar)
	resp.Pagination = format
	return resp
}

func (service *WorkflowService) PostEventType(c *gin.Context) *piazza.JsonResponse {
	eventType := &EventType{}
	err := c.BindJSON(eventType)
	if err != nil {
		return statusBadRequest(err)
	}

	//log.Printf("New EventType with id: %s\n", eventType.EventTypeId)

	eventType.EventTypeId, err = service.newIdent()
	if err != nil {
		return statusBadRequest(err)
	}

	id, err := service.eventTypeDB.PostData(eventType, eventType.EventTypeId)
	if err != nil {
		return statusBadRequest(err)
	}

	//log.Printf("New EventType with id: %s\n", eventType.EventTypeId)

	err = service.eventDB.AddMapping(eventType.Name, eventType.Mapping)
	if err != nil {
		service.eventTypeDB.DeleteByID(id)
		return statusBadRequest(err)
	}

	//log.Printf("EventType Mapping: %s, Name: %s\n", eventType.Mapping, eventType.Name)

	return statusCreated(eventType)
}

//------------------------------------------

func (service *WorkflowService) GetEvent(c *gin.Context) *piazza.JsonResponse {
	// eventType := c.Param("eventType")
	s := c.Param("id")
	id := Ident(s)
	// event, err := server.eventDB.GetOne(eventType, id)
	mapping, err := service.lookupEventTypeNameByEventID(id)
	if err != nil {
		return statusBadRequest(err)
	}

	//log.Printf("The Mapping is:  %s\n", mapping)

	event, err := service.eventDB.GetOne(mapping, id)
	if err != nil {
		return statusBadRequest(err)

	}
	if event == nil {
		return statusNotFound(id)
	}

	return statusOK(event)
}

func (service *WorkflowService) GetAllEvents(c *gin.Context) *piazza.JsonResponse {
	params := piazza.NewQueryParams(c.Request)

	format, err := piazza.NewJsonPagination(params, defaultPagination)
	if err != nil {
		return statusBadRequest(err)
	}

	eventTypeId := c.Query("eventTypeId")

	//log.Printf("FFF %s", eventTypeId)

	query := ""

	// Get the eventTypeName corresponding to the eventTypeId
	if eventTypeId != "" {
		eventType, _ := service.eventTypeDB.GetOne(Ident(eventTypeId))
		query = eventType.Name
	}

	m, count, err := service.eventDB.GetAll(query, format)
	if err != nil {
		return statusBadRequest(err)
	}

	bar := make([]interface{}, len(*m))

	for i, e := range *m {
		bar[i] = e
	}

	format.Count = int(count)

	resp := statusOK(bar)
	resp.Pagination = format
	return resp
}

func (service *WorkflowService) PostEvent(c *gin.Context) *piazza.JsonResponse {
	var event Event
	err := c.BindJSON(&event)
	if err != nil {
		return statusBadRequest(err)
	}

	eventTypeId := event.EventTypeId
	eventType, err := service.eventTypeDB.GetOne(eventTypeId)
	if err != nil {
		return statusBadRequest(err)
	}

	event.EventId, err = service.newIdent()
	if err != nil {
		return statusBadRequest(err)
	}

	_, err = service.eventDB.PostData(eventType.Name, event, event.EventId)
	if err != nil {
		return statusBadRequest(err)
	}

	{
		// Find triggers associated with event
		//log.Printf("Looking for triggers with eventType %s and matching %v", eventType.Name, event.Data)
		triggerIDs, err := service.eventDB.PercolateEventData(eventType.Name, event.Data, event.EventId)
		if err != nil {
			return statusBadRequest(err)
		}

		// For each trigger,  apply the event data and submit job
		var waitGroup sync.WaitGroup

		results := make(map[Ident]*piazza.JsonResponse)

		for _, triggerID := range *triggerIDs {
			waitGroup.Add(1)
			go func(triggerID Ident) {
				defer waitGroup.Done()

				//log.Printf("\ntriggerID: %v\n", triggerID)
				trigger, err := service.triggerDB.GetOne(triggerID)
				if err != nil {
					results[triggerID] = statusBadRequest(err)
					return
				}
				if trigger == nil {
					results[triggerID] = statusNotFound(triggerID)
					return
				}
				// Not the best way to do this, but should disallow Triggers from firing if they
				// don't have the same Eventtype as the Event
				// Would rather have this done via the percolation itself ...
				matches := false
				for _, eventtype_id := range trigger.Condition.EventTypeIds {
					if eventtype_id == eventType.EventTypeId {
						matches = true
						break
					}
				}
				if matches == false {
					return
				}

				// JobID gets sent through Kafka as the key
				Job := trigger.Job
				JobID, err := service.newIdent()
				if err != nil {
					panic(99) // TODO
					//return statusBadRequest(err)
				}

				jobInstance, err := json.Marshal(Job)
				jobString := string(jobInstance)

				//log.Printf("trigger: %v\n", trigger)
				//log.Printf("\tJob: %v\n\n", jobString)

				// Not very robust,  need to find a better way
				for key, value := range event.Data {
					jobString = strings.Replace(jobString, "$"+key, fmt.Sprintf("%v", value), 1)
				}

				//log.Printf("jobInstance: %s\n\n", jobString)

				//server.logger.Info("job submission: %s\n", jobString)

				service.sendToKafka(jobString, JobID)

				// Send alert
				newid, err := service.newIdent()
				if err != nil {
					panic(99) // TODO
				}

				alert := Alert{AlertId: newid, EventId: event.EventId, TriggerId: triggerID, JobId: JobID}

				//log.Printf("alert: id: %s, EventID: %s, TriggerID: %s, JobID: %s", alert.ID, alert.EventID, alert.TriggerID, alert.JobID)

				_, alert_err := service.alertDB.PostData(&alert, alert.AlertId)
				if alert_err != nil {
					results[triggerID] = statusBadRequest(alert_err)
					return
				}

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

		for _, v := range results {
			if v != nil {
				return v
			}
		}
	}

	return statusCreated(event)
}

func (service *WorkflowService) DeleteEvent(c *gin.Context) *piazza.JsonResponse {
	s := c.Param("id")
	id := Ident(s)
	// eventType := c.Param("eventType")
	mapping, err := service.lookupEventTypeNameByEventID(id)
	if err != nil {
		return statusBadRequest(err)
	}

	//log.Printf("The Mapping is:  %s\n", mapping)

	ok, err := service.eventDB.DeleteByID(mapping, Ident(id))
	if err != nil {
		return statusBadRequest(err)
	}
	if !ok {
		return statusNotFound(id)
	}

	return statusOK(nil)
}

//------------------------------------------

func (service *WorkflowService) GetTrigger(c *gin.Context) *piazza.JsonResponse {
	id := Ident(c.Param("id"))

	trigger, err := service.triggerDB.GetOne(Ident(id))
	if err != nil {
		return statusBadRequest(err)
	}
	if trigger == nil {
		return statusNotFound(id)
	}
	return statusOK(trigger)
}

func (service *WorkflowService) GetAllTriggers(c *gin.Context) *piazza.JsonResponse {
	params := piazza.NewQueryParams(c.Request)

	format, err := piazza.NewJsonPagination(params, defaultPagination)
	if err != nil {
		return statusBadRequest(err)
	}

	m, count, err := service.triggerDB.GetAll(format)
	if err != nil {
		return statusBadRequest(err)
	}

	bar := make([]interface{}, len(*m))

	for i, e := range *m {
		bar[i] = e
	}

	format.Count = int(count)

	resp := statusOK(bar)
	resp.Pagination = format
	return resp
}

func (service *WorkflowService) PostTrigger(c *gin.Context) *piazza.JsonResponse {
	trigger := &Trigger{}
	err := c.BindJSON(trigger)
	if err != nil {
		return statusBadRequest(err)
	}

	trigger.TriggerId, err = service.newIdent()
	if err != nil {
		return statusBadRequest(err)
	}

	_, err = service.triggerDB.PostTrigger(trigger, trigger.TriggerId)
	if err != nil {
		return statusBadRequest(err)
	}

	return statusCreated(trigger)
}

func (service *WorkflowService) DeleteTrigger(c *gin.Context) *piazza.JsonResponse {
	id := Ident(c.Param("id"))
	ok, err := service.triggerDB.DeleteTrigger(Ident(id))
	if err != nil {
		return statusBadRequest(err)
	}
	if !ok {
		return statusNotFound(id)
	}

	return statusOK(nil)
}

//------------------------------------------

func (service *WorkflowService) GetAlert(c *gin.Context) *piazza.JsonResponse {
	id := Ident(c.Param("id"))

	alert, err := service.alertDB.GetOne(id)
	if err != nil {
		return statusBadRequest(err)
	}
	if alert == nil {
		return statusNotFound(id)
	}

	return statusOK(alert)
}

func (service *WorkflowService) GetAllAlerts(c *gin.Context) *piazza.JsonResponse {
	// TODO: conditionID := c.Query("condition")
	//log.Printf("%#v", c.Request)
	triggerId := c.Query("triggerId")

	params := piazza.NewQueryParams(c.Request)

	format, err := piazza.NewJsonPagination(params, defaultPagination)
	if err != nil {
		return statusBadRequest(err)
	}

	var all *[]Alert
	var count int64

	if isUuid(triggerId) {
		//log.Printf("Getting alerts with trigger %s", triggerId)
		all, count, err = service.alertDB.GetAllByTrigger(format, triggerId)
		if err != nil {
			return statusBadRequest(err)
		}
	} else if triggerId == "" {
		//log.Printf("Getting all alerts %#v", service)
		all, count, err = service.alertDB.GetAll(format)
		if err != nil {
			return statusBadRequest(err)
		}
	} else { // Malformed triggerId
		return statusBadRequest(errors.New("Malformed triggerId query parameter"))
	}

	//log.Printf("Making bar")
	bar := make([]interface{}, len(*all))

	//log.Printf("Adding values to bar")
	for i, e := range *all {
		bar[i] = e
	}

	format.Count = int(count)

	resp := statusOK(bar)
	resp.Pagination = format
	return resp
}

func (service *WorkflowService) PostAlert(c *gin.Context) *piazza.JsonResponse {
	var alert Alert
	err := c.BindJSON(&alert)
	if err != nil {
		return statusBadRequest(err)
	}

	alert.AlertId, err = service.newIdent()
	if err != nil {
		return statusBadRequest(err)
	}

	_, err = service.alertDB.PostData(&alert, alert.AlertId)
	if err != nil {
		return statusInternalServerError(err)
	}

	return statusCreated(alert)
}

func (service *WorkflowService) DeleteAlert(c *gin.Context) *piazza.JsonResponse {
	id := Ident(c.Param("id"))
	ok, err := service.alertDB.DeleteByID(id)
	if err != nil {
		return statusBadRequest(err)
	}
	if !ok {
		return statusNotFound(id)
	}

	return statusOK(nil)
}
