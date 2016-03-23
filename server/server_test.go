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
	var tmp loggerPkg.ILoggerService = theLogger
	clogger := loggerPkg.NewCustomLogger(&tmp, piazza.PzWorkflow, config.GetAddress())

	theUuidgen, err := uuidgenPkg.NewMockUuidGenService(sys)
	if err != nil {
		log.Fatal(err)
	}

	es, err := elasticsearch.NewElasticsearchClient(sys, true)
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
	assert.NoError(err)
	assert.Len(*ts, 0)

	es, err := workflow.GetAllEvents("")
	assert.NoError(err)
	assert.Len(*es, 0)

	as, err := workflow.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*as, 0)

	xs, err := workflow.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*xs, 0)
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

func makeTestEvent(eventTypeId Ident) *Event {
	event := &Event{
		EventTypeId: eventTypeId,
		Date:        time.Now(),
		Data: map[string]interface{}{
			"num": 17,
		},
	}
	return event
}

func makeTestTrigger(eventTypeId Ident) *Trigger {
	trigger := &Trigger{
		Title: "MY TRIGGER TITLE",
		Condition: Condition{
			EventId: eventTypeId,
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

func (suite *ServerTester) Test_01_EventType() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), suite.workflow)
	defer assertNoData(suite.T(), suite.workflow)

	typs, err := workflow.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*typs, 0)

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

func (suite *ServerTester) xTest_02_Event() {
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

	eventTypeId, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)

	event := makeTestEvent(eventTypeId)
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

	err = workflow.DeleteOneEventType(eventTypeId)
	assert.NoError(err)
}

func (suite *ServerTester) xTest_03_Trigger() {
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

	eventTypeId, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)

	trigger := makeTestTrigger(eventTypeId)
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

	err = workflow.DeleteOneEventType(eventTypeId)
	assert.NoError(err)
}

func (suite *ServerTester) xTest_04_Alert() {
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

	eventTypeId, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)

	trigger := makeTestTrigger(eventTypeId)
	triggerId, err := workflow.PostOneTrigger(trigger)
	assert.NoError(err)

	event := makeTestEvent(eventTypeId)
	eventId, err := workflow.PostOneEvent(eventTypeName, event)
	assert.NoError(err)

	alert := &Alert{
		TriggerId: triggerId,
		EventId:   eventId,
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

	err = workflow.DeleteOneEventType(eventTypeId)
	assert.NoError(err)
	err = workflow.DeleteOneEvent(eventTypeName, eventId)
	assert.NoError(err)
	err = workflow.DeleteOneTrigger(triggerId)
	assert.NoError(err)
}

//---------------------------------------------------------------------------

func (suite *ServerTester) xTest_05_EventMapping() {
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

		eventTypeId, err := workflow.PostOneEventType(eventType)
		assert.NoError(err)
		defer piazza.HTTPDelete("/eventtypes/" + string(eventType.ID))

		eventTypeX, err := workflow.GetOneEventType(eventTypeId)
		assert.NoError(err)

		assert.EqualValues(eventTypeId, eventTypeX.ID)

		return eventTypeId
	}

	eventF := func(eventTypeId Ident, eventTypeName string, value int) Ident {
		event := &Event{
			EventTypeId: eventTypeId,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": value,
			},
		}

		eventId, err := workflow.PostOneEvent(eventTypeName, event)
		assert.NoError(err)
		defer piazza.HTTPDelete("/events/" + eventTypeName + "/" + string(eventId))

		eventX, err := workflow.GetOneEvent(eventTypeName, eventId)
		assert.NoError(err)

		assert.EqualValues(eventId, eventX.ID)

		return eventId
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

func (suite *ServerTester) xTest_06_Workflow() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow
	var err error

	assertNoData(suite.T(), suite.workflow)
	defer assertNoData(suite.T(), suite.workflow)

	var eventTypeName = "EventTypeA"

	var et1Id Ident
	{
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num":    elasticsearch.MappingElementTypeInteger,
			"str":    elasticsearch.MappingElementTypeString,
			"apiKey": elasticsearch.MappingElementTypeString,
			"jobId":  elasticsearch.MappingElementTypeString,
		}

		eventType := &EventType{Name: eventTypeName, Mapping: mapping}

		et1Id, err = workflow.PostOneEventType(eventType)
		assert.NoError(err)
		defer piazza.HTTPDelete("/eventtypes/" + string(et1Id))
	}

	var t1Id Ident
	{
		trigger := &Trigger{
			Title: "the x1 trigger",
			Condition: Condition{
				EventId: et1Id,
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

		t1Id, err = workflow.PostOneTrigger(trigger)
		assert.NoError(err)
	}

	var e1Id Ident
	{
		// will cause trigger TRG1
		event := &Event{
			EventTypeId: et1Id,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num":    17,
				"str":    "quick",
				"apiKey": "my-api-key-38n987",
				"jobId":  "789a6531-85a9-4098-aa3c-e90d07d9b8a3",
			},
		}

		e1Id, err = workflow.PostOneEvent(eventTypeName, event)
		assert.NoError(err)
	}

	{
		// will cause no triggers
		event := &Event{
			EventTypeId: et1Id,
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
		assert.EqualValues(e1Id, alert0.EventId)
		assert.EqualValues(t1Id, alert0.TriggerId)
	}
}

func (suite *ServerTester) xTest_99_Noop() {
	t := suite.T()
	assert := assert.New(t)
	assert.Equal(17, 10+7)
}
