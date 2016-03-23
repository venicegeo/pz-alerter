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
	"time"

	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
)

type ClientTester struct {
	suite.Suite
	logger     loggerPkg.ILoggerService
	uuidgenner uuidgenPkg.IUuidGenService
	workflow   *PzWorkflowService
	sys        *piazza.System
}

func (suite *ClientTester) SetupSuite() {
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

	suite.logger, err = loggerPkg.NewMockLoggerService(sys)
	if err != nil {
		log.Fatal(err)
	}
	var tmp loggerPkg.ILoggerService = suite.logger
	clogger := loggerPkg.NewCustomLogger(&tmp, piazza.PzWorkflow, config.GetAddress())

	suite.uuidgenner, err = uuidgenPkg.NewMockUuidGenService(sys)
	if err != nil {
		log.Fatal(err)
	}

	es, err := elasticsearch.NewElasticsearchClient(sys, true)
	if err != nil {
		log.Fatal(err)
	}
	sys.Services[piazza.PzElasticSearch] = es

	routes, err := CreateHandlers(sys, clogger, suite.uuidgenner)
	if err != nil {
		log.Fatal(err)
	}

	_ = sys.StartServer(routes)

	suite.workflow, err = NewPzWorkflowService(sys, sys.Config.GetBindToAddress())
	if err != nil {
		log.Fatal(err)
	}

	suite.sys = sys

	assert.Len(sys.Services, 5)

	suite.assertNoData()
}

func (suite *ClientTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func (suite *ClientTester) assertNoData() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	var err error

	{
		ts, err := workflow.GetAllEventTypes()
		log.Printf("***** %#v ***** %#v *****", ts, err)
	}

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

//---------------------------------------------------------------------------

func (suite *ClientTester) TestAdmin() {
	t := suite.T()
	assert := assert.New(t)

	workflow := suite.workflow

	settings, err := workflow.GetFromAdminSettings()
	assert.NoError(err)
	if settings.Debug != false {
		t.Error("settings not false")
	}

	settings.Debug = true
	err = workflow.PostToAdminSettings(settings)
	assert.NoError(err)

	settings, err = workflow.GetFromAdminSettings()
	assert.NoError(err)
	if settings.Debug != true {
		t.Error("settings not true")
	}
}

func (suite *ClientTester) TestAlertResource() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	suite.assertNoData()
	defer suite.assertNoData()

	var err error

	a1 := Alert{TriggerId: "dummyT1", EventId: "dummyE1"}
	id, err := workflow.PostOneAlert(&a1)
	assert.NoError(err)

	alerts, err := workflow.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 1)
	assert.EqualValues(id, (*alerts)[0].ID)
	assert.EqualValues("dummyT1", (*alerts)[0].TriggerId)
	assert.EqualValues("dummyE1", (*alerts)[0].EventId)

	alert, err := workflow.GetOneAlert(id)
	assert.NoError(err)
	assert.EqualValues(id, alert.ID)

	alert, err = workflow.GetOneAlert("nosuchalert1")
	assert.Error(err)

	err = workflow.DeleteOneAlert("nosuchalert2")
	assert.Error(err)

	err = workflow.DeleteOneAlert(id)
	assert.NoError(err)

	alert, err = workflow.GetOneAlert(id)
	assert.Error(err)
	assert.Nil(alert)

	alerts, err = workflow.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 0)
}

func (suite *ClientTester) TestEventResource() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	suite.assertNoData()
	defer suite.assertNoData()

	var err error

	mapping := map[string]elasticsearch.MappingElementTypeName{
		"myint": elasticsearch.MappingElementTypeString,
		"mystr": elasticsearch.MappingElementTypeString,
	}
	eventTypeName := "mytype"
	eventType := &EventType{Name: eventTypeName, Mapping: mapping}
	etId, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)
	defer func() {
		err = workflow.DeleteOneEventType(etId)
		assert.NoError(err)
	}()

	event := &Event{
		EventTypeId: etId,
		Date:        time.Now(),
		Data: map[string]interface{}{
			"myint": 17,
			"mystr": "quick",
		},
	}
	eId, err := workflow.PostOneEvent(eventTypeName, event)
	assert.NoError(err)

	defer func() {
		err = workflow.DeleteOneEvent(eventTypeName, eId)
		assert.NoError(err)
	}()

	events, err := workflow.GetAllEvents("")
	assert.NoError(err)
	assert.Len(*events, 1)
	assert.EqualValues(eId, (*events)[0].ID)

	tmp, err := workflow.GetOneEvent(eventTypeName, eId)
	assert.NoError(err)
	assert.EqualValues(eId, tmp.ID)
}

