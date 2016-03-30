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
	"log"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
)

type ServerTester struct {
	suite.Suite
	sys      *piazza.SystemConfig
	workflow *PzWorkflowService
}

func assertNoData(t *testing.T, workflow *PzWorkflowService) {
	assert := assert.New(t)

	var err error

	ts, err := workflow.GetAllEventTypes()
	if err == nil {
		assert.Len(*ts, 0)
	}

	es, err := workflow.GetAllEvents("")
	if err == nil {
		assert.Len(*es, 0)
	}

	as, err := workflow.GetAllAlerts()
	if err == nil {
		assert.Len(*as, 0)
	}

	xs, err := workflow.GetAllTriggers()
	if err == nil {
		assert.Len(*xs, 0)
	}
}

func TestRunSuite(t *testing.T) {

	endpoints := &piazza.ServicesMap{
		piazza.PzElasticSearch: "https://search-venice-es-pjebjkdaueu2gukocyccj4r5m4.us-east-1.es.amazonaws.com",
		piazza.PzLogger:        "",
	}

	sys, err := piazza.NewSystemConfig(piazza.PzWorkflow, endpoints)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := loggerPkg.NewMockLoggerService(sys)
	if err != nil {
		log.Fatal(err)
	}
	clogger := loggerPkg.NewCustomLogger(&logger, piazza.PzWorkflow, sys.Address)

	uuidgen, err := uuidgenPkg.NewMockUuidGenService(sys)
	if err != nil {
		log.Fatal(err)
	}

	es, err := elasticsearch.NewClient(sys, true)
	if err != nil {
		log.Fatal(err)
	}

	// start server
	{
		routes, err := CreateHandlers(sys, clogger, uuidgen, es)
		if err != nil {
			log.Fatal(err)
		}

		_ = sys.StartServer(routes)
	}

	workflow, err := NewPzWorkflowService(sys, clogger, es)
	if err != nil {
		log.Fatal(err)
	}

	serverTester := &ServerTester{workflow: workflow, sys: sys}
	suite.Run(t, serverTester)

	clientTester := &ClientTester{workflow: workflow, sys: sys}
	suite.Run(t, clientTester)
}

//---------------------------------------------------------------------------

func (suite *ServerTester) SetupSuite() {
	assertNoData(suite.T(), suite.workflow)
}
func (suite *ServerTester) TearDownSuite() {
	assertNoData(suite.T(), suite.workflow)
}

//---------------------------------------------------------------------------

func makeTestEventTypeName() string {
	return "MYTYPE"
}

func makeTestEventType(eventTypeName string) *EventType {
	mapping := map[string]elasticsearch.MappingElementTypeName{
		"num": elasticsearch.MappingElementTypeInteger,
	}
	return &EventType{Name: eventTypeName, Mapping: mapping}
}

func makeTestEvent(eventTypeID Ident) *Event {
	event := &Event{
		EventTypeID: eventTypeID,
		Date:        time.Now(),
		Data: map[string]interface{}{
			"num": 17,
		},
	}
	return event
}

func makeTestTrigger(eventTypeID Ident) *Trigger {
	trigger := &Trigger{
		Title: "MY TRIGGER TITLE",
		Condition: Condition{
			EventTypeID: eventTypeID,
			Query: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"num": 31,
					},
				},
			},
		},
		Job: Job{
			Task: "MY TASK",
		},
	}
	return trigger
}

//---------------------------------------------------------------------------
//---------------------------------------------------------------------------
func (suite *ServerTester) Test01EventType() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), suite.workflow)
	defer assertNoData(suite.T(), suite.workflow)

	log.Printf("Getting list of event types:")
	typs, err := workflow.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*typs, 0)
	printJSON("EventTypes", typs)
	log.Printf("Creating new Event Type:")

	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)
	printJSON("event type", eventType)

	id, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)

	printJSON("event type id", id)

	log.Printf("Getting list of event types:")
	typs, err = workflow.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*typs, 1)

	printJSON("EventTypes", typs)

	log.Printf("Getting event type by Id: %s", id)
	typ, err := workflow.GetOneEventType(id)
	assert.NoError(err)
	assert.EqualValues(string(id), string(typ.ID))

	printJSON("Got Event type", typ)
	log.Printf("Deleting Event type by Id: %s", id)

	err = workflow.DeleteOneEventType(id)
	assert.NoError(err)

	log.Printf("Getting list of event types:")
	typs, err = workflow.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*typs, 0)

	printJSON("EventTypes", typs)
}

