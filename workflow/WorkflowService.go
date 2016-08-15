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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	cron "github.com/vegertar/cron"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	pzlogger "github.com/venicegeo/pz-logger/logger"
	pzuuidgen "github.com/venicegeo/pz-uuidgen/uuidgen"
)

//------------------------------------------------------------------------------

// WorkflowService TODO
type WorkflowService struct {
	eventTypeDB *EventTypeDB
	eventDB     *EventDB
	triggerDB   *TriggerDB
	alertDB     *AlertDB
	cronDB      *CronDB

	stats workflowStats

	logger  pzlogger.IClient
	uuidgen pzuuidgen.IClient

	sys *piazza.SystemConfig

	cron *cron.Cron

	origin string
}

//------------------------------------------------------------------------------

// Init TODO
func (service *WorkflowService) Init(
	sys *piazza.SystemConfig,
	logger pzlogger.IClient,
	uuidgen pzuuidgen.IClient,
	eventtypesIndex elasticsearch.IIndex,
	eventsIndex elasticsearch.IIndex,
	triggersIndex elasticsearch.IIndex,
	alertsIndex elasticsearch.IIndex,
	cronIndex elasticsearch.IIndex) error {

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

	service.cronDB, err = NewCronDB(service, cronIndex)
	if err != nil {
		return err
	}

	service.cron = cron.New()

	// allow the database time to settle
	//time.Sleep(time.Second * 5)
	pollingFn := elasticsearch.GetData(func() (bool, error) {
		exists, err := eventtypesIndex.IndexExists()
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
		types, err := eventtypesIndex.GetTypes()
		if err != nil {
			return false, err
		}
		//log.Printf("Getting %d types...\n", len(types))
		if len(types) == 0 {
			return false, nil
		}
		/*for _, typ := range types {
			log.Println(typ)
		}*/
		return exists, nil
	})

	_, err = elasticsearch.PollFunction(pollingFn)
	if err != nil {
		//log.Printf("ERROR: %#v", err)
		return err
	}
	//log.Printf("SETUP INDEX: %t", ok)

	// Ingest event type
	println("Creating injest event type")
	ingestEventType := &EventType{}
	ingestEventType.Name = "piazza:ingest"
	ingestEventTypeMapping := map[string]interface{}{
		"dataId":   "string",
		"dataType": "string",
		"epsg":     "short",
		"minX":     "long",
		"minY":     "long",
		"maxX":     "long",
		"maxY":     "long",
		"hosted":   "boolean",
	}
	ingestEventType.Mapping = ingestEventTypeMapping
	//log.Println("  Creating piazza:ingest eventtype")
	postedIngestEventType := service.PostEventType(ingestEventType)
	//log.Printf("  Created piazza:ingest eventtype: %d", postedIngestEventType.StatusCode)
	if postedIngestEventType.StatusCode == 201 { // everything is ok
		service.logger.Info("  SUCCESS Created piazza:ingest eventtype: %d", postedIngestEventType.StatusCode)
	} else { // something is wrong
		service.logger.Info("  ERROR creating piazza:ingest eventtype: %d", postedIngestEventType.StatusCode)
	}

	// Execution Completed event type
	println("Creating execution completed event type")
	executionCompletedType := &EventType{}
	executionCompletedType.Name = "piazza:executionComplete"
	executionCompletedTypeMapping := map[string]interface{}{
		"jobId":  "string",
		"status": "string",
		"dataId": "string",
	}
	executionCompletedType.Mapping = executionCompletedTypeMapping
	log.Println("  Creating piazza:executionComplete eventtype")
	postedExecutionCompletedType := service.PostEventType(executionCompletedType)
	log.Printf("  Created piazza:executionComplete eventtype: %d", postedIngestEventType.StatusCode)
	if postedExecutionCompletedType.StatusCode == 201 { // everything is ok
		service.logger.Info("  SUCCESS Created piazza:executionComplete eventtype: %d", postedExecutionCompletedType.StatusCode)
	} else { // something is wrong or it was already there
		service.logger.Info("  ERROR creating piazza:excutionComplete eventtype: %d", postedExecutionCompletedType.StatusCode)
	}

	service.origin = string(sys.Name)

	return nil
}

