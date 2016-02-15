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

package main

import (
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-workflow/client"
	"github.com/venicegeo/pz-workflow/server"
	"github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
	"log"
	"testing"
	"time"
)

type WorkflowTester struct {
	suite.Suite
	logger     loggerPkg.ILoggerService
	uuidgenner uuidgenPkg.IUuidGenService
	workflow   client.IWorkflowService
	sys        *piazza.System
}

func (suite *WorkflowTester) SetupSuite() {
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

	routes, err := server.CreateHandlers(sys, suite.logger, suite.uuidgenner)
	if err != nil {
		log.Fatal(err)
	}

	_ = sys.StartServer(routes)

	suite.workflow, err = client.NewPzWorkflowService(sys, sys.Config.GetBindToAddress())
	if err != nil {
		log.Fatal(err)
	}

	suite.sys = sys

	assert.Len(sys.Services, 5)

	suite.assertNoData()
}

func (suite *WorkflowTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func (suite *WorkflowTester) assertNoData() {
	t := suite.T()

	var err error

	es, err := suite.workflow.GetFromEvents()
	assert.NoError(t, err)
	assert.Len(t, *es, 0)

	as, err := suite.workflow.GetFromAlerts()
	assert.NoError(t, err)
	assert.Len(t, *as, 0)

	xs, err := suite.workflow.GetFromTriggers()
	assert.NoError(t, err)
	assert.Len(t, *xs, 0)
}

func TestRunSuite(t *testing.T) {
	s := new(WorkflowTester)
	suite.Run(t, s)
}

//---------------------------------------------------------------------------

func (suite *WorkflowTester) TestAAResource() {
	t := suite.T()
	assert := assert.New(t)
	//workflow := suite.workflow

	es := suite.sys.ElasticSearchService

	var a1 client.Event
	a1.Type = client.EventFoo
	a1.Date = time.Now()

	var a2 client.Event
	a2.Type = client.EventBar
	a2.Date = time.Now()

	db, err := client.NewResourceDB(es, "event", "Event")
	assert.NoError(err)

	a1Id, err := db.PostData(&a1, client.NewResourceID())
	assert.NoError(err)
	a2Id, err := db.PostData(&a2, client.NewResourceID())
	assert.NoError(err)

	{
		raws, err := db.GetAll()
		assert.NoError(err)
		assert.Len(raws, 2)

		objs, err := client.ConvertRawsToEvents(raws)
		assert.NoError(err)

		ok1 := (objs[0].Type == a1.Type) && (objs[1].Type == a2.Type)
		ok2 := (objs[1].Type == a1.Type) && (objs[0].Type == a2.Type)
		assert.True((ok1 || ok2) && !(ok1 && ok2))
	}

	var t2 client.Event
	ok, err := db.GetById(a2Id, &t2)
	assert.NoError(err)
	assert.True(ok)
	assert.EqualValues(a2.Type, t2.Type)

	var t1 client.Event
	ok, err = db.GetById(a1Id, &t1)
	assert.NoError(err)
	assert.True(ok)
	assert.EqualValues(a1.Type, t1.Type)
}

func (suite *WorkflowTester) TestAlerts() {
	t := suite.T()
	assert := assert.New(t)

	suite.assertNoData()

	workflow := suite.workflow

	var err error
	var idResponse *client.WorkflowIdResponse

	var a1 client.Alert
	a1.TriggerId = "this is trigger 1"
	idResponse, err = workflow.PostToAlerts(&a1)
	assert.NoError(err)
	a1ID := idResponse.ID
	assert.EqualValues("A1", a1ID)

	var a2 client.Alert
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


func (suite *WorkflowTester) TestTriggers() {
	t := suite.T()
	assert := assert.New(t)

	suite.assertNoData()

	workflow := suite.workflow

	var err error
	var idResponse *client.WorkflowIdResponse

	x1 := client.Trigger{
		Title: "the x1 trigger",
		Condition: client.Condition{
			Type: client.EventFoo,
			Query: "the x1 condition query",
		},
		Job: client.Job{
			Task: "the x1 task",
		},
	}
	idResponse, err = workflow.PostToTriggers(&x1)
	assert.NoError(err)
	x1Id := idResponse.ID

	x2 := client.Trigger{
		Title: "the x2 trigger",
		Condition: client.Condition{
			Type: client.EventBar,
			Query: "the x2 condition query",
		},
		Job: client.Job{
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

	suite.assertNoData()
}

func (suite *WorkflowTester) TestEvents() {
	t := suite.T()
	assert := assert.New(t)

	workflow := suite.workflow
	suite.assertNoData()

	var err error
	var idResponse *client.WorkflowIdResponse

	var e1 client.Event
	e1.Type = client.EventDataIngested
	e1.Date = time.Now()
	e1.Data = nil
	idResponse, err = workflow.PostToEvents(&e1)
	assert.NoError(err)
	e1ID := idResponse.ID
	assert.EqualValues("E1", e1ID)

	var e2 client.Event
	e2.Type = client.EventDataAccessed
	e2.Date = time.Now()
	e2.Data = nil
	idResponse, err = workflow.PostToEvents(&e2)
	assert.NoError(err)
	e2ID := idResponse.ID
	assert.EqualValues("E2", e2ID)

	es, err := workflow.GetFromEvents()
	assert.NoError(err)
	assert.Len(*es, 2)
	ok1 := false
	ok2 := false
	for _, v := range *es {
		if v.ID == "E1" {
			ok1 = true
		}
		if v.ID == "E2" {
			ok2 = true
		}
	}
	assert.True(ok1 && ok2)

	err = workflow.DeleteOfEvent("E1")
	assert.NoError(err)
	err = workflow.DeleteOfEvent("E2")
	assert.NoError(err)

	suite.assertNoData()
}

func (suite *WorkflowTester) TestTriggering() {

	t := suite.T()
	assert := assert.New(t)

	workflow := suite.workflow

	suite.assertNoData()

	var err error
	var idResponse *client.WorkflowIdResponse

	////////////////

	x1 := client.Trigger{
		Title: "the x1 trigger",
		Condition: client.Condition{
			Type: client.EventFoo,
			Query: "the x1 condition query",
		},
		Job: client.Job{
			Task: "the x1 task",
		},
	}
	idResponse, err = workflow.PostToTriggers(&x1)
	assert.NoError(err)
	x1Id := idResponse.ID

	x2 := client.Trigger{
		Title: "the x2 trigger",
		Condition: client.Condition{
			Type: client.EventBar,
			Query: "the x2 condition query",
		},
		Job: client.Job{
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
	var e1 client.Event
	e1.Type = client.EventFoo
	e1.Date = time.Now()
	e1.Data = map[string]string{"file": "e1.tif"}
	idResponse, err = workflow.PostToEvents(&e1)
	assert.NoError(err)
	e1Id := idResponse.ID

	// will cause trigger X2
	var e2 client.Event
	e2.Type = client.EventBar
	e2.Date = time.Now()
	e2.Data = map[string]string{"file": "e2.tif"}
	idResponse, err = workflow.PostToEvents(&e2)
	assert.NoError(err)
	e2Id := idResponse.ID

	// will cause no triggers
	var e3 client.Event
	e3.Type = client.EventBuz
	e3.Date = time.Now()
	idResponse, err = workflow.PostToEvents(&e3)
	assert.NoError(err)
	e3Id := idResponse.ID

	////////////////////////////////////

	as, err := workflow.GetFromAlerts()
	assert.NoError(err)
	assert.Len(*as, 2)

	var list client.AlertList
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
	workflow.DeleteOfEvent(e1Id)
	workflow.DeleteOfEvent(e2Id)
	workflow.DeleteOfEvent(e3Id)
	workflow.DeleteOfAlert(alerts[0].ID)
	workflow.DeleteOfAlert(alerts[1].ID)

	suite.assertNoData()
}

func (suite *WorkflowTester) TestAdmin() {
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
