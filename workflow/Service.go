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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	syslogger "github.com/venicegeo/pz-gocommon/syslog"
	pzlogger "github.com/venicegeo/pz-logger/logger"
	pzuuidgen "github.com/venicegeo/pz-uuidgen/uuidgen"
	cron "github.com/venicegeo/vegertar-cron"
)

//------------------------------------------------------------------------------

const ingestTypeName = "piazza:ingest"
const executeTypeName = "piazza:executionComplete"

type Service struct {
	eventTypeDB         *EventTypeDB
	eventDB             *EventDB
	triggerDB           *TriggerDB
	alertDB             *AlertDB
	cronDB              *CronDB
	testElasticsearchDB *TestElasticsearchDB

	stats Stats
	sync.Mutex

	syslogger *syslogger.Logger
	uuidgen   pzuuidgen.IClient

	sys *piazza.SystemConfig

	cron *cron.Cron

	origin string
}

//------------------------------------------------------------------------------

// Init TODO
func (service *Service) Init(
	sys *piazza.SystemConfig,
	logger pzlogger.IClient,
	uuidgen pzuuidgen.IClient,
	eventtypesIndex elasticsearch.IIndex,
	eventsIndex elasticsearch.IIndex,
	triggersIndex elasticsearch.IIndex,
	alertsIndex elasticsearch.IIndex,
	cronIndex elasticsearch.IIndex,
	testElasticsearchIndex elasticsearch.IIndex) error {

	service.sys = sys

	service.stats.CreatedOn = time.Now()

	var err error

	writer := &pzlogger.SyslogElkWriter{
		Client: logger,
	}

	service.syslogger = syslogger.NewLogger(writer, "pz-workflow")

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

	service.testElasticsearchDB, err = NewTestElasticsearchDB(service, testElasticsearchIndex)
	if err != nil {
		return err
	}

	service.cron = cron.New()

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

	_, err = elasticsearch.PollFunction(pollingFn)
	if err != nil {
		//log.Printf("ERROR: %#v", err)
		return err
	}
	//log.Printf("SETUP INDEX: %t", ok)

	// Ingest event type
	ingestEventType := &EventType{}
	ingestEventType.Name = ingestTypeName
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
	if postedIngestEventType.StatusCode == 201 {
		// everything is ok
		service.syslogger.Info("  SUCCESS Created piazza:ingest eventtype: %s", postedIngestEventType.StatusCode)
	} else {
		// something is wrong
		service.syslogger.Info("  ERROR creating piazza:ingest eventtype: %s", postedIngestEventType.StatusCode)
	}

	// Execution Completed event type
	executionCompletedType := &EventType{}
	executionCompletedType.Name = executeTypeName
	executionCompletedTypeMapping := map[string]interface{}{
		"jobId":  "string",
		"status": "string",
		"dataId": "string",
	}
	executionCompletedType.Mapping = executionCompletedTypeMapping
	postedExecutionCompletedType := service.PostEventType(executionCompletedType)
	if postedExecutionCompletedType.StatusCode == 201 {
		// everything is ok
		service.syslogger.Info("  SUCCESS Created piazza:executionComplete eventtype: %s", postedExecutionCompletedType.StatusCode)
	} else {
		// something is wrong or it was already there
		service.syslogger.Info("  ERROR creating piazza:excutionComplete eventtype: %s", postedExecutionCompletedType.StatusCode)
	}

	service.origin = string(sys.Name)

	return nil
}

func (service *Service) newIdent() (piazza.Ident, error) {
	uuid, err := service.uuidgen.GetUUID()
	if err != nil {
		return piazza.NoIdent, err
	}

	return piazza.Ident(uuid), nil
}

