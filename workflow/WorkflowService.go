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
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
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

var defaultEventTypePagination = &piazza.JsonPagination{
	PerPage: 50,
	Page:    0,
	SortBy:  "eventTypeId",
	Order:   piazza.PaginationOrderAscending,
}

var defaultEventPagination = &piazza.JsonPagination{
	PerPage: 50,
	Page:    0,
	SortBy:  "eventId",
	Order:   piazza.PaginationOrderAscending,
}

var defaultTriggerPagination = &piazza.JsonPagination{
	PerPage: 50,
	Page:    0,
	SortBy:  "triggerId",
	Order:   piazza.PaginationOrderAscending,
}

var defaultAlertPagination = &piazza.JsonPagination{
	PerPage: 50,
	Page:    0,
	SortBy:  "alertId",
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

func (s *WorkflowService) newIdent() (piazza.Ident, error) {
	uuid, err := s.uuidgen.GetUuid()
	if err != nil {
		return piazza.NoIdent, err
	}

	return piazza.Ident(uuid), nil
}

func (service *WorkflowService) lookupEventTypeNameByEventID(id piazza.Ident) (string, error) {
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

func (service *WorkflowService) sendToKafka(jobInstance string, jobID piazza.Ident) error {
	//log.Printf("***********************\n")
	//log.Printf("%s\n", jobInstance)

	kafkaAddress, err := service.sys.GetAddress(piazza.PzKafka)
	if err != nil {
		return errors.New("Kafka-related failure (1): " + err.Error())
	}

	space := service.sys.Space

	topic := fmt.Sprintf("Request-Job-%s", space)
	message := jobInstance

	//log.Printf("%s\n", kafkaAddress)
	//log.Printf("%s\n", topic)

	producer, err := sarama.NewSyncProducer([]string{kafkaAddress}, nil)
	if err != nil {
		return errors.New("Kafka-related failure (2): " + err.Error())
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalf("Kafka-related failure (3): " + err.Error())
		}
	}()

	msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(message), Key: sarama.StringEncoder(jobID)}
	partition, offset, err := producer.SendMessage(msg)
	_ = partition
	_ = offset
	if err != nil {
		return errors.New("Kafka-related failure (4): " + err.Error())
	} else {
		//log.Printf("> message sent to partition %d at offset %d\n", partition, offset)
	}

	//log.Printf("***********************\n")

	return nil
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

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	return req, err
}

//------------------------------------------

func statusOK(obj interface{}) *piazza.JsonResponse {
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: obj}
	err := resp.SetType()
	if err != nil {
		return statusInternalServerError(err)
	}
	return resp
}

func statusCreated(obj interface{}) *piazza.JsonResponse {
	resp := &piazza.JsonResponse{StatusCode: http.StatusCreated, Data: obj}
	err := resp.SetType()
	if err != nil {
		return statusInternalServerError(err)
	}
	return resp
}

func statusBadRequest(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{StatusCode: http.StatusBadRequest, Message: err.Error()}
}

func statusInternalServerError(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{StatusCode: http.StatusInternalServerError, Message: err.Error()}
}

