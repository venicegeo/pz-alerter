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

//------------------------------------------------------------------------------

type LockedAdminStats struct {
	sync.Mutex
	WorkflowAdminStats
	origin string
}

var origin string

type WorkflowService struct {
	eventTypeDB *EventTypeDB
	eventDB     *EventDB
	triggerDB   *TriggerDB
	alertDB     *AlertDB

	stats LockedAdminStats

	logger  pzlogger.IClient
	uuidgen pzuuidgen.IClient

	sys *piazza.SystemConfig

	origin string
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

//------------------------------------------------------------------------------

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

	service.origin = string(sys.Name)

	return nil
}

func (s *WorkflowService) newIdent() (piazza.Ident, error) {
	uuid, err := s.uuidgen.GetUuid()
	if err != nil {
		return piazza.NoIdent, err
	}

	return piazza.Ident(uuid), nil
}

func (service *WorkflowService) sendToKafka(jobInstance string, jobID piazza.Ident) error {
	kafkaAddress, err := service.sys.GetAddress(piazza.PzKafka)
	if err != nil {
		return LoggedError("Kafka-related failure (1): %s", err.Error())
	}

	space := service.sys.Space

	topic := fmt.Sprintf("Request-Job-%s", space)
	message := jobInstance

	producer, err := sarama.NewSyncProducer([]string{kafkaAddress}, nil)
	if err != nil {
		return LoggedError("Kafka-related failure (2): %s", err.Error())
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalf("Kafka-related failure (3): " + err.Error())
		}
	}()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
		Key:   sarama.StringEncoder(jobID)}
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		return LoggedError("Kafka-related failure (4): %s", err.Error())
	}

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

//------------------------------------------------------------------------------

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
	return &piazza.JsonResponse{
		StatusCode: http.StatusBadRequest,
		Message:    err.Error(),
		Origin:     origin,
	}
}

func statusInternalServerError(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusInternalServerError,
		Message:    err.Error(),
		Origin:     origin,
	}
}

func statusNotFound(id piazza.Ident) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusNotFound,
		Message:    string(id),
		Origin:     origin,
	}
}

//------------------------------------------------------------------------------

func (service *WorkflowService) GetAdminStats() *piazza.JsonResponse {
	service.stats.Lock()
	t := service.stats.WorkflowAdminStats
	service.stats.Unlock()
	return statusOK(t)
}

//------------------------------------------------------------------------------

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

	eventtypes, err := service.eventTypeDB.GetAll(format)
	if err != nil {
		return statusBadRequest(err)
	}
	if eventtypes == nil {
		return statusInternalServerError(errors.New("getalleventtypes returned nil"))
	}

	resp := statusOK(eventtypes)

	if len(eventtypes) > 0 {
		format.Count = int(len(eventtypes))
		resp.Pagination = format
	}

	return resp
}

func (service *WorkflowService) PostEventType(eventType *EventType) *piazza.JsonResponse {

	// Check if our EventType.Name already exists
	name := eventType.Name
	if service.eventDB.NameExists(name) {
		id, err := service.eventTypeDB.GetIDByName(name)
		if err != nil {
			return statusInternalServerError(err)
		}
		return statusBadRequest(
			LoggedError("EventType Name already exists under EventTypeId %s", id))
	}

	eventTypeID, err := service.newIdent()
	if err != nil {
		return statusBadRequest(err)
	}
	eventType.EventTypeId = eventTypeID

	eventType.CreatedOn = time.Now()

	id, err := service.eventTypeDB.PostData(eventType, eventTypeID)
	if err != nil {
		return statusBadRequest(err)
	}

	mapping := eventType.Mapping

	err = service.eventDB.AddMapping(name, mapping)
	if err != nil {
		service.eventTypeDB.DeleteByID(id)
		return statusBadRequest(err)
	}

	service.logger.Info("Posted EventType with EventTypeId %s", eventType, eventTypeID)

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

	service.logger.Info("Deleted EventType with EventTypeId %s", id)

	return statusOK(nil)
}

//------------------------------------------------------------------------------

func (service *WorkflowService) GetEvent(id piazza.Ident) *piazza.JsonResponse {
	mapping, err := service.eventDB.lookupEventTypeNameByEventID(id)
	if err != nil {
		return statusNotFound(id)
	}

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
	eventTypeId, err := params.AsString("eventTypeId", nil)
	if err != nil {
		return statusBadRequest(err)
	}
	eventTypeName, err := params.AsString("eventTypeName", nil)
	if err != nil {
		return statusBadRequest(err)
	}

	query := ""

	// Get the eventTypeName corresponding to the eventTypeId
	if eventTypeId != nil {
		eventType, err := service.eventTypeDB.GetOne(piazza.Ident(*eventTypeId))
		if err != nil {
			return statusBadRequest(err)
		}
		query = eventType.Name
	} else if eventTypeName != nil {
		query = *eventTypeName
	}

	events, err := service.eventDB.GetAll(query, format)
	if err != nil {
		return statusBadRequest(err)
	}

	resp := statusOK(events)

	if len(events) > 0 {
		format.Count = int(len(events))
		resp.Pagination = format
	}

	return resp
}

