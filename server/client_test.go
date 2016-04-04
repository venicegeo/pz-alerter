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
	sys        *piazza.SystemConfig
}

func (suite *ClientTester) SetupSuite() {
	assertNoData(suite.T(), suite.workflow)
}

func (suite *ClientTester) TearDownSuite() {
	assertNoData(suite.T(), suite.workflow)
}

//---------------------------------------------------------------------------

func (suite *ClientTester) Test11Admin() {
	t := suite.T()
	assert := assert.New(t)

	workflow := suite.workflow

	log.Printf("AdminSettings:")
	settings, err := workflow.GetFromAdminSettings()
	assert.NoError(err)
	if settings.Debug != false {
		t.Error("settings not false")
	}
	printJSON("before", settings)

	settings.Debug = true
	err = workflow.PostToAdminSettings(settings)
	assert.NoError(err)

	settings, err = workflow.GetFromAdminSettings()
	assert.NoError(err)
	if settings.Debug != true {
		t.Error("settings not true")
	}
	printJSON("after", settings)
}

func (suite *ClientTester) Test12AlertResource() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), workflow)
	defer assertNoData(suite.T(), workflow)

	var err error

	a1 := Alert{TriggerID: "dummyT1", EventID: "dummyE1"}
	id, err := workflow.PostOneAlert(&a1)
	assert.NoError(err)

	alerts, err := workflow.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 1)
	assert.EqualValues(id, (*alerts)[0].ID)
	assert.EqualValues("dummyT1", (*alerts)[0].TriggerID)
	assert.EqualValues("dummyE1", (*alerts)[0].EventID)

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

func (suite *ClientTester) xTest13EventResource() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), workflow)
	defer assertNoData(suite.T(), workflow)

	var err error

	mapping := map[string]elasticsearch.MappingElementTypeName{
		"myint": elasticsearch.MappingElementTypeString,
		"mystr": elasticsearch.MappingElementTypeString,
	}
	eventTypeName := "mytype"
	eventType := &EventType{Name: eventTypeName, Mapping: mapping}
	etID, err := workflow.PostOneEventType(eventType)
	assert.NoError(err)
	defer func() {
		err = workflow.DeleteOneEventType(etID)
		assert.NoError(err)
	}()

	event := &Event{
		EventTypeID: etID,
		Date:        time.Now(),
		Data: map[string]interface{}{
			"myint": 17,
			"mystr": "quick",
		},
	}
	eID, err := workflow.PostOneEvent(eventTypeName, event)
	assert.NoError(err)

	defer func() {
		err = workflow.DeleteOneEvent(eventTypeName, eID)
		assert.NoError(err)
	}()

	events, err := workflow.GetAllEvents("")
	assert.NoError(err)
	assert.Len(*events, 1)
	assert.EqualValues(eID, (*events)[0].ID)

	tmp, err := workflow.GetOneEvent(eventTypeName, eID)
	assert.NoError(err)
	assert.EqualValues(eID, tmp.ID)
}

func (suite *ClientTester) Test14EventTypeResource() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), workflow)
	defer assertNoData(suite.T(), workflow)

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

func (suite *ClientTester) xTest15One() {

	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), workflow)
	defer assertNoData(suite.T(), workflow)

	var err error
	var eventTypeName = "EventTypeA"

	var etID Ident
	{
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num": elasticsearch.MappingElementTypeInteger,
			"str": elasticsearch.MappingElementTypeString,
		}

		eventType := &EventType{Name: eventTypeName, Mapping: mapping}

		etID, err = workflow.PostOneEventType(eventType)
		assert.NoError(err)

		defer func() {
			err := workflow.DeleteOneEventType(etID)
			assert.NoError(err)
		}()
	}

	var tID Ident
	{
		x1 := &Trigger{
			Title: "the x1 trigger",
			Condition: Condition{
				EventTypeID: etID,
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

		tID, err = workflow.PostOneTrigger(x1)
		assert.NoError(err)

		defer func() {
			err := workflow.DeleteOneTrigger(tID)
			assert.NoError(err)
		}()
	}

	var e1ID Ident
	{
		// will cause trigger t1ID
		e1 := &Event{
			EventTypeID: etID,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": 17,
				"str": "quick",
			},
		}

		e1ID, err = workflow.PostOneEvent(eventTypeName, e1)
		assert.NoError(err)

		defer func() {
			err := workflow.DeleteOneEvent(eventTypeName, e1ID)
			assert.NoError(err)
		}()
	}

	var e2ID Ident
	{
		// will cause no triggers
		e2 := &Event{
			EventTypeID: etID,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": 18,
				"str": "brown",
			},
		}

		e2ID, err = workflow.PostOneEvent(eventTypeName, e2)
		assert.NoError(err)

		defer func() {
			err := workflow.DeleteOneEvent(eventTypeName, e2ID)
			assert.NoError(err)
		}()
	}

	{
		ary, err := workflow.GetAllEvents("")
		assert.NoError(err)
		assert.Len(*ary, 2)
	}

	var aID Ident
	{
		alerts, err := workflow.GetAllAlerts()
		assert.NoError(err)
		assert.Len(*alerts, 1)
		alert0 := (*alerts)[0]
		assert.EqualValues(e1ID, alert0.EventID)
		assert.EqualValues(tID, alert0.TriggerID)

		aID = alert0.ID

		defer func() {
			err := workflow.DeleteOneAlert(aID)
			assert.NoError(err)
		}()
	}
}

func (suite *ClientTester) Test16TriggerResource() {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), workflow)
	defer assertNoData(suite.T(), workflow)

	var err error

	mapping := map[string]elasticsearch.MappingElementTypeName{
		"myint": elasticsearch.MappingElementTypeString,
		"mystr": elasticsearch.MappingElementTypeString,
	}
	eventType := &EventType{Name: "typnam", Mapping: mapping}
	etID, err := workflow.PostOneEventType(eventType)

	defer func() {
		err = workflow.DeleteOneEventType(etID)
		assert.NoError(err)
	}()

	t1 := Trigger{
		Title: "the x1 trigger",
		Condition: Condition{
			EventTypeID: etID,
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
	t1ID, err := workflow.PostOneTrigger(&t1)
	assert.NoError(err)

	defer func() {
		err = workflow.DeleteOneTrigger(t1ID)
		assert.NoError(err)
	}()

	tmp, err := workflow.GetOneTrigger(t1ID)
	assert.NoError(err)
	assert.EqualValues(t1ID, tmp.ID)

	triggers, err := workflow.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*triggers, 1)
	assert.EqualValues(t1ID, (*triggers)[0].ID)
}

func (suite *ClientTester) xTest17Triggering() {

	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	assertNoData(suite.T(), workflow)
	defer assertNoData(suite.T(), workflow)

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
				EventTypeID: etC,
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
				EventTypeID: etD,
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
			EventTypeID: etC,
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
			EventTypeID: etD,
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
			EventTypeID: etE,
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
		if (*alerts)[0].EventID == eF {
			alert0 = &(*alerts)[0]
			alert1 = &(*alerts)[1]
		} else {
			alert0 = &(*alerts)[1]
			alert1 = &(*alerts)[0]
		}

		aI = alert0.ID
		aJ = alert1.ID

		assert.EqualValues(alert0.TriggerID, tA)
		assert.EqualValues(alert0.EventID, eF)
		assert.EqualValues(alert1.TriggerID, tB)
		assert.EqualValues(alert1.EventID, eG)

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
