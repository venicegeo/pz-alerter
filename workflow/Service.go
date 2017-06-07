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
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	pzsyslog "github.com/venicegeo/pz-gocommon/syslog"
	cron "github.com/venicegeo/vegertar-cron"

	"time"
)

//------------------------------------------------------------------------------

const ingestTypeName = "piazza:ingest"
const executeTypeName = "piazza:executionComplete"

const keyEventTypes = "eventtypes"
const keyEvents = "events"
const keyTriggers = "triggers"
const keyAlerts = "alerts"
const keyCrons = "crons"
const keyTestElasticsearch = "testElasticsearch"

type Service struct {
	eventTypeDB         *EventTypeDB
	eventDB             *EventDB
	triggerDB           *TriggerDB
	alertDB             *AlertDB
	cronDB              *CronDB
	testElasticsearchDB *TestElasticsearchDB

	stats Stats
	sync.Mutex

	syslogger *pzsyslog.Logger

	sys *piazza.SystemConfig

	cron *cron.Cron

	origin string
}

//------------------------------------------------------------------------------

// Init TODO
func (service *Service) Init(
	sys *piazza.SystemConfig,
	logWriter pzsyslog.Writer,
	auditWriter pzsyslog.Writer,
	indices *map[string]elasticsearch.IIndex,
	pen string,
) error {

	eventtypesIndex := (*indices)[keyEventTypes]
	eventsIndex := (*indices)[keyEvents]
	triggersIndex := (*indices)[keyTriggers]
	alertsIndex := (*indices)[keyAlerts]
	cronIndex := (*indices)[keyCrons]
	testElasticsearchIndex := (*indices)[keyTestElasticsearch]

	var err error

	service.syslogger = pzsyslog.NewLogger(logWriter, auditWriter, string(piazza.PzWorkflow), pen)
	defer service.handlePanic()

	service.sys = sys

	service.stats.CreatedOn = piazza.NewTimeStamp()

	if service.eventTypeDB, err = NewEventTypeDB(service, eventtypesIndex); err != nil {
		return err
	}

	if service.eventDB, err = NewEventDB(service, eventsIndex); err != nil {
		return err
	}

	if service.triggerDB, err = NewTriggerDB(service, triggersIndex); err != nil {
		return err
	}

	if service.alertDB, err = NewAlertDB(service, alertsIndex); err != nil {
		return err
	}

	if service.cronDB, err = NewCronDB(service, cronIndex); err != nil {
		return err
	}

	if service.testElasticsearchDB, err = NewTestElasticsearchDB(service, testElasticsearchIndex); err != nil {
		return err
	}

	service.cron = cron.New()
	service.origin = string(sys.Name)

	// allow the database time to settle
	//time.Sleep(time.Second * 5)
	pollingFn := elasticsearch.GetData(func() (bool, error) {
		var exists bool
		exists, err = eventtypesIndex.IndexExists()
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
		var types []string
		types, err = eventtypesIndex.GetTypes()
		if err != nil {
			return false, err
		}
		//log.Printf("Getting %d types...", len(types))
		if len(types) == 0 {
			return false, nil
		}
		/*for _, typ := range types {
			log.Printf(typ)
		}*/
		return exists, nil
	})

	if _, err = elasticsearch.PollFunction(pollingFn); err != nil {
		//log.Printf("ERROR: %#v", err)
		return err
	}
	//log.Printf("SETUP INDEX: %t", ok)

	// Ingest event type
	ingestEventType := &EventType{Name: ingestTypeName, CreatedBy: sys.PiazzaSystem}
	ingestEventTypeMapping := map[string]interface{}{
		"dataId":   "string",
		"dataType": "string",
		"epsg":     "integer",
		"minX":     "double",
		"minY":     "double",
		"maxX":     "double",
		"maxY":     "double",
		"hosted":   "boolean",
	}
	ingestEventType.Mapping = ingestEventTypeMapping
	//log.Println("  Creating piazza:ingest eventtype")
	postedIngestEventType := service.PostEventType(ingestEventType)
	//log.Printf("  Created piazza:ingest eventtype: %d", postedIngestEventType.StatusCode)
	if postedIngestEventType.StatusCode == 201 { // everything is ok
		service.syslogger.Info("  SUCCESS Created piazza:ingest eventtype: %s", postedIngestEventType.StatusCode)
	} else { // something is wrong
		service.syslogger.Info("  ERROR creating piazza:ingest eventtype: %s", postedIngestEventType.StatusCode)
	}

	// Execution Completed event type
	executionCompletedType := &EventType{Name: executeTypeName, CreatedBy: sys.PiazzaSystem}
	executionCompletedTypeMapping := map[string]interface{}{
		"jobId":  "string",
		"status": "string",
		"dataId": "string",
	}
	executionCompletedType.Mapping = executionCompletedTypeMapping
	postedExecutionCompletedType := service.PostEventType(executionCompletedType)
	if postedExecutionCompletedType.StatusCode == 201 { // everything is ok
		service.syslogger.Info("  SUCCESS Created piazza:executionComplete eventtype: %s", postedExecutionCompletedType.StatusCode)
	} else { // something is wrong or it was already there
		service.syslogger.Info("  ERROR creating piazza:excutionComplete eventtype: %s", postedExecutionCompletedType.StatusCode)
	}

	return nil
}

func (service *Service) newIdent() piazza.Ident {
	return piazza.Ident(piazza.NewUuid().String())
}

func (service *Service) handlePanic() {
	if r := recover(); r != nil {
		report := fmt.Sprintf("Recovered from panic: [%s]: %v\n%s", reflect.TypeOf(r), r, string(debug.Stack()))
		service.syslogger.Error(report)
		fmt.Fprintln(os.Stderr, report)
	}
}

func (service *Service) sendToKafka(jobInstance string, jobID piazza.Ident, actor string) error {
	service.syslogger.Audit(actor, "creatingJob", "kafka", "User [%s] is sending job [%s] to kafka", actor, jobID)
	kafkaAddress, err := service.sys.GetAddress(piazza.PzKafka)
	if err != nil {
		service.syslogger.Audit(actor, "creatingJobFailure", "kafka", "User [%s] sending job [%s] to kafka failed (1)", actor, jobID)
		return LoggedError("Kafka-related failure (1): %s", err.Error())
	}

	topic := fmt.Sprintf("Request-Job-%s", service.sys.Space)
	message := jobInstance

	producer, err := sarama.NewSyncProducer(strings.Split(kafkaAddress, ","), nil)
	if err != nil {
		service.syslogger.Audit(actor, "creatingJobFailure", "kafka", "User [%s] sending job [%s] to kafka failed (2)", actor, jobID)
		return LoggedError("Kafka-related failure (2): %s", err.Error())
	}
	defer func() {
		if errC := producer.Close(); errC != nil {
			service.syslogger.Audit(actor, "creatingJobFailure", "kafka", "User [%s] sending job [%s] to kafka failed (3)", actor, jobID)
			log.Fatalf("Kafka-related failure (3): " + errC.Error())
		}
	}()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
		Key:   sarama.StringEncoder(jobID)}
	if _, _, err = producer.SendMessage(msg); err != nil {
		service.syslogger.Audit(actor, "creatingJobFailure", "kafka", "User [%s] sending job [%s] to kafka (4)", actor, jobID)
		return LoggedError("Kafka-related failure (4): %s", err.Error())
	}
	service.syslogger.Audit(actor, "createdJob", "kafka", "User [%s] sent job [%s] to kafka", actor, jobID)

	return nil
}

