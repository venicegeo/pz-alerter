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
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	pzlogger "github.com/venicegeo/pz-logger/logger"
	pzuuidgen "github.com/venicegeo/pz-uuidgen/uuidgen"
)

type ServerTester struct {
	suite.Suite
	sys    *piazza.SystemConfig
	client *Client
}

func assertNoData(t *testing.T, client *Client) {
	assert := assert.New(t)

	var err error

	ts, err := client.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*ts, 2)

	//es, err := workflow.GetAllEvents("")
	//if err == nil {
	//	assert.Len(*es, 0)
	//}

	as, err := client.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*as, 0)

	xs, err := client.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*xs, 0)
}

func TestRunSuite(t *testing.T) {

	var required []piazza.ServiceName
	required = []piazza.ServiceName{}

	sys, err := piazza.NewSystemConfig(piazza.PzWorkflow, required)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := pzlogger.NewMockClient()
	if err != nil {
		log.Fatal(err)
	}

	var uuidgen pzuuidgen.IClient

	uuidgen, err = pzuuidgen.NewMockClient()
	if err != nil {
		log.Fatal(err)
	}

	var eventtypesIndex, eventsIndex, triggersIndex, alertsIndex, cronIndex elasticsearch.IIndex
	eventtypesIndex = elasticsearch.NewMockIndex("eventtypes")
	eventsIndex = elasticsearch.NewMockIndex("events")
	triggersIndex = elasticsearch.NewMockIndex("triggers")
	alertsIndex = elasticsearch.NewMockIndex("alerts")
	cronIndex = elasticsearch.NewMockIndex("crons")

	workflowService := &WorkflowService{}
	err = workflowService.Init(sys, logger, uuidgen, eventtypesIndex, eventsIndex, triggersIndex, alertsIndex, cronIndex)
	if err != nil {
		log.Fatal(err)
	}
	workflowServer := &WorkflowServer{}
	err = workflowServer.Init(workflowService)
	if err != nil {
		log.Fatal(err)
	}

	genericServer := piazza.GenericServer{Sys: sys}

	err = genericServer.Configure(workflowServer.Routes)
	if err != nil {
		log.Fatal(err)
	}

	_, err = genericServer.Start()
	if err != nil {
		log.Fatal(err)
	}

	client, err := NewClient(sys, logger)
	if err != nil {
		log.Fatal(err)
	}

	mappingTester := &mappingTester{}
	suite.Run(t, mappingTester)

	serverTester := &ServerTester{client: client, sys: sys}
	suite.Run(t, serverTester)

	clientTester := &ClientTester{client: client, sys: sys}
	suite.Run(t, clientTester)

	err = eventtypesIndex.Delete()
	if err != nil {
		log.Fatal(err)
	}

	err = eventsIndex.Delete()
	if err != nil {
		log.Fatal(err)
	}

	err = triggersIndex.Delete()
	if err != nil {
		log.Fatal(err)
	}

	err = alertsIndex.Delete()
	if err != nil {
		log.Fatal(err)
	}

}

//---------------------------------------------------------------------------

// Generate random names
func makeTestEventTypeName() string {
	return "MYTYPE" + strconv.Itoa(rand.Int())
}

func makeTestEventType(eventTypeName string) *EventType {
	mapping := map[string]interface{}{
		"num": elasticsearch.MappingElementTypeInteger,
	}
	return &EventType{Name: eventTypeName, Mapping: mapping}
}

func makeTestEvent(eventTypeID piazza.Ident) *Event {
	event := &Event{
		EventTypeId: eventTypeID,
		CreatedOn:   time.Now(),
		Data: map[string]interface{}{
			"num": 17,
		},
	}
	return event
}

func makeTestTrigger(eventTypeIDs []piazza.Ident) *Trigger {
	trigger := &Trigger{
		Name:    "MY TRIGGER TITLE",
		Enabled: true,
		Condition: Condition{
			EventTypeIds: eventTypeIDs,
			Query: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"num": 31,
					},
				},
			},
		},
		Job: JobRequest{
			CreatedBy: "test",
			JobType: JobType{
				Type: "execute-service",
				Data: map[string]interface{}{
					// "dataInputs": map[string]interface{},
					// "dataOutput": map[string]interface{},
					"serviceId": "ddd5134",
				},
			},
		},
	}
	return trigger
}

