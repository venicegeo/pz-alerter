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

	suite.alerter, err = client.NewPzAlerterService(sys)
	if err != nil {
		log.Fatal(err)
	}

	suite.sys = sys

	assert.Len(t, sys.Services, 4)
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

	alerter := suite.alerter

	var err error
	var idResponse *client.AlerterIdResponse

	var c1 client.Condition
	c1.Title = "c1"
	c1.Type = "=type="
	c1.UserID = "=userid="
	c1.Date = "=date="
	idResponse, err = alerter.PostToConditions(&c1)
	assert.NoError(t, err)
	c1ID := idResponse.ID
	assert.Equal(t, c1ID, "1")

	var c2 client.Condition
	c2.Title = "c2"
	c2.Type = "=type="
	c2.UserID = "=userid="
	c2.Date = "=date="
	idResponse, err = alerter.PostToConditions(&c2)
	assert.NoError(t, err)
	c2ID := idResponse.ID
	assert.Equal(t, c2ID, "2")

	cs, err := alerter.GetFromConditions()
	assert.NoError(t, err)
	assert.Len(t, *cs, 2)
	ok1 := false
	ok2 := false
	for k, _ := range *cs {
		if k == "1" {
			ok1 = true
		}
		if k == "2" {
			ok2 = true
		}
	}
	assert.True(t, ok1 && ok2)

	cond, err := alerter.GetFromCondition("1")
	assert.NoError(t, err)
	assert.NotNil(t, cond)

	err = alerter.DeleteOfCondition("1")
	assert.NoError(t, err)

	cond, err = alerter.GetFromCondition("1")
	assert.Error(t, err) // TODO: should be more refined error here
	assert.Nil(t, cond)

	err = alerter.DeleteOfCondition("2")
	assert.NoError(t, err)

	cs, err = alerter.GetFromConditions()
	assert.NoError(t, err)
	assert.Len(t, *cs, 0)
}

func (suite *AlerterTester) TestEvents() {
	t := suite.T()

	alerter := suite.alerter

	var err error
	var idResponse *client.AlerterIdResponse

	var e1 client.Event
	e1.Type = client.EventDataIngested
	e1.Date = "22 Jan 2016"
	e1.Data = nil
	idResponse, err = alerter.PostToEvents(&e1)
	assert.NoError(t, err)
	e1ID := idResponse.ID
	assert.Equal(t, "E1", e1ID)

	var e2 client.Event
	e2.Type = client.EventDataAccessed
	e2.Date = "22 Jan 2016"
	e2.Data = nil
	idResponse, err = alerter.PostToEvents(&e2)
	assert.NoError(t, err)
	e2ID := idResponse.ID
	assert.Equal(t, "E2", e2ID)

	es, err := alerter.GetFromEvents()
	assert.NoError(t, err)
	assert.Len(t, *es, 2)
	ok1 := false
	ok2 := false
	for k, _ := range *es {
		if k == "E1" {
			ok1 = true
		}
		if k == "E2" {
			ok2 = true
		}
	}
	assert.True(t, ok1 && ok2)
}

func (suite *AlerterTester) TestTriggering() {
	t := suite.T()

	alerter := suite.alerter

	var err error
	var idResponse *client.AlerterIdResponse

	cs, err := alerter.GetFromConditions()
	assert.NoError(t, err)
	assert.Len(t, *cs, 0)

	var rawC3 client.Condition
	rawC3.Title = "cond1 title"
	rawC3.Description = "cond1 descr"
	rawC3.Type = client.EventDataIngested
	rawC3.UserID = "user1"
	rawC3.Date = time.Now().String()
	idResponse, err = alerter.PostToConditions(&rawC3)
	assert.NoError(t, err)
	c3ID := idResponse.ID
	assert.Equal(t, "3", c3ID)

	var rawC4 client.Condition
	rawC4.Title = "cond2 title"
	rawC4.Description = "cond2 descr"
	rawC4.Type = client.EventDataAccessed
	rawC4.UserID = "user2"
	rawC4.Date = time.Now().String()
	idResponse, err = alerter.PostToConditions(&rawC4)
	assert.NoError(t, err)
	c4ID := idResponse.ID
	assert.Equal(t, "4", c4ID)

	var rawC5 client.Condition
	rawC5.Title = "cond2 title"
	rawC5.Description = "cond2 descr"
	rawC5.Type = client.EventFoo
	rawC5.UserID = "user2"
	rawC5.Date = time.Now().String()
	idResponse, err = alerter.PostToConditions(&rawC5)
	assert.NoError(t, err)
	c5ID := idResponse.ID
	assert.Equal(t, "5", c5ID)

	cs, err = alerter.GetFromConditions()
	assert.NoError(t, err)
	assert.Len(t, *cs, 3)

	var e3 client.Event
	e3.Type = client.EventDataAccessed
	e3.Date = time.Now().String()
	e3.Data = map[string]string{"file": "111.tif"}
	idResponse, err = alerter.PostToEvents(&e3)
	assert.NoError(t, err)
	e3ID := idResponse.ID
	assert.Equal(t, "E3", e3ID)

	var e4 client.Event
	e4.Type = client.EventDataIngested
	e4.Date = time.Now().String()
	e4.Data = map[string]string{"file": "111.tif"}
	idResponse, err = alerter.PostToEvents(&e4)
	assert.NoError(t, err)
	e4ID := idResponse.ID
	assert.Equal(t, "E4", e4ID)

	var e5 client.Event
	e5.Type = client.EventBar
	e5.Date = time.Now().String()
	idResponse, err = alerter.PostToEvents(&e5)
	assert.NoError(t, err)
	e5ID := idResponse.ID
	assert.Equal(t, "E5", e5ID)

	as, err := alerter.GetFromAlerts()
	assert.NoError(t, err)
	assert.Len(t, *as, 2)

	// TODO: dependent on order of returned results
	v1, ok := (*as)["A1"]
	assert.True(t, ok)
	v2, ok := (*as)["A2"]
	assert.True(t, ok)

	assert.Equal(t, "A1", v1.ID)
	assert.Equal(t, "E3", v1.EventID)
	assert.Equal(t, "4", v1.ConditionID)
	assert.Equal(t, "A2", v2.ID)
	assert.Equal(t, "E4", v2.EventID)
	assert.Equal(t, "3", v2.ConditionID)
}

func (suite *AlerterTester) TestAdmin() {
	t := suite.T()

	alerter := suite.alerter

	settings, err := alerter.GetFromAdminSettings()
	assert.NoError(t, err)
	if settings.Debug != false {
		t.Error("settings not false")
	}

	settings.Debug = true
	err = alerter.PostToAdminSettings(settings)
	assert.NoError(t, err)

	settings, err = alerter.GetFromAdminSettings()
	assert.NoError(t, err)
	if settings.Debug != true {
		t.Error("settings not true")
	}
}
