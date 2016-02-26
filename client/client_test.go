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

package client

import (
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
	"github.com/venicegeo/pz-workflow/common"
	_server "github.com/venicegeo/pz-workflow/server"
	"log"
	"testing"
	"time"
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

	suite.uuidgenner, err = uuidgenPkg.NewMockUuidGenService(sys)
	if err != nil {
		log.Fatal(err)
	}

	routes, err := _server.CreateHandlers(sys, suite.logger, suite.uuidgenner)
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

	var err error

	es, err := suite.workflow.GetFromEvents()
	assert.NoError(t, err)
	assert.Len(t, *es, 0)

	ts, err := suite.workflow.GetFromEventTypes()
	assert.NoError(t, err)
	assert.Len(t, *ts, 0)

	as, err := suite.workflow.GetFromAlerts()
	assert.NoError(t, err)
	assert.Len(t, *as, 0)

	xs, err := suite.workflow.GetFromTriggers()
	assert.NoError(t, err)
	assert.Len(t, *xs, 0)
}

func (suite *ClientTester) createEventType(name string, mapping map[string]piazza.MappingElementTypeName) common.Ident {
	t := suite.T()
	assert := assert.New(t)
	workflow := suite.workflow

	if mapping == nil {
		mapping = map[string]piazza.MappingElementTypeName{
			"int":  piazza.MappingElementTypeString,
			"str": piazza.MappingElementTypeString,
		}
	}

	et := &common.EventType{Name: name, Mapping: mapping}
	idResp, err := workflow.PostToEventTypes(et)
	assert.NoError(err)
	assert.NotNil(idResp)

	return idResp.ID
}

func TestRunSuite(t *testing.T) {
	s := new(ClientTester)
	suite.Run(t, s)
}

//---------------------------------------------------------------------------

func (suite *ClientTester) TestAlertResource() {
	t := suite.T()
	assert := assert.New(t)

	suite.assertNoData()

	workflow := suite.workflow

	var err error
	var idResponse *common.WorkflowIdResponse

	var a1 common.Alert
	a1.TriggerId = "this is trigger 1"
	idResponse, err = workflow.PostToAlerts(&a1)
	assert.NoError(err)
	a1ID := idResponse.ID
	assert.EqualValues("A1", a1ID)

	var a2 common.Alert
	a2.TriggerId = "this is trigger 2"
	idResponse, err = workflow.PostToAlerts(&a2)
	assert.NoError(err)
	a2ID := idResponse.ID
	assert.EqualValues("A2", a2ID)

	as, err := workflow.GetFromAlerts()
	assert.NoError(err)
	assert.Len(*as, 2)
	ok1 := false
	ok2 := false
	for _, v := range *as {
		if v.ID == "A1" {
			ok1 = true
		}
		if v.ID == "A2" {
			ok2 = true
		}
	}

	assert.True(ok1 && ok2)
	alert, err := workflow.GetFromAlert("A1")
	assert.NoError(err)
	assert.NotNil(alert)

	err = workflow.DeleteOfAlert("A1")
	assert.NoError(err)

	alert, err = workflow.GetFromAlert("A1")
	assert.Error(err)
	assert.Nil(alert)

	err = workflow.DeleteOfAlert("A2")
	assert.NoError(err)

	as, err = workflow.GetFromAlerts()
	assert.NoError(err)
	assert.Len(*as, 0)

	suite.assertNoData()
}

