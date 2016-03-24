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
	sys      *piazza.System
	workflow *PzWorkflowService
}

func startServer() *piazza.System {
	config, err := piazza.NewConfig(piazza.PzWorkflow, piazza.ConfigModeTest)
	if err != nil {
		log.Fatal(err)
	}

	sys, err := piazza.NewSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	theLogger, err := loggerPkg.NewMockLoggerService(sys)
	if err != nil {
		log.Fatal(err)
	}
	var tmp = theLogger
	clogger := loggerPkg.NewCustomLogger(&tmp, piazza.PzWorkflow, config.GetAddress())

	theUuidgen, err := uuidgenPkg.NewMockUuidGenService(sys)
	if err != nil {
		log.Fatal(err)
	}

	es, err := elasticsearch.NewClient(sys, true)
	if err != nil {
		log.Fatal(err)
	}
	sys.Services[piazza.PzElasticSearch] = es

	routes, err := CreateHandlers(sys, clogger, theUuidgen)
	if err != nil {
		log.Fatal(err)
	}

	_ = sys.StartServer(routes)

	return sys
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

	sys := startServer()

	workflow, err := NewPzWorkflowService(sys, sys.Config.GetBindToAddress())
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

	typs, err := workflow.GetAllEventTypes()
	assert.Error(err)

	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)
	id, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)

	typs, err = workflow.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*typs, 1)

	typ, err := workflow.GetOneEventType(id)
	assert.NoError(err)
	assert.EqualValues(string(id), string(typ.ID))

	err = workflow.DeleteOneEventType(id)
	assert.NoError(err)

	typs, err = workflow.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*typs, 0)
}

func (suite *ServerTester) xTest02Event() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), suite.workflow)
	defer assertNoData(suite.T(), suite.workflow)

	events, err := workflow.GetAllEvents("")
	assert.NoError(err)
	assert.Len(*events, 0)

	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)

	eventTypeID, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)

	event := makeTestEvent(eventTypeID)
	id, err := workflow.PostOneEvent(eventTypeName, event)
	assert.NoError(err)

	events, err = workflow.GetAllEvents(eventTypeName)
	assert.NoError(err)
	assert.Len(*events, 1)

	event, err = workflow.GetOneEvent(eventTypeName, id)
	assert.NoError(err)
	assert.EqualValues(string(id), string(event.ID))

	err = workflow.DeleteOneEvent(eventTypeName, id)
	assert.NoError(err)

	events, err = workflow.GetAllEvents(eventTypeName)
	assert.NoError(err)
	assert.Len(*events, 0)

	err = workflow.DeleteOneEventType(eventTypeID)
	assert.NoError(err)
}

func (suite *ServerTester) xTest03Trigger() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), suite.workflow)
	defer assertNoData(suite.T(), suite.workflow)

	triggers, err := workflow.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*triggers, 0)

	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)

	eventTypeID, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)

	trigger := makeTestTrigger(eventTypeID)
	id, err := workflow.PostOneTrigger(trigger)

	triggers, err = workflow.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*triggers, 1)

	trigger, err = workflow.GetOneTrigger(id)
	assert.NoError(err)
	assert.EqualValues(string(id), string(trigger.ID))

	err = workflow.DeleteOneTrigger(id)
	assert.NoError(err)

	triggers, err = workflow.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*triggers, 0)

	err = workflow.DeleteOneEventType(eventTypeID)
	assert.NoError(err)
}

func (suite *ServerTester) xTest04Alert() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), suite.workflow)
	defer assertNoData(suite.T(), suite.workflow)

	alerts, err := workflow.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 0)

	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)

	eventTypeID, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)

	trigger := makeTestTrigger(eventTypeID)
	triggerID, err := workflow.PostOneTrigger(trigger)
	assert.NoError(err)

	event := makeTestEvent(eventTypeID)
	eventID, err := workflow.PostOneEvent(eventTypeName, event)
	assert.NoError(err)

	alert := &Alert{
		TriggerID: triggerID,
		EventID:   eventID,
	}
	id, err := workflow.PostOneAlert(alert)
	assert.NoError(err)

	alerts, err = workflow.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 1)

	alert, err = workflow.GetOneAlert(id)
	assert.NoError(err)
	assert.EqualValues(string(id), string(alert.ID))

	err = workflow.DeleteOneAlert(id)
	assert.NoError(err)

	alerts, err = workflow.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 0)

	err = workflow.DeleteOneEventType(eventTypeID)
	assert.NoError(err)
	err = workflow.DeleteOneEvent(eventTypeName, eventID)
	assert.NoError(err)
	err = workflow.DeleteOneTrigger(triggerID)
	assert.NoError(err)
}

//---------------------------------------------------------------------------

func (suite *ServerTester) xTest05EventMapping() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow
	var err error

	assertNoData(suite.T(), suite.workflow)
	defer assertNoData(suite.T(), suite.workflow)

	var eventTypeName1 = "Type1"
	var eventTypeName2 = "Type2"

	eventtypeF := func(typ string) Ident {
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num": elasticsearch.MappingElementTypeInteger,
		}

		eventType := &EventType{Name: typ, Mapping: mapping}

		eventTypeID, err := workflow.PostOneEventType(eventType)
		assert.NoError(err)
		defer piazza.HTTPDelete("/eventtypes/" + string(eventType.ID))

		eventTypeX, err := workflow.GetOneEventType(eventTypeID)
		assert.NoError(err)

		assert.EqualValues(eventTypeID, eventTypeX.ID)

		return eventTypeID
	}

	eventF := func(eventTypeID Ident, eventTypeName string, value int) Ident {
		event := &Event{
			EventTypeID: eventTypeID,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": value,
			},
		}

		eventID, err := workflow.PostOneEvent(eventTypeName, event)
		assert.NoError(err)
		defer piazza.HTTPDelete("/events/" + eventTypeName + "/" + string(eventID))

		eventX, err := workflow.GetOneEvent(eventTypeName, eventID)
		assert.NoError(err)

		assert.EqualValues(eventID, eventX.ID)

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

func (suite *ServerTester) xTest06Workflow() {
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

		eventType := &EventType{Name: eventTypeName, Mapping: mapping}

		et1ID, err = workflow.PostOneEventType(eventType)
		assert.NoError(err)
		defer piazza.HTTPDelete("/eventtypes/" + string(et1ID))
	}

	var t1ID Ident
	{
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

		t1ID, err = workflow.PostOneTrigger(trigger)
		assert.NoError(err)
	}

	var e1ID Ident
	{
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

		e1ID, err = workflow.PostOneEvent(eventTypeName, event)
		assert.NoError(err)
	}

	{
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

		_, err = workflow.PostOneEvent(eventTypeName, event)
		assert.NoError(err)
	}

	{
		alerts, err := workflow.GetAllAlerts()
		assert.NoError(err)
		assert.Len(*alerts, 1)

		alert0 := (*alerts)[0]
		assert.EqualValues(e1ID, alert0.EventID)
		assert.EqualValues(t1ID, alert0.TriggerID)
	}
}

func (suite *ServerTester) xTest99Noop() {
	t := suite.T()
	assert := assert.New(t)
	assert.Equal(17, 10+7)
}
