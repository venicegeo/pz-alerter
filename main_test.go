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

	es, err := suite.alerter.GetFromEvents()
	assert.NoError(t, err)
	assert.Len(t, *es, 0)

	as, err := suite.alerter.GetFromAlerts()
	assert.NoError(t, err)
	assert.Len(t, *as, 0)

	xs, err := suite.alerter.GetFromTriggers()
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

func (suite *AlerterTester) TestAlerts() {
	t := suite.T()
	assert := assert.New(t)

	suite.assertNoData()

	alerter := suite.alerter

	var err error
	var idResponse *client.AlerterIdResponse

	var a1 client.Alert
	a1.TriggerId = "this is trigger 1"
	idResponse, err = alerter.PostToAlerts(&a1)
	assert.NoError(err)
	a1ID := idResponse.ID
	assert.EqualValues("A1", a1ID)

	var a2 client.Alert
	a2.TriggerId = "this is trigger 2"
	idResponse, err = alerter.PostToAlerts(&a2)
	assert.NoError(err)
	a2ID := idResponse.ID
	assert.EqualValues("A2", a2ID)

	as, err := alerter.GetFromAlerts()
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


func (suite *AlerterTester) TestTriggers() {
	t := suite.T()
	assert := assert.New(t)

	suite.assertNoData()

	alerter := suite.alerter

	var err error
	var idResponse *client.AlerterIdResponse

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
	idResponse, err = alerter.PostToTriggers(&x1)
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
	idResponse, err = alerter.PostToTriggers(&x2)
	assert.NoError(err)
	x2Id := idResponse.ID

	cs, err := alerter.GetFromTriggers()
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

	tmp, err := alerter.GetFromTrigger(x1Id)
	assert.NoError(err)
	assert.NotNil(tmp)

	err = alerter.DeleteOfTrigger(x1Id)
	assert.NoError(err)
	err = alerter.DeleteOfTrigger(x2Id)
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
	for _, v := range *es {
		if v.ID == "E1" {
			ok1 = true
		}
		if v.ID == "E2" {
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
	idResponse, err = alerter.PostToTriggers(&x1)
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
	idResponse, err = alerter.PostToTriggers(&x2)
	assert.NoError(err)
	x2Id := idResponse.ID

	xs, err := alerter.GetFromTriggers()
	assert.NoError(err)
	assert.Len(*xs, 2)

	/////////////////////

	// will cause trigger X1
	var e1 client.Event
	e1.Type = client.EventFoo
	e1.Date = time.Now()
	e1.Data = map[string]string{"file": "e1.tif"}
	idResponse, err = alerter.PostToEvents(&e1)
	assert.NoError(err)
	e1Id := idResponse.ID

	// will cause trigger X2
	var e2 client.Event
	e2.Type = client.EventBar
	e2.Date = time.Now()
	e2.Data = map[string]string{"file": "e2.tif"}
	idResponse, err = alerter.PostToEvents(&e2)
	assert.NoError(err)
	e2Id := idResponse.ID

	// will cause no triggers
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

	alerter.DeleteOfTrigger(x1Id)
	alerter.DeleteOfTrigger(x2Id)
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