func (suite *ClientTester) TestTriggerResource() {
	t := suite.T()
	assert := assert.New(t)

	suite.assertNoData()

	workflow := suite.workflow

	var err error
	var idResponse *common.WorkflowIdResponse

	et3 := suite.createEventType("EventTypeC", nil)
	et4 := suite.createEventType("EventTypeD", nil)

	x1 := common.Trigger{
		Title: "the x1 trigger",
		Condition: common.Condition{
			EventType: et3,
			Query:     "the x1 condition query",
		},
		Job: common.Job{
			Task: "the x1 task",
		},
	}
	idResponse, err = workflow.PostToTriggers(&x1)
	assert.NoError(err)
	x1Id := idResponse.ID

	x2 := common.Trigger{
		Title: "the x2 trigger",
		Condition: common.Condition{
			EventType: et4,
			Query:     "the x2 condition query",
		},
		Job: common.Job{
			Task: "the x2 task",
		},
	}
	idResponse, err = workflow.PostToTriggers(&x2)
	assert.NoError(err)
	x2Id := idResponse.ID

	cs, err := workflow.GetFromTriggers()
	assert.NoError(err)
	assert.Len(*cs, 2)
	ok1 := false
	ok2 := false
	for _, v := range *cs {
		if v.ID == x1Id {
			ok1 = true
		}
		if v.ID == x2Id {
			ok2 = true
		}
	}
	assert.True(ok1 && ok2)

	tmp, err := workflow.GetFromTrigger(x1Id)
	assert.NoError(err)
	assert.NotNil(tmp)

	err = workflow.DeleteOfTrigger(x1Id)
	assert.NoError(err)
	err = workflow.DeleteOfTrigger(x2Id)
	assert.NoError(err)

	err = workflow.DeleteOfEventType(et3)
	assert.NoError(err)
	err = workflow.DeleteOfEventType(et4)
	assert.NoError(err)

	suite.assertNoData()
}

func (suite *ClientTester) TestEventResource() {
	t := suite.T()
	assert := assert.New(t)

	workflow := suite.workflow
	suite.assertNoData()

	var err error
	var idResponse *common.WorkflowIdResponse

	mappingE := map[string]piazza.MappingElementTypeName{
		"myint":  piazza.MappingElementTypeInteger,
		"mystr": piazza.MappingElementTypeString,
	}
	mappingF := map[string]piazza.MappingElementTypeName{
		"thestr": piazza.MappingElementTypeString,
	}

	et5 := suite.createEventType("EventTypeE", mappingE)
	et6 := suite.createEventType("EventTypeF", mappingF)

	eventE := common.Event{EventType: "EventTypeE", Date: time.Now(),
		Data: map[string]interface{}{"myint": 47, "mystr": "forty-seven"}}
	eventF := common.Event{EventType: "EventTypeF", Date: time.Now(),
		Data: map[string]interface{}{"thestr": "quick brown fox"}}

	idResponse, err = workflow.PostToEvents("EventTypeE", &eventE)
	assert.NoError(err)
	e1ID := idResponse.ID
	assert.EqualValues("E1", e1ID)

	idResponse, err = workflow.PostToEvents("EventTypeF", &eventF)
	assert.NoError(err)
	e2ID := idResponse.ID
	assert.EqualValues("E2", e2ID)

	es, err := workflow.GetFromEvents()
	assert.NoError(err)
	assert.Len(*es, 2)
	ok1 := false
	ok2 := false
	for _, v := range *es {
		log.Printf("GOT EVENT: %#v", v)
		log.Printf("%s: %s %s", v.ID, e1ID, e2ID)
		if v.ID == e1ID {
			ok1 = true
		} else
		if v.ID == e2ID {
			ok2 = true
		} else {
			assert.False(true)
		}
	}
	assert.True(ok1 && ok2)

	err = workflow.DeleteOfEvent("EventTypeE", "E1")
	assert.NoError(err)
	err = workflow.DeleteOfEvent("EventTypeF", "E2")
	assert.NoError(err)

	err = workflow.DeleteOfEventType(et5)
	assert.NoError(err)
	err = workflow.DeleteOfEventType(et6)
	assert.NoError(err)

	suite.assertNoData()
}

func (suite *ClientTester) TestEventTypeResource() {
	t := suite.T()
	assert := assert.New(t)

	workflow := suite.workflow
	suite.assertNoData()

	var err error

	mapping := map[string]piazza.MappingElementTypeName{
		"myint":  piazza.MappingElementTypeString,
		"mystr": piazza.MappingElementTypeString,
	}

	et7 := suite.createEventType("MyTestObj", mapping)

	es, err := workflow.GetFromEventTypes()
	assert.NoError(err)
	assert.Len(*es, 1)

	assert.EqualValues((*es)[0].ID, et7)

	err = workflow.DeleteOfEventType(et7)
	assert.NoError(err)

	suite.assertNoData()
}