//---------------------------------------------------------------------------
func (suite *ServerTester) Test00Version() {
t:
	suite.T()
	assert := assert.New(t)

	client := suite.client

	version, err := client.GetVersion()
	assert.NoError(err)
	assert.EqualValues("1.0.0", version.Version)
}

func (suite *ServerTester) Test01EventType() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), suite.client)
	defer assertNoData(suite.T(), suite.client)

	//log.Printf("Getting list of event types:")
	typs, err := client.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*typs, 2)
	//printJSON("EventTypes", typs)

	//log.Printf("Creating new Event Type:")
	eventTypeName := makeTestEventTypeName()
	//printJSON("event type name", eventTypeName)
	eventType := makeTestEventType(eventTypeName)
	//printJSON("event type", eventType)

	respEventType, err := client.PostEventType(eventType)
	id := respEventType.EventTypeId
	assert.NoError(err)
	//log.Printf("New event: %#v", respEventType)

	//log.Printf("Getting list of event types:")
	typs, err = client.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*typs, 3)

	//printJSON("EventTypes", typs)

	//log.Printf("Getting event type by Id: %s", id)
	respTyp, err := client.GetEventType(id)
	assert.NoError(err)
	assert.EqualValues(string(id), string(respTyp.EventTypeId))

	//printJSON("Got Event type", typ)
	//log.Printf("Deleting Event type by Id: %s", id)

	err = client.DeleteEventType(id)
	assert.NoError(err)

	//log.Printf("Getting list of event types:")
	typs, err = client.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*typs, 2)

	//printJSON("EventTypes", typs)
}

func (suite *ServerTester) Test02Event() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), suite.client)
	defer assertNoData(suite.T(), suite.client)

	//log.Printf("Getting list of events (type=\"\"):")
	events, err := client.GetAllEvents()
	assert.NoError(err)
	assert.Len(*events, 0)
	//printJSON("Events", events)

	//log.Printf("Creating new event type:")
	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)
	//printJSON("event type", eventType)
	//log.Printf("CCC %#v", eventType)
	respEventType, err := client.PostEventType(eventType)
	//log.Printf("BBB %#v", respEventType)
	eventTypeID := respEventType.EventTypeId
	assert.NoError(err)
	//printJSON("event type id", eventTypeID)

	//log.Printf("Creating new event:")
	event := makeTestEvent(eventTypeID)
	respEvent, err := client.PostEvent(event)
	//log.Printf("CCC %#v", event)
	id := respEvent.EventId
	assert.NoError(err)
	//printJSON("event id", id)

	//log.Printf("Getting list of events (type=%s):", eventTypeID)
	events, err = client.GetAllEventsByEventType(eventTypeID)
	assert.NoError(err)
	assert.Len(*events, 1)
	//printJSON("Events", events)

	//log.Printf("Getting list of events (type=\"\"):")
	events, err = client.GetAllEvents()
	assert.NoError(err)
	assert.Len(*events, 1)
	//printJSON("Events", events)

	//log.Printf("Getting event by id: %s", id)
	event, err = client.GetEvent(id)
	//printJSON("Got event", event)
	assert.NoError(err)
	assert.EqualValues(string(id), string(event.EventId))

	//log.Printf("Deleting event by id: %s", id)
	err = client.DeleteEvent(id)
	assert.NoError(err)

	//log.Printf("Getting list of events (type=%s):", eventTypeName)
	events, err = client.GetAllEventsByEventType(eventTypeID)
	assert.NoError(err)
	assert.Len(*events, 0)
	//printJSON("Events", events)

	//log.Printf("Getting list of events (type=\"\"):")
	events, err = client.GetAllEvents()
	assert.NoError(err)
	assert.Len(*events, 0)
	//printJSON("Events", events)

	//log.Printf("Deleting event type by id: %s", eventTypeID)
	err = client.DeleteEventType(eventTypeID)
	assert.NoError(err)
}

