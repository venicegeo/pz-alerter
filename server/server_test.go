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
	sys *piazza.System
	//url string
	workflow *PzWorkflowService
}

func (suite *ServerTester) SetupSuite() {
	t := suite.T()
	assert := assert.New(t)

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

	suite.workflow, err = NewPzWorkflowService(sys, sys.Config.GetBindToAddress())
	if err != nil {
		log.Fatal(err)
	}

	suite.sys = sys

	//suite.url = fmt.Sprintf("http://%s/v1", sys.Config.GetBindToAddress())

	assert.Len(sys.Services, 5)
}

func (suite *ServerTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func TestRunSuite(t *testing.T) {
	s := new(ServerTester)
	suite.Run(t, s)
	c := new(ClientTester)
	suite.Run(t, c)
}

//---------------------------------------------------------------------------

func (suite *ServerTester) getAllEventTypes() *[]EventType {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	eventTypes, err := workflow.GetAllEventTypes()
	assert.NoError(err)

	return eventTypes
}

func (suite *ServerTester) postOneEventType(eventType *EventType) Ident {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	id, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)

	return id
}

func (suite *ServerTester) getOneEventType(id Ident) *EventType {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	eventType, err := workflow.GetOneEventType(id)
	assert.NoError(err)

	return eventType
}

func (suite *ServerTester) deleteOneEventType(id Ident) {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	err := workflow.DeleteOneEventType(id)
	assert.NoError(err)
}

//---------------------------------------------------------------------------

func (suite *ServerTester) getAllEvents(eventTypeName string) *[]Event {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	events, err := workflow.GetAllEvents(eventTypeName)
	assert.NoError(err)

	return events
}

func (suite *ServerTester) postOneEvent(eventTypeName string, event *Event) Ident {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	id, err := workflow.PostOneEvent(eventTypeName, event)
	assert.NoError(err)

	return id
}

func (suite *ServerTester) getOneEvent(eventTypeName string, id Ident) *Event {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	event, err := workflow.GetOneEvent(eventTypeName, id)
	assert.NoError(err)

	return event
}

func (suite *ServerTester) deleteOneEvent(eventTypeName string, id Ident) {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	err := workflow.DeleteOneEvent(eventTypeName, id)
	assert.NoError(err)
}

//---------------------------------------------------------------------------

func (suite *ServerTester) getAllTriggers() *[]Trigger {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	triggers, err := workflow.GetAllTriggers()
	assert.NoError(err)

	return triggers
}

func (suite *ServerTester) postOneTrigger(trigger *Trigger) Ident {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	id, err := workflow.PostOneTrigger(trigger)
	assert.NoError(err)

	return id
}

func (suite *ServerTester) getOneTrigger(id Ident) *Trigger {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	trigger, err := workflow.GetOneTrigger(id)
	assert.NoError(err)

	return trigger
}

func (suite *ServerTester) deleteOneTrigger(id Ident) {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	err := workflow.DeleteOneTrigger(id)
	assert.NoError(err)
}

//---------------------------------------------------------------------------

func (suite *ServerTester) getAllAlerts() *[]Alert {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	alerts, err := workflow.GetAllAlerts()
	assert.NoError(err)

	return alerts
}

func (suite *ServerTester) postOneAlert(eventId, triggerId Ident) Ident {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	alert := &Alert{
		TriggerId: triggerId,
		EventId:   eventId,
	}

	id, err := workflow.PostOneAlert(alert)
	assert.NoError(err)

	return id
}

func (suite *ServerTester) getOneAlert(id Ident) *Alert {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	alert, err := workflow.GetOneAlert(id)
	assert.NoError(err)

	return alert
}

func (suite *ServerTester) deleteOneAlert(id Ident) {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	err := workflow.DeleteOneAlert(id)
	assert.NoError(err)
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

	typs := suite.getAllEventTypes()
	assert.Len(*typs, 0)

	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)
	id := suite.postOneEventType(eventType)

	typs = suite.getAllEventTypes()
	assert.Len(*typs, 1)

	typ := suite.getOneEventType(id)
	assert.EqualValues(string(id), string(typ.ID))

	suite.deleteOneEventType(id)

	typs = suite.getAllEventTypes()
	assert.Len(*typs, 0)
}

func (suite *ServerTester) Test_02_Event() {
	t := suite.T()
	assert := assert.New(t)

	events := suite.getAllEvents("")
	assert.Len(*events, 0)

	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)

	eventTypeId := suite.postOneEventType(eventType)

	event := makeTestEvent(eventTypeId)
	id := suite.postOneEvent(eventTypeName, event)

	events = suite.getAllEvents(eventTypeName)
	assert.Len(*events, 1)

	event = suite.getOneEvent(eventTypeName, id)
	assert.EqualValues(string(id), string(event.ID))

	suite.deleteOneEvent(eventTypeName, id)

	events = suite.getAllEvents(eventTypeName)
	assert.Len(*events, 0)

	suite.deleteOneEventType(eventTypeId)
}

