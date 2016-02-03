package main

import (
	"bytes"
	"encoding/json"
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	piazza "github.com/venicegeo/pz-gocommon"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type AlerterTester struct {
	suite.Suite
	client *PzAlerterClient
}

func (suite *AlerterTester) SetupSuite() {
	t := suite.T()

	done := make(chan bool, 1)
	go Main(done, true)
	<-done

	err := pzService.WaitForService(pzService.Name, 1000)
	if err != nil {
		t.Fatal(err)
	}

	suite.client = NewPzAlerterClient("localhost:12342")
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

	client := suite.client

	var err error
	var idResponse *piazza.AlerterIdResponse

	var c1 piazza.Condition
	c1.Title = "c1"
	c1.Type = "=type="
	c1.UserID = "=userid="
	c1.Date = "=date="
	idResponse, err = client.PostToConditions(&c1)
	assert.NoError(t, err)
	c1ID := idResponse.ID
	assert.Equal(t, c1ID, "1")

	var c2 piazza.Condition
	c2.Title = "c2"
	c2.Type = "=type="
	c2.UserID = "=userid="
	c2.Date = "=date="
	idResponse, err = client.PostToConditions(&c2)
	assert.NoError(t, err)
	c2ID := idResponse.ID
	assert.Equal(t, c2ID, "2")

	cs, err := client.GetFromConditions()
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

	cond, err := client.GetFromCondition("1")
	assert.NoError(t, err)
	assert.NotNil(t, cond)

	err = client.DeleteOfCondition("1")
	assert.NoError(t, err)

	cond, err = client.GetFromCondition("1")
	assert.Error(t, err) // TODO: should be more refined error here
	assert.Nil(t, cond)

	err = client.DeleteOfCondition("2")
	assert.NoError(t, err)

	cs, err = client.GetFromConditions()
	assert.NoError(t, err)
	assert.Len(t, *cs, 0)
}

func (suite *AlerterTester) TestEvents() {
	t := suite.T()

	client := suite.client

	var err error
	var idResponse *piazza.AlerterIdResponse

	var e1 piazza.Event
	e1.Type = piazza.EventDataIngested
	e1.Date = "22 Jan 2016"
	e1.Data = nil
	idResponse, err = client.PostToEvents(&e1)
	assert.NoError(t, err)
	e1ID := idResponse.ID
	assert.Equal(t, "E1", e1ID)

	var e2 piazza.Event
	e2.Type = piazza.EventDataAccessed
	e2.Date = "22 Jan 2016"
	e2.Data = nil
	idResponse, err = client.PostToEvents(&e2)
	assert.NoError(t, err)
	e2ID := idResponse.ID
	assert.Equal(t, "E2", e2ID)

	es, err := client.GetFromEvents()
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

	client := suite.client

	var err error
	var idResponse *piazza.AlerterIdResponse

	cs, err := client.GetFromConditions()
	assert.NoError(t, err)
	assert.Len(t, *cs, 0)

	var rawC3 piazza.Condition
	rawC3.Title = "cond1 title"
	rawC3.Description = "cond1 descr"
	rawC3.Type = piazza.EventDataIngested
	rawC3.UserID = "user1"
	rawC3.Date = time.Now().String()
	idResponse, err = client.PostToConditions(&rawC3)
	assert.NoError(t, err)
	c3ID := idResponse.ID
	assert.Equal(t, "3", c3ID)

	var rawC4 piazza.Condition
	rawC4.Title = "cond2 title"
	rawC4.Description = "cond2 descr"
	rawC4.Type = piazza.EventDataAccessed
	rawC4.UserID = "user2"
	rawC4.Date = time.Now().String()
	idResponse, err = client.PostToConditions(&rawC4)
	assert.NoError(t, err)
	c4ID := idResponse.ID
	assert.Equal(t, "4", c4ID)

	var rawC5 piazza.Condition
	rawC5.Title = "cond2 title"
	rawC5.Description = "cond2 descr"
	rawC5.Type = piazza.EventFoo
	rawC5.UserID = "user2"
	rawC5.Date = time.Now().String()
	idResponse, err = client.PostToConditions(&rawC5)
	assert.NoError(t, err)
	c5ID := idResponse.ID
	assert.Equal(t, "5", c5ID)

	cs, err = client.GetFromConditions()
	assert.NoError(t, err)
	assert.Len(t, *cs, 3)

	var e3 piazza.Event
	e3.Type = piazza.EventDataAccessed
	e3.Date = time.Now().String()
	e3.Data = map[string]string{"file": "111.tif"}
	idResponse, err = client.PostToEvents(&e3)
	assert.NoError(t, err)
	e3ID := idResponse.ID
	assert.Equal(t, "E3", e3ID)

	var e4 piazza.Event
	e4.Type = piazza.EventDataIngested
	e4.Date = time.Now().String()
	e4.Data = map[string]string{"file": "111.tif"}
	idResponse, err = client.PostToEvents(&e4)
	assert.NoError(t, err)
	e4ID := idResponse.ID
	assert.Equal(t, "E4", e4ID)

	var e5 piazza.Event
	e5.Type = piazza.EventBar
	e5.Date = time.Now().String()
	idResponse, err = client.PostToEvents(&e5)
	assert.NoError(t, err)
	e5ID := idResponse.ID
	assert.Equal(t, "E5", e5ID)

	as, err := client.GetFromAlerts()
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

	resp, err := http.Get("http://localhost:12342/v1/admin/settings")
	if err != nil {
		t.Fatalf("admin settings get failed: %s", err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	sm := map[string]string{}
	err = json.Unmarshal(data, &sm)
	if err != nil {
		t.Fatalf("admin settings get failed: %s", err)
	}
	if sm["debug"] != "false" {
		t.Error("settings get had invalid response")
	}

	m := map[string]string{"debug": "true"}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("admin settings %s", err)
	}
	resp, err = http.Post("http://localhost:12342/v1/admin/settings", "application/json", bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("admin settings post failed: %s", err)
	}

	resp, err = http.Get("http://localhost:12342/v1/admin/settings")
	if err != nil {
		t.Fatalf("admin settings get failed: %s", err)
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	sm = map[string]string{}
	err = json.Unmarshal(data, &sm)
	if err != nil {
		t.Fatalf("admin settings get failed: %s", err)
	}
	if sm["debug"] != "true" {
		t.Error("settings get had invalid response")
	}

}