func (suite *ServerTester) Test03Trigger() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), suite.client)
	defer assertNoData(suite.T(), suite.client)

	//log.Printf("Getting list of triggers:")
	triggers, err := client.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*triggers, 0)

	//printJSON("triggers", triggers)

	//log.Printf("Creating new event type:")
	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)
	//printJSON("event type", eventType)
	respEventType, err := client.PostEventType(eventType)
	eventTypeID := respEventType.EventTypeId
	assert.NoError(err)
	//printJSON("event type id", eventTypeID)

	//log.Printf("Creating new trigger:")
	trigger := makeTestTrigger([]piazza.Ident{eventTypeID})
	//printJSON("trigger", trigger)
	respTrigger, err := client.PostTrigger(trigger)
	id := respTrigger.TriggerId
	//printJSON("trigger id", id)

	//log.Printf("Getting list of triggers:")
	triggers, err = client.GetAllTriggers()
	assert.NoError(err)
	//printJSON("triggers", triggers)

	//log.Printf("Getting trigger by id: %s", id)
	trigger, err = client.GetTrigger(id)
	assert.NoError(err)
	assert.EqualValues(string(id), string(trigger.TriggerId))
	//printJSON("Trigger", trigger)

	//log.Printf("Delete trigger by id: %s", id)
	err = client.DeleteTrigger(id)
	assert.NoError(err)

	//log.Printf("Getting list of triggers:")
	triggers, err = client.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*triggers, 0)
	//printJSON("triggers", triggers)

	//log.Printf("Delete event type by id: %s", eventTypeID)
	err = client.DeleteEventType(eventTypeID)
	assert.NoError(err)
}

func (suite *ServerTester) Test04Alert() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), suite.client)
	defer assertNoData(suite.T(), suite.client)

	//log.Printf("Getting list of alerts:")
	alerts, err := client.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 0)
	//printJSON("alerts", alerts)

	//log.Printf("Creating new event type:")
	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)
	//printJSON("event type", eventType)
	respEventType, err := client.PostEventType(eventType)
	eventTypeID := respEventType.EventTypeId
	assert.NoError(err)
	//printJSON("event type id:", eventTypeID)

	//log.Printf("Creating new trigger:")
	trigger := makeTestTrigger([]piazza.Ident{eventTypeID})
	//printJSON("Trigger", trigger)
	respTrigger, err := client.PostTrigger(trigger)
	triggerID := respTrigger.TriggerId
	assert.NoError(err)
	//printJSON("Trigger ID", triggerID)

	sleep()

	//log.Printf("Creating new event:")
	event := makeTestEvent(eventTypeID)
	//printJSON("event", event)
	respPostEvent, err := client.PostEvent(event)
	eventID := respPostEvent.EventId
	assert.NoError(err)
	//printJSON("eventID", eventID)

	//log.Printf("Creating new alert:")
	alert := &Alert{
		TriggerId: triggerID,
		EventId:   eventID,
	}
	//printJSON("alert", alert)
	respAlert, err := client.PostAlert(alert)
	id := respAlert.AlertId
	assert.NoError(err)
	//printJSON("alert id", id)

	//log.Printf("Getting list of alerts:")
	alerts, err = client.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 1)
	//printJSON("alerts", alerts)

	//log.Printf("Get alert by id: %s", id)
	alert, err = client.GetAlert(id)
	assert.NoError(err)
	assert.EqualValues(string(id), string(alert.AlertId))
	//printJSON("alert", alert)

	//log.Printf("Delete alert by id: %s", id)
	err = client.DeleteAlert(id)
	assert.NoError(err)

	//log.Printf("Getting list of alerts:")
	alerts, err = client.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 0)
	//printJSON("alerts", alerts)

	err = client.DeleteEventType(eventTypeID)
	assert.NoError(err)
	err = client.DeleteEvent(eventID)
	assert.NoError(err)
	err = client.DeleteTrigger(triggerID)
	assert.NoError(err)
}

//---------------------------------------------------------------------------