//---------------------------------------------------------------------

func (service *Service) statusOK(obj interface{}) *piazza.JsonResponse {
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: obj}
	if err := resp.SetType(); err != nil {
		return service.statusInternalError(err)
	}
	return resp
}

func (service *Service) statusPutOK(message string) *piazza.JsonResponse {
	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Type:       "success",
		Message:    message,
		Origin:     service.origin,
	}
	return resp
}

func (service *Service) statusCreated(obj interface{}) *piazza.JsonResponse {
	resp := &piazza.JsonResponse{StatusCode: http.StatusCreated, Data: obj}
	if err := resp.SetType(); err != nil {
		return service.statusInternalError(err)
	}
	return resp
}

func (service *Service) statusBadRequest(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusBadRequest,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

func (service *Service) statusForbidden(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusForbidden,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

func (service *Service) statusInternalError(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusInternalServerError,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

func (service *Service) statusNotFound(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusNotFound,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

//------------------------------------------------------------------------------

// GetStats TODO
func (service *Service) GetStats() *piazza.JsonResponse {
	defer service.handlePanic()
	service.Lock()
	t := service.stats
	service.Unlock()
	return service.statusOK(t)
}

//------------------------------------------------------------------------------

// GetEventType TODO
func (service *Service) GetEventType(id piazza.Ident, actor string) *piazza.JsonResponse {
	defer service.handlePanic()
	service.syslogger.Audit(actor, "gettingEventType", id, "Service.GetEventType: User is getting eventType")
	eventType, found, err := service.eventTypeDB.GetOne(id, actor)
	if !found {
		service.syslogger.Audit(actor, "gettingEventTypeFailure", id, "Service.GetEventType: User [%s] failed to get eventType [%s]", actor, id)
		return service.statusNotFound(err)
	}
	if err != nil {
		service.syslogger.Audit(actor, "gettingEventTypeFailure", id, "Service.GetEventType: User [%s] failed to get eventType [%s]", actor, id)
		return service.statusBadRequest(err)
	}
	service.syslogger.Audit(actor, "gotEventType", id, "Service.GetEventType: User [%s] successfully got eventType [%s]", actor, id)

	eventType.Mapping = service.removeUniqueParams(eventType.Name, eventType.Mapping)
	return service.statusOK(eventType)
}

// GetAllEventTypes TODO
func (service *Service) GetAllEventTypes(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	defer service.handlePanic()
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	var totalHits int64
	var eventtypes []EventType
	nameParam, err := params.GetAsString("name", "")
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit("pz-workflow", "gettingAllEventTypes", service.eventTypeDB.mapping, "Service.GetAllEventTypes: User is getting all eventTypes")

	if nameParam != "" {
		nameParamValue := nameParam
		var foundName bool
		var eventtypeid *piazza.Ident
		eventtypeid, foundName, err = service.eventTypeDB.GetIDByName(format, nameParamValue, "pz-workflow")
		var foundType = false
		var eventtype *EventType
		if foundName && eventtypeid != nil {
			if err != nil {
				service.syslogger.Audit("pz-workflow", "gettingAllEventTypesFailure", service.eventTypeDB.mapping, "Service.GetAllEventTypes: User failed to get all eventTypes")
				return service.statusBadRequest(err)
			}
			eventtype, foundType, err = service.eventTypeDB.GetOne(*eventtypeid, "pz-workflow")
			if err != nil {
				service.syslogger.Audit("pz-workflow", "gettingAllEventTypesFailure", service.eventTypeDB.mapping, "Service.GetAllEventTypes: User failed to get all eventTypes")
				return service.statusInternalError(err)
			}
		}
		eventtypes = make([]EventType, 0)
		if foundType && eventtype != nil {
			eventtypes = append(eventtypes, *eventtype)
		}
		totalHits = int64(len(eventtypes))
	} else {
		eventtypes, totalHits, err = service.eventTypeDB.GetAll(format, "pz-workflow")
		if err != nil {
			service.syslogger.Audit("pz-workflow", "gettingAllEventTypesFailure", service.eventTypeDB.mapping, "Service.GetAllEventTypes: User failed to get all eventTypes")
			return service.statusInternalError(err)
		}
	}
	if eventtypes == nil {
		service.syslogger.Audit("pz-workflow", "gettingAllEventTypesFailure", service.eventTypeDB.mapping, "Service.GetAllEventTypes: User failed to get all eventTypes")
		return service.statusInternalError(errors.New("getalleventtypes returned nil"))
	}
	for i := 0; i < len(eventtypes); i++ {
		eventtypes[i].Mapping = service.removeUniqueParams(eventtypes[i].Name, eventtypes[i].Mapping)
	}
	resp := service.statusOK(eventtypes)

	service.syslogger.Audit("pz-workflow", "gotAllEventTypes", service.eventTypeDB.mapping, "Service.GetAllEventTypes: User successfully got all eventTypes")

	format.Count = int(totalHits)
	resp.Pagination = format

	return resp
}

func (service *Service) QueryEventTypes(dslString string, params *piazza.HttpQueryParams) *piazza.JsonResponse {
	defer service.handlePanic()
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	var totalHits int64
	var eventtypes []EventType
	if dslString, err = syncPagination(dslString, *format); err != nil {
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit("pz-workflow", "queryingEventTypes", service.eventTypeDB.mapping, "Service.QueryEventTypes: User is querying eventTypes")

	eventtypes, totalHits, err = service.eventTypeDB.GetEventTypesByDslQuery(dslString, "pz-workflow")
	if err != nil {
		service.syslogger.Audit("pz-workflow", "queryingEventTypesFailure", service.eventTypeDB.mapping, "Service.QueryEventTypes: User failed to query eventTypes")
		return service.statusBadRequest(err)
	}
	if eventtypes == nil {
		service.syslogger.Audit("pz-workflow", "queryingEventTypesFailure", service.eventTypeDB.mapping, "Service.QueryEventTypes: User failed to query eventTypes")
		return service.statusInternalError(errors.New("queryeventtypes returned nil"))
	}
	for i := 0; i < len(eventtypes); i++ {
		eventtypes[i].Mapping = service.removeUniqueParams(eventtypes[i].Name, eventtypes[i].Mapping)
	}
	resp := service.statusOK(eventtypes)

	service.syslogger.Audit("pz-workflow", "queriedEventTypes", service.eventTypeDB.mapping, "Service.QueryEventTypes: User successfully queried eventTypes")

	format.Count = int(totalHits)
	resp.Pagination = format

	return resp
}

// PostEventType TODO
func (service *Service) PostEventType(eventType *EventType) *piazza.JsonResponse {
	defer service.handlePanic()
	// Check if our EventType.Name already exists
	found, err := service.eventDB.NameExists(eventType.Name, eventType.CreatedBy)
	if err != nil {
		return service.statusInternalError(err)
	}
	if found {
		return service.statusBadRequest(LoggedError("EventType Name already exists"))
	}
	id1, found, err := service.eventTypeDB.GetIDByName(nil, eventType.Name, eventType.CreatedBy)
	if err != nil {
		return service.statusInternalError(err)
	}
	if found {
		return service.statusBadRequest(
			LoggedError("EventType Name already exists under EventTypeId %s", id1))
	}

	eventType.EventTypeID = service.newIdent()
	eventType.CreatedOn = piazza.NewTimeStamp()

	vars, err := piazza.GetVarsFromStruct(eventType.Mapping)
	if err != nil {
		return service.statusBadRequest(LoggedError("EventTypeDB.PostData failed: %s", err))
	}
	for k := range vars {
		if strings.Contains(k, "~") {
			return service.statusBadRequest(LoggedError("EventTypeDB.PostData failed: Variable names cannot contain '%s~': [%s]", eventType.Name, k))
		}
	}

	response := *eventType

	eventType.Mapping = service.addUniqueParams(eventType.Name, eventType.Mapping)

	service.syslogger.Audit(eventType.CreatedBy, "creatingEventType", eventType.EventTypeID, "Service.PostEventType: User [%s] is creating eventType [%s]", eventType.CreatedBy, eventType.EventTypeID)

	if err = service.eventTypeDB.PostData(eventType); err != nil {
		service.syslogger.Audit(eventType.CreatedBy, "creatingEventTypeFailure", eventType.EventTypeID, "Service.PostEventType: User [%s] failed to create eventType [%s]", eventType.CreatedBy, eventType.EventTypeID)
		if strings.HasSuffix(err.Error(), "was not recognized as a valid mapping type") {
			return service.statusBadRequest(err)
		}
		return service.statusInternalError(err)
	}

	if err = service.eventDB.AddMapping(eventType.Name, eventType.Mapping, eventType.CreatedBy); err != nil {
		service.syslogger.Audit(eventType.CreatedBy, "creatingEventTypeFailure", eventType.EventTypeID, "Service.PostEventType: User [%s] failed to create eventType [%s]", eventType.CreatedBy, eventType.EventTypeID)
		_, _ = service.eventTypeDB.DeleteByID(eventType.EventTypeID, eventType.CreatedBy)
		return service.statusInternalError(err)
	}

	service.syslogger.Audit(eventType.CreatedBy, "createdEventType", eventType.EventTypeID, "Service.PostEventType: User [%s] successfully created eventType [%s]", eventType.CreatedBy, eventType.EventTypeID)

	service.stats.IncrEventTypes()

	return service.statusCreated(&response)
}

// IsSystemEvent returns true if the event was generated within Piazza.
//
// TODO: Instead, check if createdBy=system
func IsSystemEvent(name string) bool {
	return name == ingestTypeName || name == executeTypeName
}

// DeleteEventType TODO
func (service *Service) DeleteEventType(id piazza.Ident) *piazza.JsonResponse {
	defer service.handlePanic()
	eventType, found, err := service.eventTypeDB.GetOne(id, "pz-workflow")
	if !found {
		service.syslogger.Audit("pz-workflow", "deletingEventTypeFailure", id, "Service.DeleteEventType: failed to get eventType [%s]", id)
		return service.statusNotFound(err)
	}
	if err != nil {
		service.syslogger.Audit("pz-workflow", "deletingEventTypeFailure", id, "Service.DeleteEventType: failed to get eventType [%s]", id)
		return service.statusBadRequest(err)
	}
	// Only check for system events or "in use" if found
	if found {
		if eventType != nil && IsSystemEvent(eventType.Name) {
			return service.statusBadRequest(errors.New("Deleting system eventTypes is prohibited"))
		}

		var triggers []Trigger
		var hits int64
		if triggers, hits, err = service.triggerDB.GetTriggersByEventTypeID(nil, id, "pz-workflow"); err != nil {
			return service.statusBadRequest(err)
		}
		if hits > 0 || len(triggers) > 0 {
			return service.statusForbidden(errors.New("Deleting eventTypes that are in use is prohibited"))
		}

		var events []Event
		if events, hits, err = service.eventDB.GetEventsByEventTypeID(nil, eventType.Name, id, "pz-workflow"); err != nil {
			return service.statusBadRequest(err)
		}
		if hits > 0 || len(events) > 0 {
			return service.statusForbidden(errors.New("Deleting eventTypes that are in use is prohibited"))
		}
	}

	service.syslogger.Audit("pz-workflow", "deletingEventType", id, "Service.DeleteEventType: User is deleting eventType [%s]", id)

	ok, err := service.eventTypeDB.DeleteByID(id, "pz-workflow")
	if !ok {
		service.syslogger.Audit("pz-workflow", "deletingEventTypeFailure", id, "Service.DeleteEventType: User failed to delete eventType [%s]", id)
		return service.statusNotFound(err)
	}
	if err != nil {
		service.syslogger.Audit("pz-workflow", "deletingEventTypeFailure", id, "Service.DeleteEventType: User failed to delete eventType [%s]", id)
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit("pz-workflow", "deletedEventType", id, "Service.DeleteEventType: User successfully deleted eventType [%s]", id)

	return service.statusOK(nil)
}

//------------------------------------------------------------------------------

// GetEvent TODO
func (service *Service) GetEvent(id piazza.Ident) *piazza.JsonResponse {
	defer service.handlePanic()
	mapping, err := service.eventDB.lookupEventTypeNameByEventID(id, "pz-workflow")
	if mapping == "" {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit("pz-workflow", "gettingEvent", id, "Service.GetEvent: User is getting event [%s]", id)
	event, found, err := service.eventDB.GetOne(mapping, id, "pz-workflow")
	if !found {
		service.syslogger.Audit("pz-workflow", "gettingEventFailure", id, "Service.GetEvent: User failed to get event [%s]", id)
		return service.statusNotFound(err)
	}
	if err != nil {
		service.syslogger.Audit("pz-workflow", "gettingEventFailure", id, "Service.GetEvent: User failed to get event [%s]", id)
		return service.statusBadRequest(err)
	}
	service.syslogger.Audit("pz-workflow", "gotEvent", id, "Service.GetEvent: User successfully got event [%s]", id)

	event.Data = service.removeUniqueParams(mapping, event.Data)
	return service.statusOK(event)
}

// GetAllEvents TODO
func (service *Service) GetAllEvents(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	defer service.handlePanic()
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	// if both specified, "by id"" wins
	eventTypeID, err := params.GetAsString("eventTypeId", "")
	if err != nil {
		return service.statusBadRequest(err)
	}
	eventTypeName, err := params.GetAsString("eventTypeName", "")
	if err != nil {
		return service.statusBadRequest(err)
	}

	var query string

	// Get the eventTypeName corresponding to the eventTypeId
	if eventTypeID != "" {
		var eventType *EventType
		var found bool
		eventType, found, err = service.eventTypeDB.GetOne(piazza.Ident(eventTypeID), "pz-workflow")
		if !found {
			return service.statusNotFound(err)
		}
		if err != nil {
			return service.statusBadRequest(err)
		}
		query = eventType.Name
	} else if eventTypeName != "" {
		query = eventTypeName
	} else {
		// no query param specified, get 'em all
		query = ""
	}

	service.syslogger.Audit("pz-workflow", "gettingAllEvents", service.eventDB.Esi.IndexName(), "Service.GetAllEvents: User is getting all events")

	events, totalHits, err := service.eventDB.GetAll(query, format, "pz-workflow")
	if err != nil {
		service.syslogger.Audit("pz-workflow", "gettingAllEventsFailure", service.eventDB.Esi.IndexName(), "Service.GetAllEvents: User failed to get all events")
		return service.statusInternalError(err)
	}
	for i := 0; i < len(events); i++ {
		eventType, found, err := service.eventTypeDB.GetOne(events[i].EventTypeID, "pz-workflow")
		if !found || err != nil {
			service.syslogger.Audit("pz-workflow", "gettingAllEventsFailure", service.eventDB.Esi.IndexName(), "Service.GetAllEvents: User failed to get all events")
			return service.statusInternalError(err)
		}
		events[i].Data = service.removeUniqueParams(eventType.Name, events[i].Data)
	}
	resp := service.statusOK(events)

	service.syslogger.Audit("pz-workflow", "gotAllEvents", service.eventDB.Esi.IndexName(), "Service.GetAllEvents: User successfully got all events")

	format.Count = int(totalHits)
	resp.Pagination = format

	return resp
}

// PostRepeatingEvent deals with events that have a "CronSchedule" field specified.
// This field is checked for validity, and then set up to repeat at the interval
// specified by the CronSchedule.
// The createdBy field of each subsequent event is filled with the eventId of
// this initial event, so that searching for events created by the initial event
// is easier.
func (service *Service) PostRepeatingEvent(event *Event) *piazza.JsonResponse {
	defer service.handlePanic()
	// Post the event in the database, WITHOUT "triggering"
	eventTypeID := event.EventTypeID
	eventType, found, err := service.eventTypeDB.GetOne(eventTypeID, event.CreatedBy)
	if err != nil || !found {
		return service.statusBadRequest(err)
	}

	//log.Println("Posted Repeating Event")
	if _, err = cron.Parse(event.CronSchedule); err != nil {
		return service.statusBadRequest(err)
	}

	event.EventID = service.newIdent()
	event.CreatedOn = piazza.NewTimeStamp()

	response := *event

	event.Data = service.addUniqueParams(eventType.Name, event.Data)

	service.syslogger.Audit(event.CreatedBy, "creatingCronEvent", event.EventID, "Service.PostRepeatingEvent: User [%s] is creating cron event [%s]", event.CreatedBy, event.EventID)

	if err = service.cron.AddJob(event.CronSchedule, cronEvent{event, eventType.Name, service}); err != nil {
		service.syslogger.Audit(event.CreatedBy, "creatingCronEventFailure", event.EventID, "Service.PostRepeatingEvent: User [%s] failed to create cron event [%s]", event.CreatedBy, event.EventID)
		return service.statusInternalError(err)
	}

	if err = service.cronDB.PostData(event); err != nil {
		service.syslogger.Audit(event.CreatedBy, "creatingCronEventFailure", event.EventID, "Service.PostRepeatingEvent: User [%s] failed to create cron event [%s]", event.CreatedBy, event.EventID)
		return service.statusInternalError(err)
	}

	if err = service.eventDB.PostData(event, eventType.Name); err != nil {
		service.syslogger.Audit(event.CreatedBy, "creatingCronEventFailure", event.EventID, "Service.PostRepeatingEvent: User [%s] failed to create cron event [%s]", event.CreatedBy, event.EventID)
		// If we fail, need to also remove from cronDB
		// We don't check for errors here because if we've reached this point,
		// the eventID will be in the cronDB
		_, _ = service.cronDB.DeleteByID(event.EventID, event.CreatedBy)
		service.cron.Remove(event.EventID.String())
		return service.statusInternalError(err)
	}

	service.syslogger.Audit(event.CreatedBy, "createdCronEvent", event.EventID, "Service.PostRepeatingEvent: User [%s] successfully created cron event [%s] on schedule [%s]", event.CreatedBy, event.EventID, event.CronSchedule)

	service.stats.IncrEvents()

	return service.statusCreated(&response)
}

// PostEvent TODO
func (service *Service) PostEvent(event *Event) *piazza.JsonResponse {
	defer service.handlePanic()
	eventType, found, err := service.eventTypeDB.GetOne(event.EventTypeID, event.CreatedBy)
	if err != nil || !found {
		return service.statusBadRequest(err)
	}

	event.EventID = service.newIdent()
	event.CreatedOn = piazza.NewTimeStamp()

	response := *event

	event.Data = service.addUniqueParams(eventType.Name, event.Data)

	service.syslogger.Audit(event.CreatedBy, "creatingEvent", event.EventID, "Service.PostEvent: User [%s] is creating event [%s]", event.CreatedBy, event.EventID)

	if err = service.eventDB.PostData(event, eventType.Name); err != nil {
		service.syslogger.Audit(event.CreatedBy, "creatingEventFailure", event.EventID, "Service.PostEvent: User [%s] failed to create event [%s]", event.CreatedBy, event.EventID)
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit(event.CreatedBy, "createdEvent", event.EventID, "Service.PostEvent: User [%s] successfully created event [%s]", event.CreatedBy, event.EventID)

	{
		// Find triggers associated with event
		triggerIDs, err1 := service.eventDB.PercolateEventData(eventType.Name, event.Data, event.EventID, event.CreatedBy)
		if err1 != nil {
			return service.statusBadRequest(err1)
		}

		// For each trigger,  apply the event data and submit job
		var waitGroup sync.WaitGroup

		results := make(map[piazza.Ident]*piazza.JsonResponse)

		for _, triggerID := range *triggerIDs {
			waitGroup.Add(1)
			go func(triggerID piazza.Ident) {
				defer waitGroup.Done()

				trigger, found, err2 := service.triggerDB.GetOne(triggerID, event.CreatedBy)
				if err2 != nil {
					results[triggerID] = service.statusBadRequest(err2)
					return
				}
				if !found {
					// Don't fail for this, just log something and continue to the next trigger id
					service.syslogger.Warning("Percolation error: Trigger %s does not exist", string(triggerID))
					return
				}
				if !trigger.Enabled {
					//results[triggerID] = statusOK(triggerID)
					return
				}

				// Not the best way to do this, but should disallow Triggers from firing if they
				// don't have the same Eventtype as the Event
				// Would rather have this done via the percolation itself ...
				if eventType.EventTypeID != trigger.EventTypeID {
					return
				}

				// jobID gets sent through Kafka as the key
				job := trigger.Job
				jobID := service.newIdent()

				jobInstance, err4 := json.Marshal(job)
				if err4 != nil {
					results[triggerID] = service.statusInternalError(err4)
					return
				}
				jobString := string(jobInstance)

				idamURL, err5 := service.sys.GetURL(piazza.PzIdam)
				service.syslogger.Info("Requesting pz-idam url: %s", idamURL)
				if err5 == nil { //Mocking
					service.syslogger.Audit("pz-workflow", "createJobRequestAccess", "pz-idam", "User [%s] POSTed event [%s] requesting access to trigger [%s] created by [%s]", event.CreatedBy, event.EventID, trigger.TriggerID, trigger.CreatedBy)
					auth, err6 := piazza.RequestAuthZAccess(idamURL, eventType.CreatedBy)
					service.syslogger.Info("Pz-idam authoriazation for user [%s]: %t", eventType.CreatedBy, auth)
					if err6 != nil {
						results[triggerID] = service.statusInternalError(err6)
						service.syslogger.Audit("pz-workflow", "createJobRequestAccessFailure", "pz-idam", "Event [%s] firing trigger [%s] could not get access to create job", event.EventID, trigger.TriggerID)
						return
					} else if !auth {
						results[triggerID] = service.statusForbidden(errors.New("Access to create job denied"))
						service.syslogger.Audit("pz-workflow", "createJobRequestAccessDenied", "pz-idam", "Event [%s] firing trigger [%s] was denied access to create job", event.EventID, trigger.TriggerID)
						return
					}
				}

				service.syslogger.Audit("pz-workflow", "createJobRequestAccessGranted", "pz-idam", "Event [%s] firing trigger [%s] was granted access to create job", event.EventID, trigger.TriggerID)
				service.syslogger.Info("job [%s] submission by event [%s] using trigger [%s]: %s\n", jobID, event.EventID, triggerID, jobString)

				// Not very robust,  need to find a better way
				for key, value := range event.Data[eventType.Name].(map[string]interface{}) {
					jobString = strings.Replace(jobString, "$"+key, fmt.Sprintf("%v", value), -1)
				}

				//log.Printf("JOB ID: %s", jobID)
				//log.Printf("JOB STRING: %s", jobString)

				err7 := service.sendToKafka(jobString, jobID, trigger.CreatedBy)
				if err7 != nil {
					results[triggerID] = service.statusInternalError(err7)
					return
				}

				service.stats.IncrTriggerJobs()

				alert := Alert{EventID: event.EventID, TriggerID: triggerID, JobID: jobID, CreatedBy: trigger.CreatedBy}
				if resp := service.PostAlert(&alert); resp.IsError() {
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

	return service.statusCreated(&response)
}

func (service *Service) QueryEvents(jsonString string, params *piazza.HttpQueryParams) *piazza.JsonResponse {
	defer service.handlePanic()
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	// if both specified, "by id"" wins
	eventTypeID, err := params.GetAsString("eventTypeId", "")
	if err != nil {
		return service.statusBadRequest(err)
	}
	eventTypeName, err := params.GetAsString("eventTypeName", "")
	if err != nil {
		return service.statusBadRequest(err)
	}

	var query string

	// Get the eventTypeName corresponding to the eventTypeId
	if eventTypeID != "" {
		var eventType *EventType
		var found bool
		eventType, found, err = service.eventTypeDB.GetOne(piazza.Ident(eventTypeID), "pz-workflow")
		if !found {
			return service.statusNotFound(err)
		}
		if err != nil {
			return service.statusBadRequest(err)
		}
		query = eventType.Name
	} else if eventTypeName != "" {
		query = eventTypeName
	} else {
		// no query param specified, get 'em all
		query = ""
	}

	if jsonString, err = syncPagination(jsonString, *format); err != nil {
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit("pz-workflow", "queryingEvents", service.eventDB.Esi.IndexName(), "Service.QueryEvents: User is querying events")

	events, totalHits, err := service.eventDB.GetEventsByDslQuery(query, jsonString, "pz-workflow")
	if err != nil {
		service.syslogger.Audit("pz-workflow", "queryingEventsFailure", service.eventDB.Esi.IndexName(), "Service.QueryEvents: User failed to query events")
		return service.statusBadRequest(err)
	}
	resp := service.statusOK(events)

	service.syslogger.Audit("pz-workflow", "queriedEvents", service.eventDB.Esi.IndexName(), "Service.QueryEvents: User successfully queried events")

	format.Count = int(totalHits)
	resp.Pagination = format

	return resp
}

func syncPagination(dslString string, format piazza.JsonPagination) (string, error) {
	// Overwrite any from/size in dsl with what's in the params
	b := []byte(dslString)
	var f interface{}
	err := json.Unmarshal(b, &f)
	if err != nil {
		return "", err
	}
	dsl := f.(map[string]interface{})
	dsl["from"] = format.Page * format.PerPage
	dsl["size"] = format.PerPage
	if dsl["sort"] == nil {
		// Since ES has more fine grained sorting allow their sorting to take precedence
		// If sorting wasn't specified in the DSL, put in sorting from Piazza
		bts := []byte("[{\"" + format.SortBy + "\":\"" + string(format.Order) + "\"}]")
		var g interface{}
		if err = json.Unmarshal(bts, &g); err != nil {
			return "", err
		}
		sortDsl := g.([]interface{})
		dsl["sort"] = sortDsl
	}
	byteArray, err := json.Marshal(dsl)
	if err != nil {
		return "", err
	}
	return string(byteArray), nil
}

func (service *Service) DeleteEvent(id piazza.Ident) *piazza.JsonResponse {
	defer service.handlePanic()
	mapping, err := service.eventDB.lookupEventTypeNameByEventID(id, "pz-workflow")
	if mapping == "" {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	if IsSystemEvent(mapping) {
		return service.statusBadRequest(errors.New("Deleting system events is prohibited"))
	}

	service.syslogger.Audit("pz-workflow", "deletingEvent", id, "Service.DeleteEvent: User is deleteing event [%s]", id)

	ok, err := service.eventDB.DeleteByID(mapping, id, "pz-workflow")
	if !ok {
		service.syslogger.Audit("pz-workflow", "deletingEventFailure", id, "Service.DeleteEvent: User failed to delete event [%s]", id)
		return service.statusNotFound(err)
	}
	if err != nil {
		service.syslogger.Audit("pz-workflow", "deletingEventFailure", id, "Service.DeleteEvent: User failed to delete event [%s]", id)
		return service.statusBadRequest(err)
	}
	service.syslogger.Audit("pz-workflow", "deletedCronEvent", id, "Service.DeleteEvent: User successfully deleted event [%s]", id)

	// If it's a cron event, remove from cronDB, stop cronjob
	ok, err = service.cronDB.itemExists(id, "pz-workflow")
	if err != nil {
		return service.statusBadRequest(err)
	}
	if ok {
		service.syslogger.Audit("pz-workflow", "deletingCronEvent", id, "Service.DeleteEvent: User is deleting cron event [%s]", id)
		ok, err := service.cronDB.DeleteByID(id, "pz-workflow")
		if !ok {
			service.syslogger.Audit("pz-workflow", "deletingCronEventFailure", id, "Service.DeleteEvent: User failed to delete cron event [%s]", id)
			return service.statusNotFound(err)
		}
		if err != nil {
			service.syslogger.Audit("pz-workflow", "deletingCronEventFailure", id, "Service.DeleteEvent: User failed to delete cron event [%s]", id)
			return service.statusBadRequest(err)
		}
		service.syslogger.Audit("pz-workflow", "deletedCronEvent", id, "Service.DeleteEvent: User successfully deleted cron event [%s]", id)
		service.cron.Remove(id.String())
	}

	return service.statusOK(nil)
}

//------------------------------------------------------------------------------

func (service *Service) GetTrigger(id piazza.Ident) *piazza.JsonResponse {
	defer service.handlePanic()
	service.syslogger.Audit("pz-workflow", "gettingTrigger", id, "Service.GetTrigger: User is getting trigger [%s]", id)
	trigger, found, err := service.triggerDB.GetOne(id, "pz-workflow")
	if !found {
		service.syslogger.Audit("pz-workflow", "gettingTriggerFailure", id, "Service.GetTrigger: User failed to get trigger [%s]", id)
		return service.statusNotFound(err)
	}
	if err != nil {
		service.syslogger.Audit("pz-workflow", "gettingTriggerFailure", id, "Service.GetTrigger: User failed to get trigger [%s]", id)
		return service.statusBadRequest(err)
	}
	eventType, found, err := service.eventTypeDB.GetOne(trigger.EventTypeID, "pz-workflow")
	if err != nil || !found {
		service.syslogger.Audit("pz-workflow", "gettingTriggerFailure", id, "Service.GetTrigger: User failed to get trigger [%s]", id)
		return service.statusBadRequest(err)
	}
	service.syslogger.Audit("pz-workflow", "gotTrigger", id, "Service.GetTrigger: User successfully got trigger [%s]", id)

	trigger.Condition = service.removeUniqueParams(eventType.Name, trigger.Condition)
	return service.statusOK(trigger)
}

func (service *Service) GetAllTriggers(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	defer service.handlePanic()
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit("pz-workflow", "gettingAllTriggers", service.triggerDB.mapping, "Service.GetAllTriggers: User is getting all triggers")

	triggers, totalHits, err := service.triggerDB.GetAll(format, "pz-workflow")
	if err != nil {
		service.syslogger.Audit("pz-workflow", "gettingAllTriggersFailure", service.triggerDB.mapping, "Service.GetAllTriggers: User failed to get all triggers")
		return service.statusInternalError(err)
	} else if triggers == nil {
		service.syslogger.Audit("pz-workflow", "gettingAllTriggersFailure", service.triggerDB.mapping, "Service.GetAllTriggers: User failed to get all triggers")
		return service.statusInternalError(errors.New("GetAllTriggers returned nil"))
	}
	for i := 0; i < len(triggers); i++ {
		eventType, found, err := service.eventTypeDB.GetOne(triggers[i].EventTypeID, "pz-workflow")
		if err != nil || !found {
			continue //v Old implementation
			//return service.statusBadRequest(err)
		}
		triggers[i].Condition = service.removeUniqueParams(eventType.Name, triggers[i].Condition)
	}
	resp := service.statusOK(triggers)

	service.syslogger.Audit("pz-workflow", "gotAllTriggers", service.triggerDB.mapping, "Service.GetAllTriggers: User successfully got all triggers")

	format.Count = int(totalHits)
	resp.Pagination = format

	return resp
}

func (service *Service) QueryTriggers(dslString string, params *piazza.HttpQueryParams) *piazza.JsonResponse {
	defer service.handlePanic()
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit("pz-workflow", "queryingTriggers", service.triggerDB.mapping, "Service.QueryTriggers: User is querying triggers")

	dslString, err = syncPagination(dslString, *format)
	if err != nil {
		service.syslogger.Audit("pz-workflow", "queryingTriggersFailure", service.triggerDB.mapping, "Service.QueryTriggers: syncPagination failed")
		return service.statusBadRequest(err)
	}
	triggers, totalHits, err := service.triggerDB.GetTriggersByDslQuery(dslString, "pz-workflow")
	if err != nil {
		service.syslogger.Audit("pz-workflow", "queryingTriggersFailure", service.triggerDB.mapping, "Service.QueryTriggers: User failed to query triggers")
		return service.statusBadRequest(err)
	} else if triggers == nil {
		service.syslogger.Audit("pz-workflow", "queryingTriggersFailure", service.triggerDB.mapping, "Service.QueryTriggers: User failed to query triggers")
		return service.statusInternalError(errors.New("QueryTriggers returned nil"))
	}
	for i := 0; i < len(triggers); i++ {
		eventType, found, err := service.eventTypeDB.GetOne(triggers[i].EventTypeID, "pz-workflow")
		if err != nil || !found {
			service.syslogger.Audit("pz-workflow", "queryingTriggersFailure", service.triggerDB.mapping, "Service.QueryTriggers: User failed to query triggers")
			return service.statusBadRequest(err)
		}
		triggers[i].Condition = service.removeUniqueParams(eventType.Name, triggers[i].Condition)
	}
	resp := service.statusOK(triggers)

	service.syslogger.Audit("pz-workflow", "queriedTriggers", service.triggerDB.mapping, "Service.QueryTriggers: User successfully queried triggers")

	format.Count = int(totalHits)
	resp.Pagination = format

	return resp
}

func (service *Service) PostTrigger(trigger *Trigger) *piazza.JsonResponse {
	defer service.handlePanic()
	var err error
	trigger.TriggerID = service.newIdent()
	trigger.CreatedOn = piazza.NewTimeStamp()

	var eventType *EventType
	{ //check eventtype id
		if trigger.EventTypeID == "" {
			return service.statusBadRequest(fmt.Errorf("TriggerDB.PostData failed: no eventTypeId was specified"))
		}
		var et *EventType
		var found bool
		et, found, err = service.eventTypeDB.GetOne(trigger.EventTypeID, trigger.CreatedBy)
		if !found || err != nil {
			return service.statusBadRequest(fmt.Errorf("TriggerDB.PostData failed: eventType %s could not be found", trigger.EventTypeID))
		}
		eventType = et
	}
	fixedQuery, ok := handleUniqueParams(trigger.Condition, eventType.Name, func(eventTypeName string, key string) string {
		return strings.Replace(key, "data.", "data."+eventTypeName+".", 1)
	}).(map[string]interface{})
	if !ok {
		return service.statusBadRequest(fmt.Errorf("TriggerEB.PostData failed: failed to parse query"))
	}
	response := *trigger
	trigger.Condition = fixedQuery

	service.syslogger.Audit(trigger.CreatedBy, "creatingTrigger", trigger.TriggerID, "Service.PostTrigger: User [%s] is creating trigger [%s]", trigger.CreatedBy, trigger.TriggerID)

	if err = service.triggerDB.PostData(trigger); err != nil {
		service.syslogger.Audit(trigger.CreatedBy, "creatingTriggerFailure", trigger.TriggerID, "Service.PostTrigger: User [%s] failed to create trigger [%s]", trigger.CreatedBy, trigger.TriggerID)
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit(trigger.CreatedBy, "createdTrigger", trigger.TriggerID, "Service.PostTrigger: User [%s] successfully created trigger [%s]", trigger.CreatedBy, trigger.TriggerID)

	service.stats.IncrTriggers()

	return service.statusCreated(&response)
}

func (service *Service) PutTrigger(id piazza.Ident, update *TriggerUpdate) *piazza.JsonResponse {
	defer service.handlePanic()
	trigger, found, err := service.triggerDB.GetOne(id, "pz-workflow")
	if !found {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit("pz-workflow", "updatingTrigger", id, "Service.PutTrigger: User is updating trigger [%s]", id)

	if _, err = service.triggerDB.PutTrigger(trigger, update, "pz-workflow"); err != nil {
		service.syslogger.Audit("pz-workflow", "updatingTriggerFailure", id, "Service.PutTrigger: User failed to update trigger [%s]", id)
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit("pz-workflow", "updatedTrigger", id, "Service.PutTrigger: User successfully updated trigger [%s] with enabled=[%v]", id, update.Enabled)

	return service.statusPutOK("Updated trigger")
}

func (service *Service) DeleteTrigger(id piazza.Ident) *piazza.JsonResponse {
	defer service.handlePanic()
	service.syslogger.Audit("pz-workflow", "deletingTrigger", id, "Service.DeleteTrigger: User is deleting trigger [%s]", id)

	ok, err := service.triggerDB.DeleteTrigger(id, "pz-workflow")
	if !ok {
		service.syslogger.Audit("pz-workflow", "deletingTriggerFailure", id, "Service.DeleteTrigger: User failed to delete trigger [%s]", id)
		return service.statusNotFound(err)
	}
	if err != nil {
		service.syslogger.Audit("pz-workflow", "deletingTriggerFailure", id, "Service.DeleteTrigger: User failed to delete trigger [%s]", id)
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit("pz-workflow", "deletedTrigger", id, "Service.DeleteTrigger: User successfully deleted trigger [%s]", id)

	return service.statusOK(nil)
}

//---------------------------------------------------------------------

func (service *Service) GetAlert(id piazza.Ident) *piazza.JsonResponse {
	defer service.handlePanic()
	service.syslogger.Audit("pz-workflow", "gettingAlert", id, "Service.GetAlert: User is getting alert [%s]", id)
	alert, found, err := service.alertDB.GetOne(id, "pz-workflow")
	if !found {
		service.syslogger.Audit("pz-workflow", "gettingAlertFailure", id, "Service.GetAlert: User failed to get alert [%s]", id)
		return service.statusNotFound(err)
	}
	if err != nil {
		service.syslogger.Audit("pz-workflow", "gettingAlertFailure", id, "Service.GetAlert: User failed to get alert [%s]", id)
		return service.statusBadRequest(err)
	}
	service.syslogger.Audit("pz-workflow", "gotAlert", id, "Service.GetAlert: User successfully got alert [%s]", id)

	return service.statusOK(alert)
}

func (service *Service) GetAllAlerts(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	defer service.handlePanic()
	triggerID, err := params.GetAsID("triggerId", "")
	if err != nil {
		return service.statusBadRequest(err)
	}

	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	var alerts []Alert
	var totalHits int64
	preresultaudit := time.Now().UTC().UnixNano()
	service.syslogger.Audit("pz-workflow", "gettingAllAlerts", service.alertDB.mapping, "Service.GetAllAlerts: User is getting all alerts")
	preresult := time.Now().UTC().UnixNano()
	if triggerID != "" && piazza.ValidUuid(triggerID.String()) {
		alerts, totalHits, err = service.alertDB.GetAllByTrigger(format, triggerID, "pz-workflow")
		if err != nil {
			service.syslogger.Audit("pz-workflow", "gettingAllAlertsFailure", service.alertDB.mapping, "Service.GetAllAlerts: User failed to get all alerts")
			return service.statusInternalError(err)
		} else if alerts == nil {
			service.syslogger.Audit("pz-workflow", "gettingAllAlertsFailure", service.alertDB.mapping, "Service.GetAllAlerts: User failed to get all alerts")
			return service.statusInternalError(errors.New("GetAllAlerts returned nil"))
		}
	} else if triggerID == "" {
		alerts, totalHits, err = service.alertDB.GetAll(format, "pz-workflow")
		if err != nil {
			service.syslogger.Audit("pz-workflow", "gettingAllAlertsFailure", service.alertDB.mapping, "Service.GetAllAlerts: User failed to get all alerts")
			return service.statusInternalError(err)
		} else if alerts == nil {
			service.syslogger.Audit("pz-workflow", "gettingAllAlertsFailure", service.alertDB.mapping, "Service.GetAllAlerts: User failed to get all alerts")
			return service.statusInternalError(errors.New("GetAllAlerts returned nil"))
		}
	} else {
		service.syslogger.Audit("pz-workflow", "gettingAllAlertsFailure", service.alertDB.mapping, "Service.GetAllAlerts: User failed to get all alerts")
		return service.statusBadRequest(errors.New("Malformed triggerId query parameter"))
	}
	postresult := time.Now().UTC().UnixNano()
	var resp *piazza.JsonResponse
	inflate := getInflateParam(params)

	if inflate {
		alertExts, err := service.inflateAlerts(alerts)
		if err != nil {
			service.syslogger.Audit("pz-workflow", "gettingAllAlertsFailure", service.alertDB.mapping, "Service.GetAllAlerts: User failed to get all alerts")
			return service.statusInternalError(err)
		}
		resp = service.statusOK(*alertExts)
	} else {
		resp = service.statusOK(alerts)
	}

	service.syslogger.Audit("pz-workflow", "gotAllAlerts", service.alertDB.mapping, "Service.GetAllAlerts: User successfully got all alerts")
	postresultaudit := time.Now().UTC().UnixNano()
	format.Count = int(totalHits)
	resp.Pagination = format

	log.Printf("Pre request Audit: %s", strconv.FormatInt(preresult - preresultaudit, 10))
	log.Printf("Request: %s", strconv.FormatInt(postresult - preresult, 10))
	log.Printf("Post request Audit: %s", strconv.FormatInt(postresultaudit - postresult, 10))

	return resp
}

func (service *Service) QueryAlerts(dslString string, params *piazza.HttpQueryParams) *piazza.JsonResponse {
	defer service.handlePanic()
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	var alerts []Alert
	var totalHits int64

	service.syslogger.Audit("pz-workflow", "queryingAlerts", service.alertDB.mapping, "Service.QueryAlerts: User is querying alerts")

	dslString, err = syncPagination(dslString, *format)
	if err != nil {
		service.syslogger.Audit("pz-workflow", "queryingAlertsFailure", service.alertDB.mapping, "Service.QueryAlerts: syncPagination failed")
		return service.statusBadRequest(err)
	}
	alerts, totalHits, err = service.alertDB.GetAlertsByDslQuery(dslString, "pz-workflow")
	if err != nil {
		service.syslogger.Audit("pz-workflow", "queryingAlertsFailure", service.alertDB.mapping, "Service.QueryAlerts: User failed to query alerts")
		return service.statusBadRequest(err)
	} else if alerts == nil {
		service.syslogger.Audit("pz-workflow", "queryingAlertsFailure", service.alertDB.mapping, "Service.QueryAlerts: User failed to query alerts")
		return service.statusInternalError(errors.New("QueryAlerts returned nil"))
	}

	var resp *piazza.JsonResponse
	inflate := getInflateParam(params)

	if inflate {
		alertExts, err := service.inflateAlerts(alerts)
		if err != nil {
			service.syslogger.Audit("pz-workflow", "queryingAlertsFailure", service.alertDB.mapping, "Service.QueryAlerts: User failed to query alerts")
			return service.statusInternalError(err)
		}
		resp = service.statusOK(*alertExts)
	} else {
		resp = service.statusOK(alerts)
	}

	service.syslogger.Audit("pz-workflow", "queriedAlerts", service.alertDB.mapping, "Service.QueryAlerts: User successfully queried alerts")

	format.Count = int(totalHits)
	resp.Pagination = format

	return resp
}

func getInflateParam(params *piazza.HttpQueryParams) bool {
	inflateString, err := params.GetAsString("inflate", "false")
	if err != nil {
		inflateString = "false"
	}
	inflate, err := strconv.ParseBool(inflateString)
	if err != nil {
		inflate = false
	}
	return inflate
}

func (service *Service) inflateAlerts(alerts []Alert) (*[]AlertExt, error) {
	alertExts := make([]AlertExt, len(alerts))
	for index, alert := range alerts {
		alertExt, err := service.inflateAlert(alert)
		if err != nil {
			return nil, err
		}
		alertExts[index] = *alertExt
	}
	return &alertExts, nil
}

func (service *Service) inflateAlert(alert Alert) (*AlertExt, error) {
	trigger, found, err := service.triggerDB.GetOne(alert.TriggerID, "pz-workflow")
	if err != nil || !found {
		trigger = &Trigger{TriggerID: alert.TriggerID}
	}

	mapping, err := service.eventDB.lookupEventTypeNameByEventID(alert.EventID, alert.CreatedBy)
	if mapping == "" || err != nil {
		// Do nothing
	}

	event, found, err := service.eventDB.GetOne(mapping, alert.EventID, "pz-workflow")
	if err != nil || !found {
		event = &Event{EventID: alert.EventID}
	}
	alertExt := &AlertExt{
		AlertID:   alert.AlertID,
		Trigger:   *trigger,
		Event:     *event,
		JobID:     alert.JobID,
		CreatedBy: alert.CreatedBy,
		CreatedOn: alert.CreatedOn,
	}
	return alertExt, nil
}

// PostAlert TODO
func (service *Service) PostAlert(alert *Alert) *piazza.JsonResponse {
	defer service.handlePanic()
	alert.AlertID = service.newIdent()
	alert.CreatedOn = piazza.NewTimeStamp()

	service.syslogger.Audit(alert.CreatedBy, "creatingAlert", alert.AlertID, "Service.PostAlert: User [%s] is creating alert [%s]", alert.CreatedBy, alert.AlertID)

	if err := service.alertDB.PostData(alert); err != nil {
		service.syslogger.Audit(alert.CreatedBy, "creatingAlertFailure", alert.AlertID, "Service.PostAlert: User [%s] failed to create alert [%s]", alert.CreatedBy, alert.AlertID)
		return service.statusInternalError(err)
	}

	service.syslogger.Audit(alert.CreatedBy, "createdAlert", alert.AlertID, "Service.PostAlert: User [%s] successfully created alert [%s]", alert.CreatedBy, alert.AlertID)

	service.stats.IncrAlerts()

	return service.statusCreated(alert)
}

// DeleteAlert TODO
func (service *Service) DeleteAlert(id piazza.Ident) *piazza.JsonResponse {
	defer service.handlePanic()
	service.syslogger.Audit("pz-workflow", "deletingAlert", id, "Service.DeleteAlert: User is deleteing alert [%s]", id)

	ok, err := service.alertDB.DeleteByID(id, "pz-workflow")
	if !ok {
		service.syslogger.Audit("pz-workflow", "deletingAlertFailure", id, "Service.DeleteAlert: User failed to delete alert [%s]", id)
		return service.statusNotFound(err)
	}
	if err != nil {
		service.syslogger.Audit("pz-workflow", "deletingAlertFailure", id, "Service.DeleteAlert: User failed to delete alert [%s]", id)
		return service.statusBadRequest(err)
	}

	service.syslogger.Audit("pz-workflow", "deletedAlert", id, "Service.DeleteAlert: User successfully deleted alert [%s]", id)

	return service.statusOK(nil)
}

func (service *Service) addUniqueParams(uniqueKey string, inputObj map[string]interface{}) map[string]interface{} {
	outputObj := map[string]interface{}{}
	outputObj[uniqueKey] = inputObj
	return outputObj
}
func (service *Service) removeUniqueParams(uniqueKey string, inputObj map[string]interface{}) map[string]interface{} {
	if _, ok := inputObj[uniqueKey]; !ok {
		return inputObj
	}
	return inputObj[uniqueKey].(map[string]interface{})
}

//---------------------------------------------------------------------

// InitCron TODO
func (service *Service) InitCron() error {
	defer service.handlePanic()
	ok, err := service.cronDB.Exists("pz-workflow")
	if err != nil {
		return err
	}
	if ok {
		events, err := service.cronDB.GetAll("pz-workflow")
		if err != nil {
			return LoggedError("WorkflowService.InitCron: Unable to get all from CronDB")
		}

		for _, e := range *events {
			eventType, found, err := service.eventTypeDB.GetOne(e.EventTypeID, "pz-workflow")
			if !found || err != nil {
				return LoggedError("WorkflowService.InitCron: Unable to retrieve event type for cron event %#v", e)
			}
			if err = service.cron.AddJob(e.CronSchedule, cronEvent{&e, eventType.Name, service}); err != nil {
				return LoggedError("WorkflowService.InitCron: Unable to register cron event %#v", e)
			}
		}
	}

	service.cron.Start()

	return nil
}

type cronEvent struct {
	*Event
	eventTypeName string
	service       *Service
}

func (c cronEvent) Run() {
	uniqueMap := c.Data[c.eventTypeName]
	if uniqueMap == nil {
		uniqueMap = make(map[string]interface{})
	}
	ev := &Event{
		EventTypeID: c.EventTypeID,
		Data:        uniqueMap.(map[string]interface{}),
		CreatedOn:   piazza.NewTimeStamp(),
		CreatedBy:   c.EventID.String(),
	}
	c.service.PostEvent(ev)
}

func (c cronEvent) Key() string {
	return c.EventID.String()
}

//---------------------------------------------------------------------

func (service *Service) TestElasticsearchVersion() *piazza.JsonResponse {
	defer service.handlePanic()
	version, err := service.testElasticsearchDB.GetVersion()
	if err != nil {
		return service.statusBadRequest(err)
	}
	if version == "" {
		return service.statusInternalError(errors.New("Service.TestElasticsearchVersion returned nil"))
	}
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: version}

	return resp
}

func (service *Service) TestElasticsearchGetOne(id piazza.Ident) *piazza.JsonResponse {
	defer service.handlePanic()
	body, found, err := service.testElasticsearchDB.GetOne(id)
	if !found {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}
	return service.statusOK(body)
}

func (service *Service) TestElasticsearchPost(body *TestElasticsearchBody) *piazza.JsonResponse {
	defer service.handlePanic()
	body.ID = service.newIdent()

	if _, err := service.testElasticsearchDB.PostData(body, body.ID); err != nil {
		if strings.HasSuffix(err.Error(), "was not recognized as a valid mapping type") {
			return service.statusBadRequest(err)
		}
		return service.statusInternalError(err)
	}

	return service.statusCreated(body)
}