func (suite *ServerTester) Test_03_Trigger() {
	t := suite.T()
	assert := assert.New(t)

	triggers := suite.getAllTriggers()
	assert.Len(*triggers, 0)

	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)

	eventTypeId := suite.postOneEventType(eventType)

	trigger := makeTestTrigger(eventTypeId)
	id := suite.postOneTrigger(trigger)

	triggers = suite.getAllTriggers()
	assert.Len(*triggers, 1)

	trigger = suite.getOneTrigger(id)
	assert.EqualValues(string(id), string(trigger.ID))

	suite.deleteOneTrigger(id)

	triggers = suite.getAllTriggers()
	assert.Len(*triggers, 0)

	suite.deleteOneEventType(eventTypeId)
}

func (suite *ServerTester) Test_04_Alert() {
	t := suite.T()
	assert := assert.New(t)

	alerts := suite.getAllAlerts()
	assert.Len(*alerts, 0)

	eventTypeName := makeTestEventTypeName()
	eventType := makeTestEventType(eventTypeName)

	eventTypeId := suite.postOneEventType(eventType)

	trigger := makeTestTrigger(eventTypeId)
	triggerId := suite.postOneTrigger(trigger)
	event := makeTestEvent(eventTypeId)
	eventId := suite.postOneEvent(eventTypeName, event)

	id := suite.postOneAlert(eventId, triggerId)

	alerts = suite.getAllAlerts()
	assert.Len(*alerts, 1)

	alert := suite.getOneAlert(id)
	assert.EqualValues(string(id), string(alert.ID))

	suite.deleteOneAlert(id)

	alerts = suite.getAllAlerts()
	assert.Len(*alerts, 0)

	suite.deleteOneEventType(eventTypeId)
	suite.deleteOneEvent(eventTypeName, eventId)
	suite.deleteOneTrigger(triggerId)
}

//---------------------------------------------------------------------------

func (suite *ServerTester) Test_06_Workflow() {

	t := suite.T()
	assert := assert.New(t)

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

		et1Id = suite.postOneEventType(eventType)
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

		t1Id = suite.postOneTrigger(trigger)
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

		e1Id = suite.postOneEvent(eventTypeName, event)
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

		_ = suite.postOneEvent(eventTypeName, event)
	}

	{
		alerts := suite.getAllAlerts()
		assert.Len(*alerts, 1)

		alert0 := (*alerts)[0]
		assert.EqualValues(e1Id, alert0.EventId)
		assert.EqualValues(t1Id, alert0.TriggerId)
	}
}

func (suite *ServerTester) Test_05_EventMapping() {

	t := suite.T()
	assert := assert.New(t)

	var eventTypeName1 = "Type1"
	var eventTypeName2 = "Type2"

	eventtypeF := func(typ string) Ident {
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num": elasticsearch.MappingElementTypeInteger,
		}

		eventType := &EventType{Name: typ, Mapping: mapping}

		eventTypeId := suite.postOneEventType(eventType)
		defer piazza.HTTPDelete("/eventtypes/" + string(eventType.ID))

		eventTypeX := suite.getOneEventType(eventTypeId)

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

		eventId := suite.postOneEvent(eventTypeName, event)
		defer piazza.HTTPDelete("/events/" + eventTypeName + "/" + string(eventId))

		eventX := suite.getOneEvent(eventTypeName, eventId)

		assert.EqualValues(eventId, eventX.ID)

		return eventId
	}

	dump := func(eventTypeName string, expected int) {
		x := suite.getAllEvents(eventTypeName)
		assert.Len(*x, expected)
	}

	et1Id := eventtypeF(eventTypeName1)
	et2Id := eventtypeF(eventTypeName2)

	{
		x := suite.getAllEventTypes()
		assert.Len(*x, 2)
	}

	dump(eventTypeName1, 0)
	dump("", 0)

	{
		x := suite.getAllEventTypes()
		assert.Len(*x, 2)
	}

	{
		x := suite.getOneEventType(et1Id)
		assert.EqualValues(string(et1Id), string((*x).ID))
	}

	e1Id := eventF(et1Id, eventTypeName1, 17)
	dump(eventTypeName1, 1)

	e2Id := eventF(et1Id, eventTypeName1, 18)
	dump(eventTypeName1, 2)

	e3Id := eventF(et2Id, eventTypeName2, 19)
	dump(eventTypeName2, 1)

	dump("", 3)

	suite.deleteOneEvent(eventTypeName1, e1Id)
	suite.deleteOneEvent(eventTypeName1, e2Id)
	suite.deleteOneEvent(eventTypeName2, e3Id)

	suite.deleteOneEventType(et1Id)
	suite.deleteOneEventType(et2Id)
}

func (suite *ServerTester) Test_99_Noop() {
	t := suite.T()
	assert := assert.New(t)
	assert.Equal(17, 10+7)
}