func (suite *ServerTester) Test02Event() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), suite.workflow)
	defer assertNoData(suite.T(), suite.workflow)

	log.Printf("Getting list of events:")
	events, err := workflow.GetAllEvents("")
	assert.NoError(err)
	assert.Len(*events, 0)
	printJSON("Events", events)

	log.Printf("Creating new event type:")
	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)
	printJSON("event type", eventType)
	eventTypeID, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)
	printJSON("event type id", eventTypeID)

	log.Printf("Creating new event:")
	event := makeTestEvent(eventTypeID)
	printJSON("event", event)
	id, err := workflow.PostOneEvent(eventTypeName, event)
	assert.NoError(err)
	printJSON("event id", id)

	log.Printf("Getting list of events:")
	events, err = workflow.GetAllEvents(eventTypeName)
	assert.NoError(err)
	assert.Len(*events, 1)
	printJSON("Events", events)

	log.Printf("Getting event by id: %s", id)
	event, err = workflow.GetOneEvent(eventTypeName, id)
	printJSON("Got event", event)
	assert.NoError(err)
	assert.EqualValues(string(id), string(event.ID))

	log.Printf("Deleting event by id: %s", id)
	err = workflow.DeleteOneEvent(eventTypeName, id)
	assert.NoError(err)

	log.Printf("Getting list of events:")
	events, err = workflow.GetAllEvents(eventTypeName)
	assert.NoError(err)
	assert.Len(*events, 0)
	printJSON("Events", events)

	log.Printf("Deleting event type by id: %s", eventTypeID)
	err = workflow.DeleteOneEventType(eventTypeID)
	assert.NoError(err)
}

func (suite *ServerTester) Test03Trigger() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), suite.workflow)
	defer assertNoData(suite.T(), suite.workflow)

	log.Printf("Getting list of triggers:")
	triggers, err := workflow.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*triggers, 0)

	printJSON("triggers", triggers)

	log.Printf("Creating new event type:")
	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)
	printJSON("event type", eventType)
	eventTypeID, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)
	printJSON("event type id", eventTypeID)

	log.Printf("Creating new trigger:")
	trigger := makeTestTrigger(eventTypeID)
	printJSON("trigger", trigger)
	id, err := workflow.PostOneTrigger(trigger)
	printJSON("trigger id", id)

	log.Printf("Getting list of triggers:")
	triggers, err = workflow.GetAllTriggers()
	assert.NoError(err)
	printJSON("triggers", triggers)

	log.Printf("Getting trigger by id: %s", id)
	trigger, err = workflow.GetOneTrigger(id)
	assert.NoError(err)
	assert.EqualValues(string(id), string(trigger.ID))
	printJSON("Trigger", trigger)

	log.Printf("Delete trigger by id: %s", id)
	err = workflow.DeleteOneTrigger(id)
	assert.NoError(err)

	log.Printf("Getting list of triggers:")
	triggers, err = workflow.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*triggers, 0)
	printJSON("triggers", triggers)

	log.Printf("Delete event type by id: %s", eventTypeID)
	err = workflow.DeleteOneEventType(eventTypeID)
	assert.NoError(err)
}

func (suite *ServerTester) Test04Alert() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), suite.workflow)
	defer assertNoData(suite.T(), suite.workflow)

	log.Printf("Getting list of alerts:")
	alerts, err := workflow.GetAllAlerts()
	assert.NoError(err)
	printJSON("alerts:", alerts)

	log.Printf("Creating new event type:")
	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)
	printJSON("event type", eventType)
	eventTypeID, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)
	printJSON("event type id:", eventTypeID)

	log.Printf("Creating new trigger:")
	trigger := makeTestTrigger(eventTypeID)
	printJSON("Trigger", trigger)
	triggerID, err := workflow.PostOneTrigger(trigger)
	assert.NoError(err)
	printJSON("Trigger ID", triggerID)

	log.Printf("Creating new event:")
	event := makeTestEvent(eventTypeID)
	printJSON("event", event)
	eventID, err := workflow.PostOneEvent(eventTypeName, event)
	assert.NoError(err)
	printJSON("eventID", eventID)

	log.Printf("Creating new alert:")
	alert := &Alert{
		TriggerID: triggerID,
		EventID:   eventID,
	}
	printJSON("alert", alert)
	id, err := workflow.PostOneAlert(alert)
	assert.NoError(err)
	printJSON("alert id", id)

	log.Printf("Getting list of alerts:")
	alerts, err = workflow.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 1)
	printJSON("alerts:", alerts)

	log.Printf("Get alert by id: %s", id)
	alert, err = workflow.GetOneAlert(id)
	assert.NoError(err)
	assert.EqualValues(string(id), string(alert.ID))
	printJSON("alert", alert)

	log.Printf("Delete alert by id: %s", id)
	err = workflow.DeleteOneAlert(id)
	assert.NoError(err)

	log.Printf("Getting list of alerts:")
	alerts, err = workflow.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 0)
	printJSON("alerts", alerts)

	err = workflow.DeleteOneEventType(eventTypeID)
	assert.NoError(err)
	err = workflow.DeleteOneEvent(eventTypeName, eventID)
	assert.NoError(err)
	err = workflow.DeleteOneTrigger(triggerID)
	assert.NoError(err)
}