func (suite *ClientTester) TestEventTypeResource() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	suite.assertNoData()
	defer suite.assertNoData()

	var err error

	mapping := map[string]elasticsearch.MappingElementTypeName{
		"myint": elasticsearch.MappingElementTypeString,
		"mystr": elasticsearch.MappingElementTypeString,
	}
	eventType := &EventType{Name: "typnam", Mapping: mapping}
	id, err := workflow.PostOneEventType(eventType)
	defer func() {
		err = workflow.DeleteOneEventType(id)
		assert.NoError(err)
	}()

	eventTypes, err := workflow.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*eventTypes, 1)
	assert.EqualValues(id, (*eventTypes)[0].ID)

	tmp, err := workflow.GetOneEventType(id)
	assert.NoError(err)
	assert.EqualValues(id, tmp.ID)
}

func (suite *ClientTester) TestOne() {

	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	suite.assertNoData()
	defer suite.assertNoData()

	var err error
	var eventTypeName = "EventTypeA"

	var etId Ident
	{
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num": elasticsearch.MappingElementTypeInteger,
			"str": elasticsearch.MappingElementTypeString,
		}

		eventType := &EventType{Name: eventTypeName, Mapping: mapping}

		etId, err = workflow.PostOneEventType(eventType)
		assert.NoError(err)

		defer func() {
			err := workflow.DeleteOneEventType(etId)
			assert.NoError(err)
		}()
	}

	var tId Ident
	{
		x1 := &Trigger{
			Title: "the x1 trigger",
			Condition: Condition{
				EventId: etId,
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"num": 17,
						},
					},
				},
			},
			Job: Job{
				Task: "the x1 task",
			},
		}

		tId, err = workflow.PostOneTrigger(x1)
		assert.NoError(err)

		defer func() {
			err := workflow.DeleteOneTrigger(tId)
			assert.NoError(err)
		}()
	}

	var e1Id Ident
	{
		// will cause trigger t1Id
		e1 := &Event{
			EventTypeId: etId,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": 17,
				"str": "quick",
			},
		}

		e1Id, err = workflow.PostOneEvent(eventTypeName, e1)
		assert.NoError(err)

		defer func() {
			err := workflow.DeleteOneEvent(eventTypeName, e1Id)
			assert.NoError(err)
		}()
	}

	var e2Id Ident
	{
		// will cause no triggers
		e2 := &Event{
			EventTypeId: etId,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": 18,
				"str": "brown",
			},
		}

		e2Id, err = workflow.PostOneEvent(eventTypeName, e2)
		assert.NoError(err)

		defer func() {
			err := workflow.DeleteOneEvent(eventTypeName, e2Id)
			assert.NoError(err)
		}()
	}

	{
		ary, err := workflow.GetAllEvents("")
		assert.NoError(err)
		assert.Len(*ary, 2)
	}

	var aId Ident
	{
		alerts, err := workflow.GetAllAlerts()
		assert.NoError(err)
		assert.Len(*alerts, 1)
		var alert0 Alert = (*alerts)[0]
		assert.EqualValues(e1Id, alert0.EventId)
		assert.EqualValues(tId, alert0.TriggerId)

		aId = alert0.ID

		defer func() {
			err := workflow.DeleteOneAlert(aId)
			assert.NoError(err)
		}()
	}
}

func (suite *ClientTester) TestTriggerResource() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	suite.assertNoData()
	defer suite.assertNoData()

	var err error

	mapping := map[string]elasticsearch.MappingElementTypeName{
		"myint": elasticsearch.MappingElementTypeString,
		"mystr": elasticsearch.MappingElementTypeString,
	}
	eventType := &EventType{Name: "typnam", Mapping: mapping}
	etId, err := workflow.PostOneEventType(eventType)

	defer func() {
		err = workflow.DeleteOneEventType(etId)
		assert.NoError(err)
	}()

	t1 := Trigger{
		Title: "the x1 trigger",
		Condition: Condition{
			EventId: etId,
			Query: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"myint": 17,
					},
				},
			},
		},
		Job: Job{
			Task: "the x1 task",
		},
	}
	t1Id, err := workflow.PostOneTrigger(&t1)
	assert.NoError(err)

	defer func() {
		err = workflow.DeleteOneTrigger(t1Id)
		assert.NoError(err)
	}()

	tmp, err := workflow.GetOneTrigger(t1Id)
	assert.NoError(err)
	assert.EqualValues(t1Id, tmp.ID)

	triggers, err := workflow.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*triggers, 1)
	assert.EqualValues(t1Id, (*triggers)[0].ID)
}