func (suite *ServerTester) Test05EventMapping() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client
	var err error

	assertNoData(suite.T(), suite.client)
	defer assertNoData(suite.T(), suite.client)

	var eventTypeName1 = "Type1"
	var eventTypeName2 = "Type2"

	createEventType := func(typ string) piazza.Ident {
		//log.Printf("Creating event type: %s\n", typ)
		mapping := map[string]interface{}{
			"num": elasticsearch.MappingElementTypeInteger,
		}
		//printJSON("mapping", mapping)

		eventType := &EventType{Name: typ, Mapping: mapping}
		//printJSON("eventType", eventType)

		respEventType, err := client.PostEventType(eventType)
		eventTypeID := respEventType.EventTypeId
		assert.NoError(err)
		//printJSON("eventTypeID", eventTypeID)

		eventTypeX, err := client.GetEventType(eventTypeID)
		assert.NoError(err)

		assert.EqualValues(eventTypeID, eventTypeX.EventTypeId)
		// printJSON("eventTypeX", eventTypeX)

		return eventTypeID
	}

	createEvent := func(eventTypeID piazza.Ident, eventTypeName string, value int) piazza.Ident {
		//log.Printf("Creating event: %s %s %d\n", eventTypeID, eventTypeName, value)
		event := &Event{
			EventTypeId: eventTypeID,
			CreatedOn:   time.Now(),
			Data: map[string]interface{}{
				"num": value,
			},
		}

		//printJSON("event", event)
		respEvent, err := client.PostEvent(event)
		eventID := respEvent.EventId
		assert.NoError(err)

		//printJSON("eventID", eventID)
		eventX, err := client.GetEvent(eventID)
		assert.NoError(err)

		assert.EqualValues(eventID, eventX.EventId)

		// printJSON("eventX", eventX)
		return eventID
	}

	checkEvents := func(eventTypeId piazza.Ident, expected int) {
		x, err := client.GetAllEventsByEventType(eventTypeId)
		assert.NoError(err)
		assert.Len(*x, expected)
	}

	et1Id := createEventType(eventTypeName1)
	et2Id := createEventType(eventTypeName2)

	{
		x, err := client.GetAllEventTypes()
		assert.NoError(err)
		assert.Len(*x, 4)
	}

	{
		// no events yet!
		x, err := client.GetAllEvents()
		assert.NoError(err)
		assert.Len(*x, 0)

		// We expect errors here because searching for an EventType that doesn't exist
		// results in an error
		x, err = client.GetAllEventsByEventType(et1Id)
		assert.Error(err)
		assert.Len(*x, 0)
		x, err = client.GetAllEventsByEventType(et2Id)
		assert.Error(err)
		assert.Len(*x, 0)
	}

	{
		x, err := client.GetEventType(et1Id)
		assert.NoError(err)
		assert.EqualValues(string(et1Id), string((*x).EventTypeId))
	}

	e1Id := createEvent(et1Id, eventTypeName1, 17)
	checkEvents(et1Id, 1)

	e2Id := createEvent(et1Id, eventTypeName1, 18)
	checkEvents(et1Id, 2)

	e3Id := createEvent(et2Id, eventTypeName2, 19)
	checkEvents(et2Id, 1)

	err = client.DeleteEvent(e1Id)
	assert.NoError(err)
	err = client.DeleteEvent(e2Id)
	assert.NoError(err)
	err = client.DeleteEvent(e3Id)
	assert.NoError(err)

	err = client.DeleteEventType(et1Id)
	assert.NoError(err)
	err = client.DeleteEventType(et2Id)
	assert.NoError(err)
}

func (suite *ServerTester) Test06Workflow() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), suite.client)
	defer assertNoData(suite.T(), suite.client)

	eventTypeName := makeTestEventTypeName()

	var et1ID piazza.Ident
	{
		mapping := map[string]interface{}{
			"num":      elasticsearch.MappingElementTypeInteger,
			"str":      elasticsearch.MappingElementTypeString,
			"userName": elasticsearch.MappingElementTypeString,
			"jobId":    elasticsearch.MappingElementTypeString,
		}

		//log.Printf("Creating event type:\n")
		eventType := &EventType{Name: eventTypeName, Mapping: mapping}
		//printJSON("event type", eventType)
		respEventType, err := client.PostEventType(eventType)
		et1ID = respEventType.EventTypeId
		//printJSON("event type id", et1ID)
		assert.NoError(err)
		defer func() {
			//log.Printf("Deleting event type by id: %s", et1ID)
			err := client.DeleteEventType(et1ID)
			assert.NoError(err)
		}()
	}

	{
		//log.Printf("Creating trigger:\n")
		trigger := &Trigger{
			Name: "the x1 trigger",
			Condition: Condition{
				EventTypeIds: []piazza.Ident{et1ID},
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"num": 17,
						},
					},
				},
			},
			Job: JobRequest{
				CreatedBy: "test",
				JobType: JobType{
					Type: "execute-service",
					Data: map[string]interface{}{
						// "dataInputs": map[string]interface{},
						// "dataOutput": map[string]interface{},
						"serviceId": "ddd5134",
					},
				},
			},
		}

		//printJSON("trigger", trigger)
		respTrigger, err := client.PostTrigger(trigger)
		t1ID := respTrigger.TriggerId
		assert.NoError(err)
		defer func() {
			//log.Printf("Deleting trigger by id: %s\n", t1ID)
			err := client.DeleteTrigger(t1ID)
			assert.NoError(err)
		}()
		//printJSON("trigger id", t1ID)
	}

	{
		//log.Printf("Creating event:\n")
		// will cause trigger TRG1
		event := &Event{
			EventTypeId: et1ID,
			CreatedOn:   time.Now(),
			Data: map[string]interface{}{
				"num":      17,
				"str":      "quick",
				"userName": "my-api-key-38n987",
				"jobId":    "43688858-b6d4-4ef9-a98b-163e1980bba8",
			},
		}

		//printJSON("event", event)
		respEvent, err := client.PostEvent(event)
		e1ID := respEvent.EventId
		assert.NoError(err)
		//printJSON("event id", e1ID)
		defer func() {
			//log.Printf("Deleting event by id: %s\n", e1ID)
			err := client.DeleteEvent(e1ID)
			assert.NoError(err)
		}()
	}

	{
		//log.Printf("Creating event:\n")

		// will cause no triggers
		event := &Event{
			EventTypeId: et1ID,
			CreatedOn:   time.Now(),
			Data: map[string]interface{}{
				"num": 18,
				"str": "brown",
				// Probably don't need the following as job shouldn't be executed.
				"userName": "my-api-key-38n987",
				"jobId":    "43688858-b6d4-4ef9-a98b-163e1980bba8",
			},
		}

		//printJSON("event", event)
		respEvent2, err := client.PostEvent(event)
		e2ID := respEvent2.EventId
		assert.NoError(err)
		//printJSON("event id", e2ID)

		defer func() {
			//log.Printf("Deleting event by id: %s\n", e2ID)
			err := client.DeleteEvent(e2ID)
			assert.NoError(err)
		}()
	}
}

