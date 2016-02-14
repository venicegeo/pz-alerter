package main

import (
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-alerter/client"
	"github.com/venicegeo/pz-alerter/server"
	"github.com/venicegeo/pz-gocommon"
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

	suite.assertNoData()
}

func (suite *AlerterTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func (suite *AlerterTester) assertNoData() {
	t := suite.T()

	var err error

	cs, err := suite.alerter.GetFromConditions()
	assert.NoError(t, err)
	assert.Len(t, *cs, 0)

	es, err := suite.alerter.GetFromEvents()
	assert.NoError(t, err)
	assert.Len(t, *es, 0)

	as, err := suite.alerter.GetFromAlerts()
	assert.NoError(t, err)
	assert.Len(t, *as, 0)

	xs, err := suite.alerter.GetFromActions()
	assert.NoError(t, err)
	assert.Len(t, *xs, 0)
}

func TestRunSuite(t *testing.T) {
	s := new(AlerterTester)
	suite.Run(t, s)
}

//---------------------------------------------------------------------------

func (suite *AlerterTester) TestAAResource() {
	t := suite.T()
	assert := assert.New(t)
	//alerter := suite.alerter

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
	err = db.GetById(a2Id, &t2)
	assert.NoError(err)
	assert.EqualValues(a2.Type, t2.Type)

	var t1 client.Event
	err = db.GetById(a1Id, &t1)
	assert.NoError(err)
	assert.EqualValues(a1.Type, t1.Type)
}

func (suite *AlerterTester) TestAlerts() {
	t := suite.T()
	assert := assert.New(t)

	suite.assertNoData()

	alerter := suite.alerter

	var err error
	var idResponse *client.AlerterIdResponse

	var a1 client.Alert
	a1.ActionId = "this is action 1"
	idResponse, err = alerter.PostToAlerts(&a1)
	assert.NoError(err)
	a1ID := idResponse.ID
	assert.EqualValues("A1", a1ID)

	var a2 client.Alert
	a2.ActionId = "this is action 2"
	idResponse, err = alerter.PostToAlerts(&a2)
	assert.NoError(err)
	a2ID := idResponse.ID
	assert.EqualValues("A2", a2ID)

	as, err := alerter.GetFromAlerts()
	assert.NoError(err)
	assert.Len(*as, 2)
	ok1 := false
	ok2 := false
	for k := range *as {
		if k == "A1" {
			ok1 = true
		}
		if k == "A2" {
			ok2 = true
		}
	}

	assert.True(ok1 && ok2)
	alert, err := alerter.GetFromAlert("A1")
	assert.NoError(err)
	assert.NotNil(alert)

	err = alerter.DeleteOfAlert("A1")
	assert.NoError(err)

	alert, err = alerter.GetFromAlert("A1")
	assert.Error(err)
	assert.Nil(alert)

	err = alerter.DeleteOfAlert("A2")
	assert.NoError(err)

	as, err = alerter.GetFromAlerts()
	assert.NoError(err)
	assert.Len(*as, 0)

	suite.assertNoData()
}

func (suite *AlerterTester) TestConditions() {
	t := suite.T()
	assert := assert.New(t)

	suite.assertNoData()

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

	suite.assertNoData()
}

func (suite *AlerterTester) TestActions() {
	t := suite.T()
	assert := assert.New(t)

	suite.assertNoData()

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

	tmp, err := alerter.GetFromAction("X1")
	assert.NoError(err)
	assert.NotNil(tmp)

	err = alerter.DeleteOfAction("X1")
	assert.NoError(err)
	err = alerter.DeleteOfAction("X2")
	assert.NoError(err)

	suite.assertNoData()
}

func (suite *AlerterTester) TestEvents() {
	t := suite.T()
	assert := assert.New(t)

	alerter := suite.alerter
	suite.assertNoData()

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

	err = alerter.DeleteOfEvent("E1")
	assert.NoError(err)
	err = alerter.DeleteOfEvent("E2")
	assert.NoError(err)

	suite.assertNoData()
}

func (suite *AlerterTester) TestTriggering() {

	t := suite.T()
	assert := assert.New(t)

	alerter := suite.alerter

	suite.assertNoData()

	var err error
	var idResponse *client.AlerterIdResponse

	////////////////
	var c1 client.Condition
	c1.Title = "c1 title"
	c1.Type = client.EventFoo
	c1.Query = "c1 query"
	idResponse, err = alerter.PostToConditions(&c1)
	assert.NoError(err)
	c1Id := idResponse.ID

	var c2 client.Condition
	c2.Title = "c2 title"
	c2.Type = client.EventBar
	c2.Query = "c2 query"
	idResponse, err = alerter.PostToConditions(&c2)
	assert.NoError(err)
	c2Id := idResponse.ID

	var c3 client.Condition
	c3.Title = "c3 title"
	c3.Type = client.EventBaz
	c3.Query = "c3 query"
	idResponse, err = alerter.PostToConditions(&c3)
	assert.NoError(err)
	c3Id := idResponse.ID

	cs, err := alerter.GetFromConditions()
	assert.NoError(err)
	assert.Len(*cs, 3)

	//////////////////////////
	var x1 client.Action
	x1.Conditions = []client.Ident{c2Id}
	x1.Job = "action 1 job"
	idResponse, err = alerter.PostToActions(&x1)
	assert.NoError(err)
	x1Id := idResponse.ID

	var x2 client.Action
	x2.Conditions = []client.Ident{c1Id}
	x2.Job = "action 2 job"
	idResponse, err = alerter.PostToActions(&x2)
	assert.NoError(err)
	x2Id := idResponse.ID

	var x3 client.Action
	x3.Conditions = []client.Ident{c3Id}
	x3.Job = "action 3 job"
	idResponse, err = alerter.PostToActions(&x3)
	assert.NoError(err)
	x3Id := idResponse.ID

	/////////////////////

	// will cause action X2
	var e1 client.Event
	e1.Type = client.EventFoo
	e1.Date = time.Now()
	e1.Data = map[string]string{"file": "e1.tif"}
	idResponse, err = alerter.PostToEvents(&e1)
	assert.NoError(err)
	e1Id := idResponse.ID

	// will cause action X1
	var e2 client.Event
	e2.Type = client.EventBar
	e2.Date = time.Now()
	e2.Data = map[string]string{"file": "e2.tif"}
	idResponse, err = alerter.PostToEvents(&e2)
	assert.NoError(err)
	e2Id := idResponse.ID

	// will cause no actions
	var e3 client.Event
	e3.Type = client.EventBuz
	e3.Date = time.Now()
	idResponse, err = alerter.PostToEvents(&e3)
	assert.NoError(err)
	e3Id := idResponse.ID

	////////////////////////////////////

	as, err := alerter.GetFromAlerts()
	assert.NoError(err)
	assert.Len(*as, 2)

	alerts := as.ToSortedArray()
	assert.Len(alerts, 2)

	t1 := (alerts[0].ActionId == x1Id)
	t2 := (alerts[1].ActionId == x2Id)
	t3 := (alerts[0].ActionId == x2Id)
	t4 := (alerts[1].ActionId == x1Id)
	assert.True((t1 && t2) || (t3 && t4))

	//////////////

	alerter.DeleteOfCondition(c1Id)
	alerter.DeleteOfCondition(c2Id)
	alerter.DeleteOfCondition(c3Id)
	alerter.DeleteOfAction(x1Id)
	alerter.DeleteOfAction(x2Id)
	alerter.DeleteOfAction(x3Id)
	alerter.DeleteOfEvent(e1Id)
	alerter.DeleteOfEvent(e2Id)
	alerter.DeleteOfEvent(e3Id)
	alerter.DeleteOfAlert(alerts[0].ID)
	alerter.DeleteOfAlert(alerts[1].ID)

	suite.assertNoData()
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