func (service *WorkflowService) newIdent() (piazza.Ident, error) {
	uuid, err := service.uuidgen.GetUuid()
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

//---------------------------------------------------------------------

func (service *WorkflowService) statusOK(obj interface{}) *piazza.JsonResponse {
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: obj}
	err := resp.SetType()
	if err != nil {
		return service.statusInternalError(err)
	}
	return resp
}

func (service *WorkflowService) statusCreated(obj interface{}) *piazza.JsonResponse {
	resp := &piazza.JsonResponse{StatusCode: http.StatusCreated, Data: obj}
	err := resp.SetType()
	if err != nil {
		return service.statusInternalError(err)
	}
	return resp
}

func (service *WorkflowService) statusBadRequest(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusBadRequest,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

func (service *WorkflowService) statusInternalError(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusInternalServerError,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

func (service *WorkflowService) statusNotFound(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusNotFound,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

//------------------------------------------------------------------------------

// GetAdminStats TODO
func (service *WorkflowService) GetAdminStats() *piazza.JsonResponse {
	service.stats.Lock()
	t := service.stats
	service.stats.Unlock()
	return service.statusOK(t)
}

//------------------------------------------------------------------------------

// GetEventType TODO
func (service *WorkflowService) GetEventType(id piazza.Ident) *piazza.JsonResponse {

	event, found, err := service.eventTypeDB.GetOne(piazza.Ident(id))
	if !found {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}
	return service.statusOK(event)
}

// GetAllEventTypes TODO
func (service *WorkflowService) GetAllEventTypes(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	var totalHits int64
	var eventtypes []EventType
	nameParam, err := params.GetAsString("name", nil)
	if err != nil {
		return service.statusBadRequest(err)
	}
	if nameParam != nil {
		nameParamValue := *nameParam
		eventtypeid, found, err := service.eventTypeDB.GetIDByName(nameParamValue)
		if !found || eventtypeid == nil {
			return service.statusNotFound(err)
		}
		if err != nil {
			return service.statusBadRequest(err)
		}
		eventtype, found, err := service.eventTypeDB.GetOne(piazza.Ident(eventtypeid.String()))
		if !found {
			return service.statusNotFound(err)
		}
		if err != nil {
			return service.statusBadRequest(err)
		}
		eventtypes = make([]EventType, 0)
		eventtypes = append(eventtypes, *eventtype)
		totalHits = 1
	} else {
		eventtypes, totalHits, err = service.eventTypeDB.GetAll(format)
		if err != nil {
			return service.statusBadRequest(err)
		}
	}
	if eventtypes == nil {
		return service.statusInternalError(errors.New("getalleventtypes returned nil"))
	}

	resp := service.statusOK(eventtypes)

	if totalHits > 0 {
		format.Count = int(totalHits)
		resp.Pagination = format
	}

	return resp
}

// PostEventType TODO
func (service *WorkflowService) PostEventType(eventType *EventType) *piazza.JsonResponse {
	// Check if our EventType.Name already exists
	name := eventType.Name
	found, err := service.eventDB.NameExists(name)
	if err != nil {
		return service.statusInternalError(err)
	}
	if found {
		return service.statusBadRequest(LoggedError("EventType Name already exists"))
	}
	id1, found, err := service.eventTypeDB.GetIDByName(name)
	if err != nil {
		return service.statusInternalError(err)
	}
	if found {
		return service.statusBadRequest(
			LoggedError("EventType Name already exists under EventTypeId %s", id1))
	}

	eventTypeID, err := service.newIdent()
	if err != nil {
		return service.statusBadRequest(err)
	}
	eventType.EventTypeId = eventTypeID

	eventType.CreatedOn = time.Now()

	id, err := service.eventTypeDB.PostData(eventType, eventTypeID)
	if err != nil {
		return service.statusBadRequest(err)
	}

	mapping := eventType.Mapping

	err = service.eventDB.AddMapping(name, mapping)
	if err != nil {
		service.eventTypeDB.DeleteByID(id)
		return service.statusBadRequest(err)
	}

	service.logger.Info("Posted EventType with EventTypeId %s", eventTypeID)

	service.stats.IncrEventTypes()

	return service.statusCreated(eventType)
}

// DeleteEventType TODO
func (service *WorkflowService) DeleteEventType(id piazza.Ident) *piazza.JsonResponse {
	ok, err := service.eventTypeDB.DeleteByID(piazza.Ident(id))
	if !ok {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.logger.Info("Deleted EventType with EventTypeId %s", id)

	return service.statusOK(nil)
}

//------------------------------------------------------------------------------

// GetEvent TODO
func (service *WorkflowService) GetEvent(id piazza.Ident) *piazza.JsonResponse {
	mapping, err := service.eventDB.lookupEventTypeNameByEventID(id)
	if err != nil {
		return service.statusNotFound(err)
	}

	event, found, err := service.eventDB.GetOne(mapping, id)
	if !found {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	return service.statusOK(event)
}

// GetAllEvents TODO
func (service *WorkflowService) GetAllEvents(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	// if both specified, "by id"" wins
	eventTypeID, err := params.GetAsString("eventTypeId", nil)
	if err != nil {
		return service.statusBadRequest(err)
	}
	eventTypeName, err := params.GetAsString("eventTypeName", nil)
	if err != nil {
		return service.statusBadRequest(err)
	}

	query := ""

	// Get the eventTypeName corresponding to the eventTypeId
	if eventTypeID != nil {
		eventType, found, err := service.eventTypeDB.GetOne(piazza.Ident(*eventTypeID))
		if !found {
			return service.statusNotFound(err)
		}
		if err != nil {
			return service.statusBadRequest(err)
		}
		query = eventType.Name
	} else if eventTypeName != nil {
		query = *eventTypeName
	}

	events, totalHits, err := service.eventDB.GetAll(query, format)
	if err != nil {
		return service.statusBadRequest(err)
	}

	resp := service.statusOK(events)

	if totalHits > 0 {
		format.Count = int(totalHits)
		resp.Pagination = format
	}

	return resp
}

// PostRepeatingEvent deals with events that have a "CronSchedule" field specified.
// This field is checked for validity, and then set up to repeat at the interval
// specified by the CronSchedule.
// The createdBy field of each subsequent event is filled with the eventId of
// this initial event, so that searching for events created by the initial event
// is easier.
func (service *WorkflowService) PostRepeatingEvent(event *Event) *piazza.JsonResponse {
	log.Println("Posted Repeating Event")
	_, err := cron.Parse(event.CronSchedule)
	if err != nil {
		return service.statusBadRequest(err)
	}

	eventID, err := service.newIdent()
	if err != nil {
		return service.statusInternalError(err)
	}
	event.EventId = eventID

	event.CreatedOn = time.Now()

	service.cron.AddJob(event.CronSchedule, cronEvent{event, service})

	err = service.cronDB.PostData(event, eventID)
	if err != nil {
		return service.statusInternalError(err)
	}

	// Post the event in the database, WITHOUT "triggering"
	eventTypeID := event.EventTypeId
	eventType, found, err := service.eventTypeDB.GetOne(eventTypeID)
	if err != nil || !found {
		service.cron.Remove(eventID.String())
		return service.statusBadRequest(err)
	}
	eventTypeName := eventType.Name

	_, err = service.eventDB.PostData(eventTypeName, event, eventID)
	if err != nil {
		// If we fail, need to also remove from cronDB
		// We don't check for errors here because if we've reached this point,
		// the eventID will be in the cronDB
		service.cronDB.DeleteByID(eventID)
		service.cron.Remove(eventID.String())
		return service.statusInternalError(err)
	}

	service.stats.IncrEvents()

	return service.statusCreated(event)
}

// PostEvent TODO
func (service *WorkflowService) PostEvent(event *Event) *piazza.JsonResponse {
	eventTypeID := event.EventTypeId
	eventType, found, err := service.eventTypeDB.GetOne(eventTypeID)
	if err != nil || !found {
		return service.statusBadRequest(err)
	}
	eventTypeName := eventType.Name

	eventID, err := service.newIdent()
	if err != nil {
		return service.statusInternalError(err)
	}
	event.EventId = eventID

	event.CreatedOn = time.Now()

	_, err = service.eventDB.PostData(eventTypeName, event, eventID)
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.logger.Info("Posted Event with EventId %s", eventID)

	{
		// Find triggers associated with event
		triggerIDs, err := service.eventDB.PercolateEventData(eventTypeName, event.Data, eventID)
		if err != nil {
			return service.statusBadRequest(err)
		}

		// For each trigger,  apply the event data and submit job
		var waitGroup sync.WaitGroup

		results := make(map[piazza.Ident]*piazza.JsonResponse)

		for _, triggerID := range *triggerIDs {
			waitGroup.Add(1)
			go func(triggerID piazza.Ident) {
				defer waitGroup.Done()

				trigger, found, err := service.triggerDB.GetOne(triggerID)
				if !found {
					results[triggerID] = service.statusNotFound(err)
					return
				}
				if err != nil {
					results[triggerID] = service.statusBadRequest(err)
					return
				}
				if trigger.Enabled == false {
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
					results[triggerID] = service.statusInternalError(err)
					return
				}

				jobInstance, err := json.Marshal(job)
				if err != nil {
					results[triggerID] = service.statusInternalError(err)
					return
				}
				jobString := string(jobInstance)

				// Not very robust,  need to find a better way
				for key, value := range event.Data {
					jobString = strings.Replace(jobString, "$"+key, fmt.Sprintf("%v", value), -1)
				}

				service.logger.Info("job submission: %s\n", jobString)

				log.Printf("JOB ID: %s", jobID)
				log.Printf("JOB STRING: %s", jobString)

				err = service.sendToKafka(jobString, jobID)
				if err != nil {
					results[triggerID] = service.statusInternalError(err)
					return
				}

				service.stats.IncrTriggerJobs()

				alert := Alert{EventId: eventID, TriggerId: triggerID, JobId: jobID}
				resp := service.PostAlert(&alert)
				if resp.IsError() {
					// resp will be a statusInternalError or statusBadRequest
					results[triggerID] = resp
					return
				}

			}(triggerID)
		}

		waitGroup.Wait()

		for _, v := range results {
			if v != nil {
				return v
			}
		}
	}

	service.stats.IncrEvents()

	return service.statusCreated(event)
}

func (service *WorkflowService) DeleteEvent(id piazza.Ident) *piazza.JsonResponse {
	mapping, err := service.eventDB.lookupEventTypeNameByEventID(id)
	if err != nil {
		return service.statusBadRequest(err)
	}

	ok, err := service.eventDB.DeleteByID(mapping, piazza.Ident(id))
	if !ok {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	// If it's a cron event, remove from cronDB, stop cronjob
	ok, err = service.cronDB.itemExists(id)
	if err != nil {
		return service.statusBadRequest(err)
	}
	if ok {
		ok, err := service.cronDB.DeleteByID(piazza.Ident(id))
		if !ok {
			return service.statusNotFound(err)
		}
		if err != nil {
			return service.statusBadRequest(err)
		}
		service.cron.Remove(id.String())
	}

	service.logger.Info("Deleted Event with EventId %s", id)

	return service.statusOK(nil)
}

//------------------------------------------------------------------------------

func (service *WorkflowService) GetTrigger(id piazza.Ident) *piazza.JsonResponse {
	trigger, found, err := service.triggerDB.GetOne(piazza.Ident(id))
	if !found {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}
	return service.statusOK(trigger)
}

func (service *WorkflowService) GetAllTriggers(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	triggers, totalHits, err := service.triggerDB.GetAll(format)
	if err != nil {
		return service.statusBadRequest(err)
	} else if triggers == nil {
		return service.statusInternalError(errors.New("GetAllTriggers returned nil"))
	}

	resp := service.statusOK(triggers)

	if totalHits > 0 {
		format.Count = int(totalHits)
		resp.Pagination = format
	}

	return resp
}

func (service *WorkflowService) PostTrigger(trigger *Trigger) *piazza.JsonResponse {
	triggerID, err := service.newIdent()
	if err != nil {
		return service.statusBadRequest(err)
	}
	trigger.TriggerId = triggerID
	trigger.CreatedOn = time.Now()

	_, err = service.triggerDB.PostTrigger(trigger, triggerID)
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.logger.Info("Posted Trigger with TriggerId %s", triggerID)

	service.stats.IncrTriggers()

	return service.statusCreated(trigger)
}

func (service *WorkflowService) DeleteTrigger(id piazza.Ident) *piazza.JsonResponse {
	ok, err := service.triggerDB.DeleteTrigger(piazza.Ident(id))
	if !ok {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.logger.Info("Deleted Trigger with TriggerId %s", id)

	return service.statusOK(nil)
}

//---------------------------------------------------------------------

func (service *WorkflowService) GetAlert(id piazza.Ident) *piazza.JsonResponse {
	alert, found, err := service.alertDB.GetOne(id)
	if !found {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	return service.statusOK(alert)
}

func (service *WorkflowService) GetAllAlerts(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	triggerID, err := params.GetAsString("triggerId", nil)
	if err != nil {
		return service.statusBadRequest(err)
	}

	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	var alerts []Alert
	var totalHits int64

	if triggerID != nil && isUUID(*triggerID) {
		alerts, totalHits, err = service.alertDB.GetAllByTrigger(format, *triggerID)
		if err != nil {
			return service.statusBadRequest(err)
		} else if alerts == nil {
			return service.statusInternalError(errors.New("GetAllAlerts returned nil"))
		}
	} else if triggerID == nil {
		alerts, totalHits, err = service.alertDB.GetAll(format)
		if err != nil {
			return service.statusBadRequest(err)
		} else if alerts == nil {
			return service.statusInternalError(errors.New("GetAllAlerts returned nil"))
		}
	} else {
		return service.statusBadRequest(errors.New("Malformed triggerId query parameter"))
	}

	resp := service.statusOK(alerts)

	if totalHits > 0 {
		format.Count = int(totalHits)
		resp.Pagination = format
	}

	return resp
}

// PostAlert TODO
func (service *WorkflowService) PostAlert(alert *Alert) *piazza.JsonResponse {
	alertID, err := service.newIdent()
	if err != nil {
		return service.statusBadRequest(err)
	}
	alert.AlertId = alertID

	alert.CreatedOn = time.Now()

	_, err = service.alertDB.PostData(&alert, alertID)
	if err != nil {
		return service.statusInternalError(err)
	}

	service.logger.Info("Posted Alert with AlertId %s", alertID)

	service.stats.IncrAlerts()

	return service.statusCreated(alert)
}

// DeleteAlert TODO
func (service *WorkflowService) DeleteAlert(id piazza.Ident) *piazza.JsonResponse {
	ok, err := service.alertDB.DeleteByID(id)
	if !ok {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.logger.Info("Deleted Alert with AlertId %s", id)

	return service.statusOK(nil)
}

//---------------------------------------------------------------------

// InitCron TODO
func (service *WorkflowService) InitCron() error {
	ok, err := service.cronDB.Exists()
	if err != nil {
		return err
	}
	if ok {
		events, err := service.cronDB.GetAll()
		if err != nil {
			return LoggedError("WorkflowService.InitCron: Unable to get all from CronDB")
		}

		for _, e := range *events {
			service.cron.AddJob(e.CronSchedule, cronEvent{&e, service})
			if err != nil {
				return LoggedError("WorkflowService.InitCron: Unable to register cron event %#v", e)
			}
		}
	}

	service.cron.Start()

	return nil
}

type cronEvent struct {
	*Event
	service *WorkflowService
}

func (c cronEvent) Run() {
	ev := &Event{
		EventTypeId: c.EventTypeId,
		Data:        c.Data,
		CreatedOn:   time.Now(),
		CreatedBy:   c.EventId.String(),
	}
	c.service.PostEvent(ev)
}

func (c cronEvent) Key() string {
	return c.EventId.String()
}