func (suite *ClientTester) TestTriggering() {

	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	suite.assertNoData()
	defer suite.assertNoData()

	var err error

	//-----------------------------------------------------

	var etC, etD, etE Ident
	{
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num": elasticsearch.MappingElementTypeInteger,
			"str": elasticsearch.MappingElementTypeString,
		}
		eventTypeC := &EventType{Name: "EventType C", Mapping: mapping}
		eventTypeD := &EventType{Name: "EventType D", Mapping: mapping}
		eventTypeE := &EventType{Name: "EventType E", Mapping: mapping}
		etC, err = workflow.PostOneEventType(eventTypeC)
		assert.NoError(err)
		etD, err = workflow.PostOneEventType(eventTypeD)
		assert.NoError(err)
		etE, err = workflow.PostOneEventType(eventTypeE)
		assert.NoError(err)

		eventTypes, err := workflow.GetAllEventTypes()
		assert.NoError(err)
		assert.Len(*eventTypes, 3)
	}
	defer func() {
		workflow.DeleteOneEventType(etC)
		assert.NoError(err)
		workflow.DeleteOneEventType(etD)
		assert.NoError(err)
		workflow.DeleteOneEventType(etE)
		assert.NoError(err)
	}()

	////////////////

	var tA, tB Ident
	{
		t1 := &Trigger{
			Title: "Trigger A",
			Condition: Condition{
				EventId: etC,
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"str": "quick",
						},
					},
				},
			},
			Job: Job{
				Task: "Trigger A task",
			},
		}
		tA, err = workflow.PostOneTrigger(t1)
		assert.NoError(err)
		defer func() {
			workflow.DeleteOneTrigger(tA)
			assert.NoError(err)
		}()

		t2 := &Trigger{
			Title: "Trigger B",
			Condition: Condition{
				EventId: etD,
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"num": 18,
						},
					},
				},
			},
			Job: Job{
				Task: "Trigger B task",
			},
		}
		tB, err = workflow.PostOneTrigger(t2)
		assert.NoError(err)
		defer func() {
			workflow.DeleteOneTrigger(tB)
			assert.NoError(err)
		}()

		triggers, err := workflow.GetAllTriggers()
		assert.NoError(err)
		assert.Len(*triggers, 2)
	}

	var eF, eG, eH Ident
	{
		// will cause trigger TA
		e1 := Event{
			EventTypeId: etC,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": 17,
				"str": "quick",
			},
		}
		eF, err = workflow.PostOneEvent("EventType C", &e1)
		assert.NoError(err)
		defer func() {
			workflow.DeleteOneEvent("EventType C", eF)
			assert.NoError(err)
		}()

		// will cause trigger TB
		e2 := Event{
			EventTypeId: etD,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": 18,
				"str": "brown",
			},
		}
		eG, err = workflow.PostOneEvent("EventType D", &e2)
		assert.NoError(err)
		defer func() {
			workflow.DeleteOneEvent("EventType D", eG)
			assert.NoError(err)
		}()

		// will cause no triggers
		e3 := Event{
			EventTypeId: etE,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": 19,
				"str": "fox",
			},
		}
		eH, err = workflow.PostOneEvent("EventType E", &e3)
		assert.NoError(err)
		defer func() {
			workflow.DeleteOneEvent("EventType E", eH)
			assert.NoError(err)
		}()
	}

	var aI, aJ Ident
	{
		alerts, err := workflow.GetAllAlerts()
		assert.NoError(err)
		assert.Len(*alerts, 2)

		var alert0, alert1 *Alert
		if (*alerts)[0].EventId == eF {
			alert0 = &(*alerts)[0]
			alert1 = &(*alerts)[1]
		} else {
			alert0 = &(*alerts)[1]
			alert1 = &(*alerts)[0]
		}

		aI = alert0.ID
		aJ = alert1.ID

		assert.EqualValues(alert0.TriggerId, tA)
		assert.EqualValues(alert0.EventId, eF)
		assert.EqualValues(alert1.TriggerId, tB)
		assert.EqualValues(alert1.EventId, eG)

		defer func() {
			workflow.DeleteOneAlert(aI)
			assert.NoError(err)
		}()
		defer func() {
			workflow.DeleteOneAlert(aJ)
			assert.NoError(err)
		}()
	}
}
