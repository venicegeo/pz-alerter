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
	loggerPkg "github.com/venicegeo/pz-logger/lib"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
)

const MOCKING = !true

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

	//es, err := workflow.GetAllEvents("")
	//if err == nil {
	//	assert.Len(*es, 0)
	//}

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

	var required []piazza.ServiceName
	if MOCKING {
		required = []piazza.ServiceName{}
	} else {
		required = []piazza.ServiceName{
			piazza.PzElasticSearch,
			piazza.PzLogger,
			piazza.PzGateway,
		}
	}

	sys, err := piazza.NewSystemConfig(piazza.PzWorkflow, required)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := loggerPkg.NewMockClient(sys)
	if err != nil {
		log.Fatal(err)
	}

	uuidgen, err := uuidgenPkg.NewMockUuidGenService(sys)
	if err != nil {
		log.Fatal(err)
	}

	var eventtypesIndex, eventsIndex, triggersIndex, alertsIndex elasticsearch.IIndex
	if MOCKING {
		eventtypesIndex = elasticsearch.NewMockIndex("eventtypes")
		eventsIndex = elasticsearch.NewMockIndex("events")
		triggersIndex = elasticsearch.NewMockIndex("triggers")
		alertsIndex = elasticsearch.NewMockIndex("alerts")
	} else {
		eventtypesIndex, err = elasticsearch.NewIndex(sys, "eventtypes$")
		if err != nil {
			log.Fatal(err)
		}
		eventsIndex, err = elasticsearch.NewIndex(sys, "events$")
		if err != nil {
			log.Fatal(err)
		}
		triggersIndex, err = elasticsearch.NewIndex(sys, "triggers$")
		if err != nil {
			log.Fatal(err)
		}
		alertsIndex, err = elasticsearch.NewIndex(sys, "alerts$")
		if err != nil {
			log.Fatal(err)
		}
	}

	// start server
	{
		routes, err := CreateHandlers(sys, logger, uuidgen,
			eventtypesIndex, eventsIndex, triggersIndex, alertsIndex)
		if err != nil {
			log.Fatal(err)
		}

		_ = sys.StartServer(routes)
	}

	workflow, err := NewPzWorkflowService(sys, logger)
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