//---------------------------------------------------------------------------

func (suite *ServerTester) Test05EventMapping() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow
	var err error

	assertNoData(suite.T(), suite.workflow)
	defer assertNoData(suite.T(), suite.workflow)

	var eventTypeName1 = "Type1"
	var eventTypeName2 = "Type2"

	eventtypeF := func(typ string) Ident {
		log.Printf("Creating event type: %s\n", typ)
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num": elasticsearch.MappingElementTypeInteger,
		}
		//printJSON("mapping", mapping)

		eventType := &EventType{Name: typ, Mapping: mapping}
		printJSON("eventType", eventType)

		eventTypeID, err := workflow.PostOneEventType(eventType)
		assert.NoError(err)
		defer piazza.HTTPDelete("/eventtypes/" + string(eventType.ID))
		printJSON("eventTypeID", eventTypeID)

		eventTypeX, err := workflow.GetOneEventType(eventTypeID)
		assert.NoError(err)

		assert.EqualValues(eventTypeID, eventTypeX.ID)
		// printJSON("eventTypeX", eventTypeX)

		return eventTypeID
	}

	eventF := func(eventTypeID Ident, eventTypeName string, value int) Ident {
		log.Printf("Creating event: %s %s %d\n", eventTypeID, eventTypeName, value)
		event := &Event{
			EventTypeID: eventTypeID,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": value,
			},
		}

		printJSON("event", event)
		eventID, err := workflow.PostOneEvent(eventTypeName, event)
		assert.NoError(err)
		defer piazza.HTTPDelete("/events/" + eventTypeName + "/" + string(eventID))

		printJSON("eventID", eventID)
		eventX, err := workflow.GetOneEvent(eventTypeName, eventID)
		assert.NoError(err)

		assert.EqualValues(eventID, eventX.ID)

		// printJSON("eventX", eventX)
		return eventID
	}

	dump := func(eventTypeName string, expected int) {
		x, err := workflow.GetAllEvents(eventTypeName)
		assert.NoError(err)
		assert.Len(*x, expected)
	}

	et1Id := eventtypeF(eventTypeName1)
	et2Id := eventtypeF(eventTypeName2)

	{
		x, err := workflow.GetAllEventTypes()
		assert.NoError(err)
		assert.Len(*x, 2)
	}

	dump(eventTypeName1, 0)
	dump("", 0)

	{
		x, err := workflow.GetAllEventTypes()
		assert.NoError(err)
		assert.Len(*x, 2)
	}

	{
		x, err := workflow.GetOneEventType(et1Id)
		assert.NoError(err)
		assert.EqualValues(string(et1Id), string((*x).ID))
	}

	e1Id := eventF(et1Id, eventTypeName1, 17)
	dump(eventTypeName1, 1)

	e2Id := eventF(et1Id, eventTypeName1, 18)
	dump(eventTypeName1, 2)

	e3Id := eventF(et2Id, eventTypeName2, 19)
	dump(eventTypeName2, 1)

	dump("", 3)

	err = workflow.DeleteOneEvent(eventTypeName1, e1Id)
	assert.NoError(err)
	err = workflow.DeleteOneEvent(eventTypeName1, e2Id)
	assert.NoError(err)
	err = workflow.DeleteOneEvent(eventTypeName2, e3Id)
	assert.NoError(err)

	err = workflow.DeleteOneEventType(et1Id)
	assert.NoError(err)
	err = workflow.DeleteOneEventType(et2Id)
	assert.NoError(err)
}