func (service *Service) sendToKafka(jobInstance string, jobID piazza.Ident, actor string) error {
	service.syslogger.Audit(actor, "createJob", "kafka", "User [%s] is sending job [%s] to kafka", actor, jobID.String())
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
		if err2 := producer.Close(); err2 != nil {
			log.Fatalf("Kafka-related failure (3): " + err2.Error())
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

func (service *Service) statusOK(obj interface{}) *piazza.JsonResponse {
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: obj}
	err := resp.SetType()
	if err != nil {
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
	err := resp.SetType()
	if err != nil {
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
	service.Lock()
	t := service.stats
	service.Unlock()
	return service.statusOK(t)
}

//------------------------------------------------------------------------------

// GetEventType TODO
func (service *Service) GetEventType(id piazza.Ident, actor string) *piazza.JsonResponse {
	eventType, found, err := service.eventTypeDB.GetOne(id, actor)
	if !found {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}
	eventType.Mapping = service.removeUniqueParams(eventType.Name, eventType.Mapping)
	return service.statusOK(eventType)
}

// GetAllEventTypes TODO
func (service *Service) GetAllEventTypes(params *piazza.HttpQueryParams) *piazza.JsonResponse {
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
	if nameParam != "" {
		nameParamValue := nameParam
		var foundName bool
		var eventtypeid *piazza.Ident
		eventtypeid, foundName, err = service.eventTypeDB.GetIDByName(nameParamValue, "pz-workflow")
		var foundType = false
		var eventtype *EventType
		if foundName && eventtypeid != nil {
			if err != nil {
				return service.statusBadRequest(err)
			}
			eventtype, foundType, err = service.eventTypeDB.GetOne(piazza.Ident(eventtypeid.String()), "pz-workflow")
			if err != nil {
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
			return service.statusInternalError(err)
		}
	}
	if eventtypes == nil {
		return service.statusInternalError(errors.New("getalleventtypes returned nil"))
	}
	for i := 0; i < len(eventtypes); i++ {
		eventtypes[i].Mapping = service.removeUniqueParams(eventtypes[i].Name, eventtypes[i].Mapping)
	}
	resp := service.statusOK(eventtypes)

	if totalHits > 0 {
		format.Count = int(totalHits)
		resp.Pagination = format
	}

	return resp
}

func (service *Service) QueryEventTypes(dslString string, params *piazza.HttpQueryParams) *piazza.JsonResponse {
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	var totalHits int64
	var eventtypes []EventType
	dslString, err = syncPagination(dslString, *format)
	if err != nil {
		return service.statusBadRequest(err)
	}
	eventtypes, totalHits, err = service.eventTypeDB.GetEventTypesByDslQuery(dslString, "pz-workflow")
	if err != nil {
		return service.statusBadRequest(err)
	}
	if eventtypes == nil {
		return service.statusInternalError(errors.New("queryeventtypes returned nil"))
	}
	for i := 0; i < len(eventtypes); i++ {
		eventtypes[i].Mapping = service.removeUniqueParams(eventtypes[i].Name, eventtypes[i].Mapping)
	}
	resp := service.statusOK(eventtypes)

	if totalHits > 0 {
		format.Count = int(totalHits)
		resp.Pagination = format
	}

	return resp
}

// PostEventType TODO
func (service *Service) PostEventType(eventType *EventType) *piazza.JsonResponse {
	// Check if our EventType.Name already exists
	name := eventType.Name
	found, err := service.eventDB.NameExists(name, eventType.CreatedBy)
	if err != nil {
		return service.statusInternalError(err)
	}
	if found {
		return service.statusBadRequest(LoggedError("EventType Name already exists"))
	}
	id1, found, err := service.eventTypeDB.GetIDByName(name, eventType.CreatedBy)
	if err != nil {
		return service.statusInternalError(err)
	}
	if found {
		return service.statusBadRequest(
			LoggedError("EventType Name already exists under EventTypeId %s", id1))
	}

	eventTypeID, err := service.newIdent()
	if err != nil {
		return service.statusInternalError(err)
	}
	eventType.EventTypeID = eventTypeID

	eventType.CreatedOn = time.Now()

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

	id, err := service.eventTypeDB.PostData(eventType, eventTypeID, eventType.CreatedBy)
	if err != nil {
		if strings.HasSuffix(err.Error(), "was not recognized as a valid mapping type") {
			return service.statusBadRequest(err)
		}
		return service.statusInternalError(err)
	}

	err = service.eventDB.AddMapping(name, eventType.Mapping, eventType.CreatedBy)
	if err != nil {
		_, _ = service.eventTypeDB.DeleteByID(id, eventType.CreatedBy)
		return service.statusInternalError(err)
	}

	go func() {
		service.syslogger.Info("User %s created EventType %s", eventType.CreatedBy, eventTypeID)
	}()

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
	eventType, found, err := service.eventTypeDB.GetOne(id, "pz-workflow")
	// Only check for system events or "in use" if found
	if found {
		if eventType != nil && IsSystemEvent(eventType.Name) {
			return service.statusBadRequest(errors.New("Deleting system eventTypes is prohibited"))
		}

		var triggers []Trigger
		var hits int64
		triggers, hits, err = service.triggerDB.GetTriggersByEventTypeID(id, "pz-workflow")
		if err != nil {
			return service.statusBadRequest(err)
		}
		if hits > 0 || len(triggers) > 0 {
			return service.statusForbidden(errors.New("Deleting eventTypes that are in use is prohibited"))
		}

		var events []Event
		events, hits, err = service.eventDB.GetEventsByEventTypeID(eventType.Name, id, "pz-workflow")
		if err != nil {
			return service.statusBadRequest(err)
		}
		if hits > 0 || len(events) > 0 {
			return service.statusForbidden(errors.New("Deleting eventTypes that are in use is prohibited"))
		}
	}
	ok, err := service.eventTypeDB.DeleteByID(id, "pz-workflow")
	if !ok {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.syslogger.Info("Deleted EventType %s", id)

	return service.statusOK(nil)
}

//------------------------------------------------------------------------------

// GetEvent TODO
func (service *Service) GetEvent(id piazza.Ident) *piazza.JsonResponse {
	mapping, err := service.eventDB.lookupEventTypeNameByEventID(id, "pz-workflow")
	if mapping == "" {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	event, found, err := service.eventDB.GetOne(mapping, id, "pz-workflow")

	if !found {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	event.Data = service.removeUniqueParams(mapping, event.Data)
	return service.statusOK(event)
}

// GetAllEvents TODO
func (service *Service) GetAllEvents(params *piazza.HttpQueryParams) *piazza.JsonResponse {
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

	events, totalHits, err := service.eventDB.GetAll(query, format, "pz-workflow")
	if err != nil {
		return service.statusInternalError(err)
	}
	for i := 0; i < len(events); i++ {
		eventType, found, err := service.eventTypeDB.GetOne(events[i].EventTypeID, "pz-workflow")
		if !found || err != nil {
			return service.statusInternalError(err)
		}
		events[i].Data = service.removeUniqueParams(eventType.Name, events[i].Data)
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
func (service *Service) PostRepeatingEvent(event *Event) *piazza.JsonResponse {
	// Post the event in the database, WITHOUT "triggering"
	eventTypeID := event.EventTypeID
	eventType, found, err := service.eventTypeDB.GetOne(eventTypeID, event.CreatedBy)
	if err != nil || !found {
		return service.statusBadRequest(err)
	}
	eventTypeName := eventType.Name

	log.Println("Posted Repeating Event")
	_, err = cron.Parse(event.CronSchedule)
	if err != nil {
		return service.statusBadRequest(err)
	}

	eventID, err := service.newIdent()
	if err != nil {
		return service.statusInternalError(err)
	}
	event.EventID = eventID

	event.CreatedOn = time.Now()

	response := *event

	event.Data = service.addUniqueParams(eventTypeName, event.Data)

	err = service.cron.AddJob(event.CronSchedule, cronEvent{event, eventTypeName, service})
	if err != nil {
		return service.statusInternalError(err)
	}

	err = service.cronDB.PostData(event, eventID, event.CreatedBy)
	if err != nil {
		return service.statusInternalError(err)
	}

	_, err = service.eventDB.PostData(eventTypeName, event, eventID, event.CreatedBy)
	if err != nil {
		// If we fail, need to also remove from cronDB
		// We don't check for errors here because if we've reached this point,
		// the eventID will be in the cronDB
		_, _ = service.cronDB.DeleteByID(eventID, event.CreatedBy)
		service.cron.Remove(eventID.String())
		return service.statusInternalError(err)
	}

	service.stats.IncrEvents()

	service.syslogger.Info("User %s created repeating Event %s on the schedule %s", event.CreatedBy, eventID, event.CronSchedule)

	return service.statusCreated(&response)
}

// PostEvent TODO
func (service *Service) PostEvent(event *Event) *piazza.JsonResponse {
	eventTypeID := event.EventTypeID
	eventType, found, err := service.eventTypeDB.GetOne(eventTypeID, event.CreatedBy)
	if err != nil || !found {
		return service.statusBadRequest(err)
	}
	eventTypeName := eventType.Name

	eventID, err := service.newIdent()
	if err != nil {
		return service.statusInternalError(err)
	}
	event.EventID = eventID

	event.CreatedOn = time.Now()

	response := *event

	event.Data = service.addUniqueParams(eventTypeName, event.Data)

	_, err = service.eventDB.PostData(eventTypeName, event, eventID, event.CreatedBy)
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.syslogger.Info("User %s created Event %s", event.CreatedBy, eventID)

	{
		// Find triggers associated with event
		triggerIDs, err := service.eventDB.PercolateEventData(eventTypeName, event.Data, eventID, event.CreatedBy)
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

				trigger, found, err := service.triggerDB.GetOne(triggerID, event.CreatedBy)
				if !found {
					// Don't fail for this, just log something and continue to the next trigger id
					service.syslogger.Warning("Percolation error: Trigger %s does not exist", string(triggerID))
					return
				}
				if err != nil {
					results[triggerID] = service.statusBadRequest(err)
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
				params := event.Data[eventTypeName]
				for key, value := range params.(map[string]interface{}) {
					jobString = strings.Replace(jobString, "$"+key, fmt.Sprintf("%v", value), -1)
				}

				service.syslogger.Info("job [%s] submission by event [%s] using trigger [%s]: %s\n", jobID, eventID.String(), triggerID.String(), jobString)

				log.Printf("JOB ID: %s", jobID)
				log.Printf("JOB STRING: %s", jobString)

				err = service.sendToKafka(jobString, jobID, trigger.CreatedBy)
				if err != nil {
					results[triggerID] = service.statusInternalError(err)
					return
				}

				service.stats.IncrTriggerJobs()

				alert := Alert{EventID: eventID, TriggerID: triggerID, JobID: jobID, CreatedBy: trigger.CreatedBy}
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

	return service.statusCreated(&response)
}

func (service *Service) QueryEvents(jsonString string, params *piazza.HttpQueryParams) *piazza.JsonResponse {
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

	jsonString, err = syncPagination(jsonString, *format)
	if err != nil {
		return service.statusBadRequest(err)
	}
	events, totalHits, err := service.eventDB.GetEventsByDslQuery(query, jsonString, "pz-workflow")
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
		err = json.Unmarshal(bts, &g)
		if err != nil {
			return "", err
		}
		sortDsl := g.([]interface{})
		dsl["sort"] = sortDsl
	}
	byteArray, err := json.Marshal(dsl)
	if err != nil {
		return "", err
	}
	s := string(byteArray)
	return s, nil
}

func (service *Service) DeleteEvent(id piazza.Ident) *piazza.JsonResponse {
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

	ok, err := service.eventDB.DeleteByID(mapping, id, "pz-workflow")
	if !ok {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	// If it's a cron event, remove from cronDB, stop cronjob
	ok, err = service.cronDB.itemExists(id, "pz-workflow")
	if err != nil {
		return service.statusBadRequest(err)
	}
	if ok {
		ok, err := service.cronDB.DeleteByID(id, "pz-workflow")
		if !ok {
			return service.statusNotFound(err)
		}
		if err != nil {
			return service.statusBadRequest(err)
		}
		service.cron.Remove(id.String())
	}

	service.syslogger.Info("Deleted Event with EventId %s", id)

	return service.statusOK(nil)
}

//------------------------------------------------------------------------------

func (service *Service) GetTrigger(id piazza.Ident) *piazza.JsonResponse {
	trigger, found, err := service.triggerDB.GetOne(id, "pz-workflow")
	if !found {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}
	eventType, found, err := service.eventTypeDB.GetOne(trigger.EventTypeID, "pz-workflow")
	if err != nil || !found {
		return service.statusBadRequest(err)
	}

	trigger.Condition = service.removeUniqueParams(eventType.Name, trigger.Condition)
	return service.statusOK(trigger)
}

func (service *Service) GetAllTriggers(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	triggers, totalHits, err := service.triggerDB.GetAll(format, "pz-workflow")
	if err != nil {
		return service.statusInternalError(err)
	} else if triggers == nil {
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

	if totalHits > 0 {
		format.Count = int(totalHits)
		resp.Pagination = format
	}

	return resp
}

func (service *Service) QueryTriggers(dslString string, params *piazza.HttpQueryParams) *piazza.JsonResponse {
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	dslString, err = syncPagination(dslString, *format)
	triggers, totalHits, err := service.triggerDB.GetTriggersByDslQuery(dslString, "pz-workflow")
	if err != nil {
		return service.statusBadRequest(err)
	} else if triggers == nil {
		return service.statusInternalError(errors.New("QueryTriggers returned nil"))
	}
	for i := 0; i < len(triggers); i++ {
		eventType, found, err := service.eventTypeDB.GetOne(triggers[i].EventTypeID, "pz-workflow")
		if err != nil || !found {
			return service.statusBadRequest(err)
		}
		triggers[i].Condition = service.removeUniqueParams(eventType.Name, triggers[i].Condition)
	}

	resp := service.statusOK(triggers)

	if totalHits > 0 {
		format.Count = int(totalHits)
		resp.Pagination = format
	}

	return resp
}

func (service *Service) PostTrigger(trigger *Trigger) *piazza.JsonResponse {
	triggerID, err := service.newIdent()
	if err != nil {
		return service.statusBadRequest(err)
	}
	trigger.TriggerID = triggerID
	trigger.CreatedOn = time.Now()

	eventType := &EventType{}
	{ //check eventtype id
		if trigger.EventTypeID == "" {
			return service.statusBadRequest(fmt.Errorf("TriggerDB.PostData failed: no eventTypeId was specified"))
		}
		var et *EventType
		var found bool
		et, found, err = service.eventTypeDB.GetOne(trigger.EventTypeID, trigger.CreatedBy)
		if !found || err != nil {
			return service.statusBadRequest(fmt.Errorf("TriggerDB.PostData failed: eventType %s could not be found", trigger.EventTypeID.String()))
		}
		eventType = et
	}
	fixedQuery, ok := service.triggerDB.addUniqueParamsToQuery(trigger.Condition, eventType).(map[string]interface{})
	if !ok {
		return service.statusBadRequest(fmt.Errorf("TriggerEB.PostData failed: failed to parse query"))
	}
	response := *trigger
	trigger.Condition = fixedQuery

	_, err = service.triggerDB.PostTrigger(trigger, triggerID, trigger.CreatedBy)
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.syslogger.Info("User %s created Trigger %s", trigger.CreatedBy, triggerID)

	service.stats.IncrTriggers()

	return service.statusCreated(&response)
}

func (service *Service) PutTrigger(id piazza.Ident, update *TriggerUpdate) *piazza.JsonResponse {
	trigger, found, err := service.triggerDB.GetOne(id, "pz-workflow")
	if !found {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}
	_, err = service.triggerDB.PutTrigger(id, trigger, update, "pz-workflow")
	if err != nil {
		return service.statusBadRequest(err)
	}
	service.syslogger.Info("Updated Trigger %s with enabled=%v", id, update.Enabled)

	return service.statusPutOK("Updated trigger")
}

func (service *Service) DeleteTrigger(id piazza.Ident) *piazza.JsonResponse {
	ok, err := service.triggerDB.DeleteTrigger(id, "pz-workflow")
	if !ok {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.syslogger.Info("Deleted Trigger with TriggerId %s", id)

	return service.statusOK(nil)
}

//---------------------------------------------------------------------

func (service *Service) GetAlert(id piazza.Ident) *piazza.JsonResponse {
	alert, found, err := service.alertDB.GetOne(id, "pz-workflow")
	if !found {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	return service.statusOK(alert)
}

func (service *Service) GetAllAlerts(params *piazza.HttpQueryParams) *piazza.JsonResponse {
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

	if triggerID != "" && isUUID(triggerID) {
		alerts, totalHits, err = service.alertDB.GetAllByTrigger(format, triggerID, "pz-workflow")
		if err != nil {
			return service.statusInternalError(err)
		} else if alerts == nil {
			return service.statusInternalError(errors.New("GetAllAlerts returned nil"))
		}
	} else if triggerID == "" {
		alerts, totalHits, err = service.alertDB.GetAll(format, "pz-workflow")
		if err != nil {
			return service.statusInternalError(err)
		} else if alerts == nil {
			return service.statusInternalError(errors.New("GetAllAlerts returned nil"))
		}
	} else {
		return service.statusBadRequest(errors.New("Malformed triggerId query parameter"))
	}

	var resp *piazza.JsonResponse
	inflate := getInflateParam(params)

	if inflate {
		alertExts, err := service.inflateAlerts(alerts)
		if err != nil {
			return service.statusInternalError(err)
		}
		resp = service.statusOK(*alertExts)
	} else {
		resp = service.statusOK(alerts)
	}

	if totalHits > 0 {
		format.Count = int(totalHits)
		resp.Pagination = format
	}

	return resp
}

func (service *Service) QueryAlerts(dslString string, params *piazza.HttpQueryParams) *piazza.JsonResponse {
	format, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.statusBadRequest(err)
	}

	var alerts []Alert
	var totalHits int64

	dslString, err = syncPagination(dslString, *format)
	alerts, totalHits, err = service.alertDB.GetAlertsByDslQuery(dslString, "pz-workflow")
	if err != nil {
		return service.statusBadRequest(err)
	} else if alerts == nil {
		return service.statusInternalError(errors.New("QueryAlerts returned nil"))
	}

	var resp *piazza.JsonResponse
	inflate := getInflateParam(params)

	if inflate {
		alertExts, err := service.inflateAlerts(alerts)
		if err != nil {
			return service.statusInternalError(err)
		}
		resp = service.statusOK(*alertExts)
	} else {
		resp = service.statusOK(alerts)
	}

	if totalHits > 0 {
		format.Count = int(totalHits)
		resp.Pagination = format
	}

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
	alertID, err := service.newIdent()
	if err != nil {
		return service.statusBadRequest(err)
	}
	alert.AlertID = alertID

	alert.CreatedOn = time.Now()

	_, err = service.alertDB.PostData(&alert, alertID, alert.CreatedBy)
	if err != nil {
		return service.statusInternalError(err)
	}

	service.syslogger.Info("User %s created Alert %s", alert.CreatedBy, alertID)

	service.stats.IncrAlerts()

	return service.statusCreated(alert)
}

// DeleteAlert TODO
func (service *Service) DeleteAlert(id piazza.Ident) *piazza.JsonResponse {
	ok, err := service.alertDB.DeleteByID(id, "pz-workflow")
	if !ok {
		return service.statusNotFound(err)
	}
	if err != nil {
		return service.statusBadRequest(err)
	}

	service.syslogger.Info("Deleted Alert with AlertId %s", id)

	return service.statusOK(nil)
}

func (service *Service) addUniqueParams(uniqueKey string, inputObj map[string]interface{}) map[string]interface{} {
	outputObj := map[string]interface{}{}
	outputObj[uniqueKey] = inputObj
	return outputObj
}
func (service *Service) removeUniqueParams(uniqueKey string, inputObj map[string]interface{}) map[string]interface{} {
	_, ok := inputObj[uniqueKey]
	if !ok {
		return inputObj
	}
	return inputObj[uniqueKey].(map[string]interface{})
}

//---------------------------------------------------------------------

// InitCron TODO
func (service *Service) InitCron() error {
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
			err = service.cron.AddJob(e.CronSchedule, cronEvent{&e, eventType.Name, service})
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
		CreatedOn:   time.Now(),
		CreatedBy:   c.EventID.String(),
	}
	c.service.PostEvent(ev)
}

func (c cronEvent) Key() string {
	return c.EventID.String()
}

//---------------------------------------------------------------------

func (service *Service) TestElasticsearchVersion() *piazza.JsonResponse {

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
	id, err := service.newIdent()
	if err != nil {
		return service.statusInternalError(err)
	}
	body.ID = id

	_, err = service.testElasticsearchDB.PostData(body, id)
	if err != nil {
		if strings.HasSuffix(err.Error(), "was not recognized as a valid mapping type") {
			return service.statusBadRequest(err)
		}
		return service.statusInternalError(err)
	}

	r := service.statusCreated(body)
	return r
}