func makeTestTrigger(eventTypeIDs []Ident) *Trigger {
	trigger := &Trigger{
		Title: "MY TRIGGER TITLE",
		Condition: Condition{
			EventTypeIDs: eventTypeIDs,
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

	//time.Sleep(2 * time.Second)

	log.Printf("Getting list of events (type=\"\"):")
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

	log.Printf("Getting list of events (type=%s):", eventTypeName)
	events, err = workflow.GetAllEvents(eventTypeName)
	assert.NoError(err)
	assert.Len(*events, 1)
	printJSON("Events", events)

	log.Printf("Getting list of events (type=\"\"):")
	events, err = workflow.GetAllEvents("")
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

	log.Printf("Getting list of events (type=%s):", eventTypeName)
	events, err = workflow.GetAllEvents(eventTypeName)
	assert.NoError(err)
	assert.Len(*events, 0)
	printJSON("Events", events)

	log.Printf("Getting list of events (type=\"\"):")
	events, err = workflow.GetAllEvents("")
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
	trigger := makeTestTrigger([]Ident{eventTypeID})
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
	assert.Len(*alerts, 0)
	printJSON("alerts", alerts)

	log.Printf("Creating new event type:")
	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)
	printJSON("event type", eventType)
	eventTypeID, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)
	printJSON("event type id:", eventTypeID)

	log.Printf("Creating new trigger:")
	trigger := makeTestTrigger([]Ident{eventTypeID})
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
	printJSON("alerts", alerts)

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

		printJSON("eventID", eventID)
		eventX, err := workflow.GetOneEvent(eventTypeName, eventID)
		assert.NoError(err)

		assert.EqualValues(eventID, eventX.ID)

		// printJSON("eventX", eventX)
		return eventID
	}

	dumpEventsF := func(eventTypeName string, expected int) {
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

	{
		// no events yet!
		x, err := workflow.GetAllEvents(eventTypeName1)
		// TODO: this is a bug, mocked and real should both return same answer
		if MOCKING {
			assert.Error(err)
		} else {
			assert.NoError(err)
			assert.Len(*x, 0)
		}

		x, err = workflow.GetAllEvents(eventTypeName2)
		if MOCKING {
			assert.Error(err)
		} else {
			assert.NoError(err)
			assert.Len(*x, 0)
		}
	}

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
	dumpEventsF(eventTypeName1, 1)

	e2Id := eventF(et1Id, eventTypeName1, 18)
	dumpEventsF(eventTypeName1, 2)

	e3Id := eventF(et2Id, eventTypeName2, 19)
	dumpEventsF(eventTypeName2, 1)

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
			"num":      elasticsearch.MappingElementTypeInteger,
			"str":      elasticsearch.MappingElementTypeString,
			"userName": elasticsearch.MappingElementTypeString,
			"jobId":    elasticsearch.MappingElementTypeString,
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
				EventTypeIDs: []Ident{et1ID},
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
				Task: `{"userName": "$userName", "jobType": {"type": "get", "jobId": "$jobId"}}`,
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
				"num":      17,
				"str":      "quick",
				"userName": "my-api-key-38n987",
				"jobId":    "43688858-b6d4-4ef9-a98b-163e1980bba8",
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
				"userName": "my-api-key-38n987",
				"jobId":    "43688858-b6d4-4ef9-a98b-163e1980bba8",
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
		if MOCKING {
			t.Skip("Skipping test, because mocking")
		}
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

func (suite *ServerTester) Test07MultiTrigger() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	var mapping = map[string]elasticsearch.MappingElementTypeName{
		"num":      elasticsearch.MappingElementTypeInteger,
		"str":      elasticsearch.MappingElementTypeString,
		"userName": elasticsearch.MappingElementTypeString,
		"jobId":    elasticsearch.MappingElementTypeString,
	}

	var data = map[string]interface{}{
		"num": 31,
		"str": "brown",
		// Probably don't need the following as job shouldn't be executed.
		"userName": "my-api-key-38n987",
		"jobId":    "43688858-b6d4-4ef9-a98b-163e1980bba8",
	}

	// Create Event Type 1
	log.Printf("\tCreating event type 1:")
	eventType1 := makeTestEventType("Event Type 1")
	eventType1.Mapping = mapping
	printJSON("\tevent type", eventType1)
	eventTypeId1, err := workflow.PostOneEventType(eventType1)
	assert.NoError(err)
	printJSON("\tevent type id", eventTypeId1)

	defer func() {
		log.Printf("\tDeleting event type: %s\n", eventTypeId1)
		err = workflow.DeleteOneEventType(eventTypeId1)
		assert.NoError(err)
	}()

	// Create Event Type 2
	log.Printf("\tCreating event type 2:")
	eventType2 := makeTestEventType("Event Type 2")
	eventType2.Mapping = mapping
	printJSON("\tevent type", eventType2)
	eventTypeId2, err := workflow.PostOneEventType(eventType2)
	assert.NoError(err)
	printJSON("\tevent type id", eventTypeId2)

	defer func() {
		log.Printf("\tDeleting event type: %s\n", eventTypeId2)
		err = workflow.DeleteOneEventType(eventTypeId2)
		assert.NoError(err)
	}()

	// Create MultiTrigger
	log.Printf("\tCreating trigger:")
	trigger := makeTestTrigger([]Ident{eventTypeId1, eventTypeId2})
	trigger.Job = Job{
		//Task: "the x1 task",
		// Using a GetJob call as it is as close to a 'noop' as I could find.
		Task: `{"userName": "$userName", "jobType": {"type": "get", "jobId": "$jobId"}}`,
	}
	printJSON("\ttrigger", trigger)
	triggerId, err := workflow.PostOneTrigger(trigger)
	assert.NoError(err)
	printJSON("\ttrigger id", triggerId)

	defer func() {
		log.Printf("\tDeleting trigger: %s\n", triggerId)
		err = workflow.DeleteOneTrigger(triggerId)
		assert.NoError(err)
	}()

	// Create Event of Type 1
	log.Printf("\tCreating new event:")
	event1 := makeTestEvent(eventTypeId1)
	event1.Data = data
	printJSON("\tevent", event1)
	eventId1, err := workflow.PostOneEvent(eventType1.Name, event1)
	assert.NoError(err)
	printJSON("\tevent id", eventId1)

	defer func() {
		log.Printf("\tDeleting event: %s\n", eventId1)
		err = workflow.DeleteOneEvent(eventType1.Name, eventId1)
		assert.NoError(err)
	}()

	// Create Event of Type 2
	log.Printf("\tCreating new event:")
	event2 := makeTestEvent(eventTypeId2)
	event2.Data = data
	printJSON("\tevent", event2)
	eventId2, err := workflow.PostOneEvent(eventType2.Name, event2)
	assert.NoError(err)
	printJSON("\tevent id", eventId2)

	defer func() {
		log.Printf("\tDeleting event: %s\n", eventId2)
		err = workflow.DeleteOneEvent(eventType2.Name, eventId2)
		assert.NoError(err)
	}()

	{
		if MOCKING {
			t.Skip("Skipping test, because mocking")
		}
		log.Printf("Getting list of alerts:\n")
		alerts, err := workflow.GetAllAlerts()
		assert.NoError(err)
		assert.Len(*alerts, 2)
		printJSON("alerts", alerts)

		// Delete Alert 1
		alert0 := (*alerts)[0]
		assert.EqualValues(eventId1, alert0.EventID)
		assert.EqualValues(triggerId, alert0.TriggerID)
		log.Printf("Delete alert by id: %s", alert0.ID)
		err = workflow.DeleteOneAlert(alert0.ID)
		assert.NoError(err)

		// Delete Alert 2
		alert1 := (*alerts)[1]
		assert.EqualValues(eventId2, alert1.EventID)
		assert.EqualValues(triggerId, alert1.TriggerID)
		log.Printf("Delete alert by id: %s", alert1.ID)
		err = workflow.DeleteOneAlert(alert1.ID)
		assert.NoError(err)

	}
}

func printJSON(msg string, input interface{}) {
	if input != nil {
		results, err := json.Marshal(input)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("\t%s: %s\n", msg, string(results))
	} else {
		log.Printf("\t%s: null\n", msg)
	}
}