func (suite *ServerTester) Test06Workflow() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow
	var err error

	assertNoData(suite.T(), suite.workflow)
	defer assertNoData(suite.T(), suite.workflow)

	var eventTypeName = "EventTypeA"

	var et1ID Ident
	{
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num":    elasticsearch.MappingElementTypeInteger,
			"str":    elasticsearch.MappingElementTypeString,
			"apiKey": elasticsearch.MappingElementTypeString,
			"jobId":  elasticsearch.MappingElementTypeString,
		}

		log.Printf("Creating event type:\n")
		eventType := &EventType{Name: eventTypeName, Mapping: mapping}
		printJSON("event type", eventType)
		et1ID, err = workflow.PostOneEventType(eventType)
		printJSON("event type id", et1ID)
		assert.NoError(err)
		defer func() {
			log.Printf("Deleting event type by id: %s", et1ID)
			err := workflow.DeleteOneEventType(et1ID)
			assert.NoError(err)
		}()
	}

	var t1ID Ident
	{
		log.Printf("Creating trigger:\n")
		trigger := &Trigger{
			Title: "the x1 trigger",
			Condition: Condition{
				EventTypeID: et1ID,
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"num": 17,
						},
					},
				},
			},
			Job: Job{
				//Task: "the x1 task",
				// Using a GetJob call as it is as close to a 'noop' as I could find.
				Task: `{"apiKey": "$apiKey", "jobType": {"type": "get", "jobId": "$jobId"}}`,
			},
		}

		printJSON("trigger", trigger)
		t1ID, err = workflow.PostOneTrigger(trigger)
		assert.NoError(err)
		defer func() {
			log.Printf("Deleting trigger by id: %s\n", t1ID)
			err := workflow.DeleteOneTrigger(t1ID)
			assert.NoError(err)
		}()
		printJSON("trigger id", t1ID)
	}

	var e1ID Ident
	{
		log.Printf("Creating event:\n")
		// will cause trigger TRG1
		event := &Event{
			EventTypeID: et1ID,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num":    17,
				"str":    "quick",
				"apiKey": "my-api-key-38n987",
				"jobId":  "789a6531-85a9-4098-aa3c-e90d07d9b8a3",
			},
		}

		printJSON("event", event)
		e1ID, err = workflow.PostOneEvent(eventTypeName, event)
		assert.NoError(err)
		printJSON("event id", e1ID)
		defer func() {
			log.Printf("Deleting event by id: %s\n", e1ID)
			err := workflow.DeleteOneEvent(eventTypeName, e1ID)
			assert.NoError(err)
		}()
	}

	{
		log.Printf("Creating event:\n")

		// will cause no triggers
		event := &Event{
			EventTypeID: et1ID,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": 18,
				"str": "brown",
				// Probably don't need the following as job shouldn't be executed.
				"apiKey": "my-api-key-38n987",
				"jobId":  "789a6531-85a9-4098-aa3c-e90d07d9b8a3",
			},
		}

		printJSON("event", event)
		e2ID, err := workflow.PostOneEvent(eventTypeName, event)
		assert.NoError(err)
		printJSON("event id", e2ID)

		defer func() {
			log.Printf("Deleting event by id: %s\n", e2ID)
			err := workflow.DeleteOneEvent(eventTypeName, e2ID)
			assert.NoError(err)
		}()
	}

	{
		log.Printf("Getting list of alerts:\n")
		alerts, err := workflow.GetAllAlerts()
		assert.NoError(err)
		assert.Len(*alerts, 1)
		printJSON("alerts", alerts)

		alert0 := (*alerts)[0]
		assert.EqualValues(e1ID, alert0.EventID)
		assert.EqualValues(t1ID, alert0.TriggerID)

		log.Printf("Delete alert by id: %s", alert0.ID)
		err = workflow.DeleteOneAlert(alert0.ID)
		assert.NoError(err)
	}
}

func (suite *ServerTester) Test99Noop() {
	t := suite.T()
	assert := assert.New(t)
	assert.Equal(17, 10+7)
}

func printJSON(msg string, input interface{}) {
	results, err := json.Marshal(input)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("\t%s: %s\n", msg, string(results))
}
