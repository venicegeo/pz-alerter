package main

import (
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-alerter/client"
	"github.com/venicegeo/pz-alerter/server"
	piazza "github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
	"log"
	"testing"
	"time"
)

type AlerterTester struct {
	suite.Suite
	logger     loggerPkg.ILoggerService
	uuidgenner uuidgenPkg.IUuidGenService
	alerter    client.IAlerterService
	sys        *piazza.System
}

func (suite *AlerterTester) SetupSuite() {
	t := suite.T()
	assert := assert.New(t)

	config, err := piazza.NewConfig(piazza.PzAlerter, piazza.ConfigModeTest)
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

	suite.alerter, err = client.NewPzAlerterService(sys, sys.Config.GetBindToAddress())
	if err != nil {
		log.Fatal(err)
	}

	suite.sys = sys

	assert.Len(sys.Services, 5)
}

func (suite *AlerterTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func TestRunSuite(t *testing.T) {
	s := new(AlerterTester)
	suite.Run(t, s)
}

//---------------------------------------------------------------------------

func (suite *AlerterTester) TestConditions() {
	t := suite.T()
	assert := assert.New(t)

	alerter := suite.alerter

	var err error
	var idResponse *client.AlerterIdResponse

	var c1 client.Condition
	c1.Title = "c1"
	c1.Type = "Foo"
	c1.Query = "query string"
	idResponse, err = alerter.PostToConditions(&c1)
	assert.NoError(err)
	c1ID := idResponse.ID
	assert.EqualValues("C1", c1ID)

	var c2 client.Condition
	c2.Title = "c2"
	c2.Type = "Bar"
	c2.Query = "another query string"
	idResponse, err = alerter.PostToConditions(&c2)
	assert.NoError(err)
	c2ID := idResponse.ID
	assert.EqualValues("C2", c2ID)

	cs, err := alerter.GetFromConditions()
	assert.NoError(err)
	assert.Len(*cs, 2)
	ok1 := false
	ok2 := false
	for k := range *cs {
		if k == "C1" {
			ok1 = true
		}
		if k == "C2" {
			ok2 = true
		}
	}

	assert.True(ok1 && ok2)
	cond, err := alerter.GetFromCondition("C1")
	assert.NoError(err)
	assert.NotNil(cond)

	err = alerter.DeleteOfCondition("C1")
	assert.NoError(err)

	cond, err = alerter.GetFromCondition("C1")
	assert.Error(err) // TODO: should be more refined error here
	assert.Nil(cond)

	err = alerter.DeleteOfCondition("C2")
	assert.NoError(err)

	cs, err = alerter.GetFromConditions()
	assert.NoError(err)
	assert.Len(*cs, 0)
}

func (suite *AlerterTester) TestActions() {
	t := suite.T()
	assert := assert.New(t)

	alerter := suite.alerter

	var err error
	var idResponse *client.AlerterIdResponse

	var x1 client.Action
	x1.Events = []client.Ident{client.Ident("e1"), client.Ident("e2")}
	x1.Conditions = []client.Ident{client.Ident("c1"), client.Ident("c2")}
	x1.Job = "job message 1"
	idResponse, err = alerter.PostToActions(&x1)
	assert.NoError(err)
	c1ID := idResponse.ID
	assert.EqualValues("X1", c1ID)

	var x2 client.Action
	x2.Events = []client.Ident{client.Ident("e3"), client.Ident("e4")}
	x2.Conditions = []client.Ident{client.Ident("c3"), client.Ident("c4")}
	x2.Job = "job message 2"
	idResponse, err = alerter.PostToActions(&x2)
	assert.NoError(err)
	c2ID := idResponse.ID
	assert.EqualValues("X2", c2ID)

	cs, err := alerter.GetFromActions()
	assert.NoError(err)
	assert.Len(*cs, 2)
	ok1 := false
	ok2 := false
	for k := range *cs {
		if k == "X1" {
			ok1 = true
		}
		if k == "X2" {
			ok2 = true
		}
	}
	assert.True(ok1 && ok2)

	cond, err := alerter.GetFromAction("X1")
	assert.NoError(err)
	assert.NotNil(cond)
}

func (suite *AlerterTester) TestEvents() {
	t := suite.T()
	assert := assert.New(t)

	alerter := suite.alerter

	var err error
	var idResponse *client.AlerterIdResponse

	var e1 client.Event
	e1.Type = client.EventDataIngested
	e1.Date = time.Now()
	e1.Data = nil
	idResponse, err = alerter.PostToEvents(&e1)
	assert.NoError(err)
	e1ID := idResponse.ID
	assert.EqualValues("E1", e1ID)

	var e2 client.Event
	e2.Type = client.EventDataAccessed
	e2.Date = time.Now()
	e2.Data = nil
	idResponse, err = alerter.PostToEvents(&e2)
	assert.NoError(err)
	e2ID := idResponse.ID
	assert.EqualValues("E2", e2ID)

	es, err := alerter.GetFromEvents()
	assert.NoError(err)
	assert.Len(*es, 2)
	ok1 := false
	ok2 := false
	for k := range *es {
		if k == "E1" {
			ok1 = true
		}
		if k == "E2" {
			ok2 = true
		}
	}
	assert.True(ok1 && ok2)
}

func (suite *AlerterTester) TestTriggering() {
	t := suite.T()
	assert := assert.New(t)

	alerter := suite.alerter

	var err error
	var idResponse *client.AlerterIdResponse

	cs, err := alerter.GetFromConditions()
	assert.NoError(err)
	assert.Len(*cs, 0)

	var rawC3 client.Condition
	rawC3.Title = "cond1 title"
	rawC3.Type = client.EventDataIngested
	rawC3.Query = "c3 query"
	idResponse, err = alerter.PostToConditions(&rawC3)
	assert.NoError(err)
	c3ID := idResponse.ID
	assert.EqualValues("C3", c3ID)

	var rawC4 client.Condition
	rawC4.Title = "cond2 title"
	rawC4.Type = client.EventDataAccessed
	rawC4.Query = "c4 query"
	idResponse, err = alerter.PostToConditions(&rawC4)
	assert.NoError(err)
	c4ID := idResponse.ID
	assert.EqualValues("C4", c4ID)

	var rawC5 client.Condition
	rawC5.Title = "cond2 title"
	rawC5.Type = client.EventFoo
	rawC5.Query = "c5 query"
	idResponse, err = alerter.PostToConditions(&rawC5)
	assert.NoError(err)
	c5ID := idResponse.ID
	assert.EqualValues("C5", c5ID)

	cs, err = alerter.GetFromConditions()
	assert.NoError(err)
	assert.Len(*cs, 3)

	var e3 client.Event
	e3.Type = client.EventDataAccessed
	e3.Date = time.Now()
	e3.Data = map[string]string{"file": "111.tif"}
	idResponse, err = alerter.PostToEvents(&e3)
	assert.NoError(err)
	e3ID := idResponse.ID
	assert.EqualValues("E3", e3ID)

	var e4 client.Event
	e4.Type = client.EventDataIngested
	e4.Date = time.Now()
	e4.Data = map[string]string{"file": "111.tif"}
	idResponse, err = alerter.PostToEvents(&e4)
	assert.NoError(err)
	e4ID := idResponse.ID
	assert.EqualValues("E4", e4ID)

	var e5 client.Event
	e5.Type = client.EventBar
	e5.Date = time.Now()
	idResponse, err = alerter.PostToEvents(&e5)
	assert.NoError(err)
	e5ID := idResponse.ID
	assert.EqualValues("E5", e5ID)

	as, err := alerter.GetFromAlerts()
	assert.NoError(err)
	assert.Len(*as, 2)

	// TODO: dependent on order of returned results
	v1, ok := (*as)["A1"]
	assert.True(ok)
	v2, ok := (*as)["A2"]
	assert.True(ok)

	assert.EqualValues(client.Ident("A1"), v1.ID)
	assert.EqualValues(client.Ident("X4"), v1.Action)
	assert.EqualValues(client.Ident("A2"), v2.ID)
	assert.EqualValues(client.Ident("X4"), v2.Action)

	alerter.DeleteOfCondition("C3")
	alerter.DeleteOfCondition("C4")
	alerter.DeleteOfCondition("C5")
}

func (suite *AlerterTester) TestAdmin() {
	t := suite.T()
	assert := assert.New(t)

	alerter := suite.alerter

	settings, err := alerter.GetFromAdminSettings()
	assert.NoError(err)
	if settings.Debug != false {
		t.Error("settings not false")
	}

	settings.Debug = true
	err = alerter.PostToAdminSettings(settings)
	assert.NoError(err)

	settings, err = alerter.GetFromAdminSettings()
	assert.NoError(err)
	if settings.Debug != true {
		t.Error("settings not true")
	}
}