func (service *WorkflowService) PostEvent(event *Event) *piazza.JsonResponse {
	eventTypeID := event.EventTypeId
	eventType, err := service.eventTypeDB.GetOne(eventTypeID)
	if err != nil {
		return statusBadRequest(err)
	}
	eventTypeName := eventType.Name

	eventID, err := service.newIdent()
	if err != nil {
		return statusBadRequest(err)
	}
	event.EventId = eventID

	event.CreatedOn = time.Now()

	_, err = service.eventDB.PostData(eventTypeName, event, eventID)
	if err != nil {
		return statusBadRequest(err)
	}

	service.logger.Info("Posted Event with EventId %s", event, eventID)

	{
		// Find triggers associated with event
		triggerIDs, err := service.eventDB.PercolateEventData(eventTypeName, event.Data, eventID)
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

				trigger, err := service.triggerDB.GetOne(triggerID)
				if err != nil {
					results[triggerID] = statusBadRequest(err)
					return
				}
				if trigger == nil {
					results[triggerID] = statusNotFound(triggerID)
					return
				}
				if trigger.Disabled == true {
					//results[triggerID] = statusOK(triggerID)
					return
				}

				// Not the best way to do this, but should disallow Triggers from firing if they
				// don't have the same Eventtype as the Event
				// Would rather have this done via the percolation itself ...
				matches := false
				for _, eventTypeID := range trigger.Condition.EventTypeIds {
					if eventTypeID == eventType.EventTypeId {
						matches = true
						break
					}
				}
				if matches == false {
					return
				}

				// jobID gets sent through Kafka as the key
				job := trigger.Job
				jobID, err := service.newIdent()
				if err != nil {
					results[triggerID] = statusInternalServerError(err)
					return
				}

				jobInstance, err := json.Marshal(job)
				jobString := string(jobInstance)

				// Not very robust,  need to find a better way
				for key, value := range event.Data {
					jobString = strings.Replace(jobString, "$"+key, fmt.Sprintf("%v", value), 1)
				}

				service.logger.Info("job submission: %s\n", jobString)

				err = service.sendToKafka(jobString, jobID)
				if err != nil {
					results[triggerID] = statusInternalServerError(err)
					return
				}

				alert := Alert{EventId: eventID, TriggerId: triggerID, JobId: jobID}
				resp := service.PostAlert(&alert)
				if resp.IsError() {
					// resp will be a statusInternalServerError or statusBadRequest
					results[triggerID] = resp
					return
				}

			}(triggerID)
		}

		waitGroup.Wait()

		//log.Printf("trigger results: %#v", results)
		for _, v := range results {
			// log.Printf("%#v %#v", k, v)
			if v != nil {
				return v
			}
		}
	}

	return statusCreated(event)
}

func (service *WorkflowService) DeleteEvent(id piazza.Ident) *piazza.JsonResponse {
	mapping, err := service.eventDB.lookupEventTypeNameByEventID(id)
	if err != nil {
		return statusBadRequest(err)
	}

	ok, err := service.eventDB.DeleteByID(mapping, piazza.Ident(id))
	if err != nil {
		return statusBadRequest(err)
	}
	if !ok {
		return statusNotFound(id)
	}

	service.logger.Info("Deleted Event with EventId %s", id)

	return statusOK(nil)
}

//------------------------------------------------------------------------------

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

	triggers, err := service.triggerDB.GetAll(format)
	if err != nil {
		return statusBadRequest(err)
	}
	if triggers == nil {
		return statusInternalServerError(errors.New("getalltriggers returned nil"))
	}

	resp := statusOK(triggers)

	if len(triggers) > 0 {
		format.Count = int(len(triggers))
		resp.Pagination = format
	}

	return resp
}

func (service *WorkflowService) PostTrigger(trigger *Trigger) *piazza.JsonResponse {
	triggerID, err := service.newIdent()
	if err != nil {
		return statusBadRequest(err)
	}
	trigger.TriggerId = triggerID
	trigger.CreatedOn = time.Now()

	_, err = service.triggerDB.PostTrigger(trigger, triggerID)
	if err != nil {
		return statusBadRequest(err)
	}

	service.logger.Info("Posted Trigger with TriggerId %s", trigger, triggerID)

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

	service.logger.Info("Deleted Trigger with TriggerId %s", id)

	return statusOK(nil)
}

//------------------------------------------------------------------------------

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
	triggerID, err := params.AsString("triggerId", nil)
	if err != nil {
		return statusBadRequest(err)
	}

	format, err := piazza.NewJsonPagination(params, defaultAlertPagination)
	if err != nil {
		return statusBadRequest(err)
	}

	var alerts []Alert

	if triggerID != nil && isUuid(*triggerID) {
		alerts, err = service.alertDB.GetAllByTrigger(format, *triggerID)
		if err != nil {
			return statusBadRequest(err)
		}
		if alerts == nil {
			return statusInternalServerError(errors.New("getallalerts returned nil"))
		}
	} else if triggerID == nil {
		alerts, err = service.alertDB.GetAll(format)
		if err != nil {
			return statusBadRequest(err)
		}
		if alerts == nil {
			return statusInternalServerError(errors.New("getallalerts returned nil"))
		}
	} else {
		return statusBadRequest(errors.New("Malformed triggerId query parameter"))
	}

	resp := statusOK(alerts)

	if len(alerts) > 0 {
		format.Count = int(len(alerts))
		resp.Pagination = format
	}

	return resp
}

func (service *WorkflowService) PostAlert(alert *Alert) *piazza.JsonResponse {
	alertID, err := service.newIdent()
	if err != nil {
		return statusBadRequest(err)
	}
	alert.AlertId = alertID

	alert.CreatedOn = time.Now()

	_, err = service.alertDB.PostData(&alert, alertID)
	if err != nil {
		return statusInternalServerError(err)
	}

	service.logger.Info("Posted Alert with AlertId %s", alert, alertID)

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

	service.logger.Info("Deleted Alert with AlertId %s", id)

	return statusOK(nil)
}