func (suite *ClientTester) TestAAATriggering() {

	t := suite.T()
	assert := assert.New(t)

	workflow := suite.workflow

	suite.assertNoData()

	var err error
	var idResponse *common.WorkflowIdResponse

	mapping := map[string]piazza.MappingElementTypeName{
		"id": piazza.MappingElementTypeString,
		"num":  piazza.MappingElementTypeInteger,
		"str": piazza.MappingElementTypeString,
	}

	et8 := suite.createEventType("EventTypeH", mapping)
	et9 := suite.createEventType("EventTypeI", mapping)
	et10 := suite.createEventType("EventTypeJ", mapping)

	////////////////

	x1 := common.Trigger{
		Title: "the x1 trigger",
		Condition: common.Condition{
			EventType: et8,
			Query:
			`{
				"query": {
					"match": {
						"str": {
							"query": "quick"
						}
					}
				}
			}`,
		},
		Job: common.Job{
			Task: "the x1 task",
		},
	}
	idResponse, err = workflow.PostToTriggers(&x1)
	assert.NoError(err)
	x1Id := idResponse.ID

	x2 := common.Trigger{
		Title: "the x2 trigger",
		Condition: common.Condition{
			EventType: et9,
			Query:
			`{
				"query": {
					"match": {
						"num": {
							"query": 18
						}
					}
				}
			}`,
		},
		Job: common.Job{
			Task: "the x2 task",
		},
	}
	idResponse, err = workflow.PostToTriggers(&x2)
	assert.NoError(err)
	x2Id := idResponse.ID

	xs, err := workflow.GetFromTriggers()
	assert.NoError(err)
	assert.Len(*xs, 2)

	/////////////////////

	// will cause trigger X1
	e1 := common.Event{
		ID: "e1",
		EventType: "EventTypeH",
		Date: time.Now(),
		Data: map[string]interface{}{
			"num": 17,
			"str": "quick",
		},
	}
	idResponse, err = workflow.PostToEvents("EventTypeH", &e1)
	assert.NoError(err)
	e1Id := idResponse.ID

	// will cause trigger X2
	e2 := common.Event{
		ID: "e2",
		EventType: "EventTypeI",
		Date: time.Now(),
		Data: map[string]interface{}{
			"num": 18,
			"str": "brown",
		},
	}
	idResponse, err = workflow.PostToEvents("EventTypeI", &e2)
	assert.NoError(err)
	e2Id := idResponse.ID

	// will cause no triggers
	e3 := common.Event{
		ID: "e3",
		EventType: "EventTypeJ",
		Date: time.Now(),
		Data: map[string]interface{}{
			"num": 19,
			"str": "fox",
		},
	}
	idResponse, err = workflow.PostToEvents("EventTypeJ", &e3)
	assert.NoError(err)
	e3Id := idResponse.ID

	////////////////////////////////////

	as, err := workflow.GetFromAlerts()
	assert.NoError(err)
	assert.Len(*as, 2)

	var list common.AlertList
	list = *as
	alerts := list.ToSortedArray()
	assert.Len(alerts, 2)

	t1 := (alerts[0].TriggerId == x1Id)
	t2 := (alerts[1].TriggerId == x2Id)
	t3 := (alerts[0].TriggerId == x2Id)
	t4 := (alerts[1].TriggerId == x1Id)
	assert.True((t1 && t2) || (t3 && t4))

	//////////////

	workflow.DeleteOfTrigger(x1Id)
	workflow.DeleteOfTrigger(x2Id)
	workflow.DeleteOfEvent("EventTypeH", e1Id)
	workflow.DeleteOfEvent("EventTypeI", e2Id)
	workflow.DeleteOfEvent("EventTypeJ", e3Id)
	workflow.DeleteOfAlert(alerts[0].ID)
	workflow.DeleteOfAlert(alerts[1].ID)

	err = workflow.DeleteOfEventType(et8)
	assert.NoError(err)
	err = workflow.DeleteOfEventType(et9)
	assert.NoError(err)
	err = workflow.DeleteOfEventType(et10)
	assert.NoError(err)

	suite.assertNoData()
}

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