func statusNotFound(id piazza.Ident) *piazza.JsonResponse {
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

func (service *WorkflowService) GetEventType(id piazza.Ident) *piazza.JsonResponse {

	event, err := service.eventTypeDB.GetOne(piazza.Ident(id))
	if err != nil {
		return statusNotFound(id)
	}
	if event == nil {
		return statusNotFound(id)
	}
	return statusOK(event)
}

func (service *WorkflowService) GetAllEventTypes(params *piazza.HttpQueryParams) *piazza.JsonResponse {

	format, err := piazza.NewJsonPagination(params, defaultEventTypePagination)
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

func (service *WorkflowService) PostEventType(eventType *EventType) *piazza.JsonResponse {
	var err error

	eventType.EventTypeId, err = service.newIdent()
	if err != nil {
		return statusBadRequest(err)
	}
	log.Printf("New EventType/1: %#v", eventType)

	eventType.CreatedOn = time.Now()

	id, err := service.eventTypeDB.PostData(eventType, eventType.EventTypeId)
	if err != nil {
		return statusBadRequest(err)
	}

	log.Printf("New EventType/2: %#v", eventType)

	err = service.eventDB.AddMapping(eventType.Name, eventType.Mapping)
	if err != nil {
		service.eventTypeDB.DeleteByID(id)
		return statusBadRequest(err)
	}

	log.Printf("EventType Mapping: %s, Name: %s\n", eventType.Mapping, eventType.Name)

	return statusCreated(eventType)
}

func (service *WorkflowService) DeleteEventType(id piazza.Ident) *piazza.JsonResponse {
	ok, err := service.eventTypeDB.DeleteByID(piazza.Ident(id))
	if err != nil {
		return statusBadRequest(err)
	}
	if !ok {
		return statusNotFound(id)
	}

	return statusOK(nil)
}

//------------------------------------------

func (service *WorkflowService) GetEvent(id piazza.Ident) *piazza.JsonResponse {
	// eventType := c.Param("eventType")
	// event, err := server.eventDB.GetOne(eventType, id)
	mapping, err := service.lookupEventTypeNameByEventID(id)
	if err != nil {
		return statusNotFound(id)
	}

	//log.Printf("The Mapping is:  %s\n", mapping)

	event, err := service.eventDB.GetOne(mapping, id)
	if err != nil {
		return statusNotFound(id)

	}
	if event == nil {
		return statusNotFound(id)
	}

	return statusOK(event)
}

func (service *WorkflowService) GetAllEvents(params *piazza.HttpQueryParams) *piazza.JsonResponse {

	format, err := piazza.NewJsonPagination(params, defaultEventPagination)
	if err != nil {
		return statusBadRequest(err)
	}

	// if both specified, "by id"" wins
	eventTypeId := params.Get("eventTypeId")
	eventTypeName := params.Get("eventTypeName")

	query := ""

	// Get the eventTypeName corresponding to the eventTypeId
	if eventTypeId != "" {
		eventType, err := service.eventTypeDB.GetOne(piazza.Ident(eventTypeId))
		if err != nil {
			return statusBadRequest(err)
		}
		query = eventType.Name
	} else if eventTypeName != "" {
		query = eventTypeName
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

func (service *WorkflowService) PostEvent(event *Event) *piazza.JsonResponse {

	eventTypeId := event.EventTypeId
	eventType, err := service.eventTypeDB.GetOne(eventTypeId)
	if err != nil {
		return statusBadRequest(err)
	}

	event.EventId, err = service.newIdent()
	if err != nil {
		return statusBadRequest(err)
	}

	event.CreatedOn = time.Now()

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

		results := make(map[piazza.Ident]*piazza.JsonResponse)

		for _, triggerID := range *triggerIDs {
			waitGroup.Add(1)
			go func(triggerID piazza.Ident) {
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
				if trigger.Disabled == 1 {
					results[triggerID] = statusOK(triggerID)
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
					results[triggerID] = statusInternalServerError(err)
					return
				}

				jobInstance, err := json.Marshal(Job)
				jobString := string(jobInstance)

				log.Printf("trigger: %v\n", trigger)
				log.Printf("\tJob: %v\n\n", jobString)

				// Not very robust,  need to find a better way
				for key, value := range event.Data {
					jobString = strings.Replace(jobString, "$"+key, fmt.Sprintf("%v", value), 1)
				}

				log.Printf("jobInstance: %s\n\n", jobString)

				service.logger.Info("job submission: %s\n", jobString)

				err = service.sendToKafka(jobString, JobID)
				if err != nil {
					results[triggerID] = statusInternalServerError(err)
					return
				}

				// TODO: should really just call service.PostAlert()
				err = service.sendAlert(event.EventId, triggerID, JobID)
				if err != nil {
					results[triggerID] = statusInternalServerError(err)
					return
				}

			}(triggerID)
		}

		waitGroup.Wait()

		//log.Printf("trigger results: %#v", results)
		for _, v := range results {
			if v != nil {
				return v
			}
		}
	}

	return statusCreated(event)
}

func (service *WorkflowService) sendAlert(
	eventId piazza.Ident,
	triggerId piazza.Ident,
	jobId piazza.Ident) error {
	// Send alert
	newid, err := service.newIdent()
	if err != nil {
		return err
	}

	alert := Alert{AlertId: newid, EventId: eventId, TriggerId: triggerId, JobId: jobId}
	alert.CreatedOn = time.Now()

	log.Printf("Alert issued: %#v", alert)

	_, alert_err := service.alertDB.PostData(&alert, alert.AlertId)
	if alert_err != nil {
		return err
	}

	return nil
}

func (service *WorkflowService) DeleteEvent(id piazza.Ident) *piazza.JsonResponse {
	// eventType := c.Param("eventType")
	mapping, err := service.lookupEventTypeNameByEventID(id)
	if err != nil {
		return statusBadRequest(err)
	}

	//log.Printf("The Mapping is:  %s\n", mapping)

	ok, err := service.eventDB.DeleteByID(mapping, piazza.Ident(id))
	if err != nil {
		return statusBadRequest(err)
	}
	if !ok {
		return statusNotFound(id)
	}

	return statusOK(nil)
}

//------------------------------------------

func (service *WorkflowService) GetTrigger(id piazza.Ident) *piazza.JsonResponse {

	trigger, err := service.triggerDB.GetOne(piazza.Ident(id))
	if err != nil {
		return statusNotFound(id)
	}
	if trigger == nil {
		return statusNotFound(id)
	}
	return statusOK(trigger)
}

func (service *WorkflowService) GetAllTriggers(params *piazza.HttpQueryParams) *piazza.JsonResponse {

	format, err := piazza.NewJsonPagination(params, defaultTriggerPagination)
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

func (service *WorkflowService) PostTrigger(trigger *Trigger) *piazza.JsonResponse {
	var err error

	trigger.TriggerId, err = service.newIdent()
	if err != nil {
		return statusBadRequest(err)
	}
	trigger.CreatedOn = time.Now()

	_, err = service.triggerDB.PostTrigger(trigger, trigger.TriggerId)
	if err != nil {
		return statusBadRequest(err)
	}

	return statusCreated(trigger)
}

func (service *WorkflowService) DeleteTrigger(id piazza.Ident) *piazza.JsonResponse {

	ok, err := service.triggerDB.DeleteTrigger(piazza.Ident(id))
	if err != nil {
		return statusBadRequest(err)
	}
	if !ok {
		return statusNotFound(id)
	}

	return statusOK(nil)
}

//------------------------------------------

func (service *WorkflowService) GetAlert(id piazza.Ident) *piazza.JsonResponse {

	alert, err := service.alertDB.GetOne(id)
	if err != nil {
		return statusNotFound(id)
	}
	if alert == nil {
		return statusNotFound(id)
	}

	return statusOK(alert)
}

func (service *WorkflowService) GetAllAlerts(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	triggerId := params.Get("triggerId")

	format, err := piazza.NewJsonPagination(params, defaultAlertPagination)
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

func (service *WorkflowService) PostAlert(alert *Alert) *piazza.JsonResponse {
	var err error

	alert.AlertId, err = service.newIdent()
	if err != nil {
		return statusBadRequest(err)
	}

	alert.CreatedOn = time.Now()

	_, err = service.alertDB.PostData(&alert, alert.AlertId)
	if err != nil {
		return statusInternalServerError(err)
	}

	return statusCreated(alert)
}

func (service *WorkflowService) DeleteAlert(id piazza.Ident) *piazza.JsonResponse {
	ok, err := service.alertDB.DeleteByID(id)
	if err != nil {
		return statusBadRequest(err)
	}
	if !ok {
		return statusNotFound(id)
	}

	return statusOK(nil)
}