func (suite *ServerTester) Test07MultiTrigger() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	var mapping = map[string]interface{}{
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
	//log.Printf("\tCreating event type 1:")
	eventType1 := makeTestEventType("Event Type 1")
	eventType1.Mapping = mapping
	//printJSON("\tevent type", eventType1)
	respEventType1, err := client.PostEventType(eventType1)
	eventTypeId1 := respEventType1.EventTypeId
	assert.NoError(err)
	//printJSON("\tevent type id", eventTypeId1)

	defer func() {
		//log.Printf("\tDeleting event type: %s\n", eventTypeId1)
		err = client.DeleteEventType(eventTypeId1)
		assert.NoError(err)
	}()

	// Create Event Type 2
	//log.Printf("\tCreating event type 2:")
	eventType2 := makeTestEventType("Event Type 2")
	eventType2.Mapping = mapping
	//printJSON("\tevent type", eventType2)
	respEventType2, err := client.PostEventType(eventType2)
	eventTypeId2 := respEventType2.EventTypeId
	assert.NoError(err)
	//printJSON("\tevent type id", eventTypeId2)

	defer func() {
		//log.Printf("\tDeleting event type: %s\n", eventTypeId2)
		err = client.DeleteEventType(eventTypeId2)
		assert.NoError(err)
	}()

	// Create MultiTrigger
	//log.Printf("\tCreating trigger:")
	trigger := makeTestTrigger([]piazza.Ident{eventTypeId1, eventTypeId2})
	//printJSON("\ttrigger", trigger)
	respTrigger, err := client.PostTrigger(trigger)
	triggerId := respTrigger.TriggerId
	assert.NoError(err)
	//printJSON("\ttrigger id", triggerId)

	defer func() {
		//log.Printf("\tDeleting trigger: %s\n", triggerId)
		err = client.DeleteTrigger(triggerId)
		assert.NoError(err)
	}()

	// Create Event of Type 1
	//log.Printf("\tCreating new event:")
	event1 := makeTestEvent(eventTypeId1)
	event1.Data = data
	//printJSON("\tevent", event1)
	respEvent1, err := client.PostEvent(event1)
	eventId1 := respEvent1.EventId
	assert.NoError(err)
	//printJSON("\tevent id", eventId1)

	defer func() {
		//log.Printf("\tDeleting event: %s\n", eventId1)
		err = client.DeleteEvent(eventId1)
		assert.NoError(err)
	}()

	// Create Event of Type 2
	//log.Printf("\tCreating new event:")
	event2 := makeTestEvent(eventTypeId2)
	event2.Data = data
	//printJSON("\tevent", event2)
	respEvent2, err := client.PostEvent(event2)
	eventId2 := respEvent2.EventId
	assert.NoError(err)
	//printJSON("\tevent id", eventId2)

	defer func() {
		//log.Printf("\tDeleting event: %s\n", eventId2)
		err = client.DeleteEvent(eventId2)
		assert.NoError(err)
	}()
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
