package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
}

func (suite *AlerterTester) SetupSuite() {
	setup(suite.T(), "12342")
}

func (suite *AlerterTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func TestRunSuite(t *testing.T) {
	s := new(AlerterTester)
	suite.Run(t, s)
}

//---------------------------------------------------------------------------

func setup(t *testing.T, port string) {
	s := fmt.Sprintf("-server localhost:%s -discover localhost:3000", port)

	done := make(chan bool, 1)
	go main2(s, done)
	<-done

	err := pzService.WaitForService(pzService.Name, 1000)
	if err != nil {
		t.Fatal(err)
	}
}

//---------------------------------------------------------------------------

func postEvent(t *testing.T, event Event) string {

	data, err := json.Marshal(event)
	assert.NoError(t, err)

	body := bytes.NewBuffer(data)

	resp, err := http.Post("http://localhost:12342/v1/events", piazza.ContentTypeJSON, body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	d, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var x map[string]string
	err = json.Unmarshal(d, &x)
	assert.NoError(t, err)
	assert.Contains(t, x, "id")

	return x["id"]
}

func getEvents(t *testing.T) []string {
	resp, err := http.Get("http://localhost:12342/v1/events")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	d, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var x map[string]interface{}
	err = json.Unmarshal(d, &x)
	assert.NoError(t, err)

	var a []string
	for k := range x {
		a = append(a, k)
	}
	return a
}

//---------------------------------------------------------------------------

func postCondition(t *testing.T, cond Condition) string {
	data, err := json.Marshal(cond)
	assert.NoError(t, err)

	body := bytes.NewBuffer(data)

	resp, err := http.Post("http://localhost:12342/v1/conditions", piazza.ContentTypeJSON, body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	d, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var x map[string]string
	err = json.Unmarshal(d, &x)
	assert.NoError(t, err)
	assert.Contains(t, x, "id")

	return x["id"]
}

func getCondition(t *testing.T, id string) bool {
	resp, err := http.Get("http://localhost:12342/v1/conditions/" + id)
	assert.NoError(t, err)

	if resp.StatusCode == http.StatusNotFound {
		return false
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	d, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var x Condition
	err = json.Unmarshal(d, &x)
	assert.NoError(t, err)

	assert.Equal(t, id, x.ID)

	return true
}

func getConditions(t *testing.T) []string {
	resp, err := http.Get("http://localhost:12342/v1/conditions")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	d, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var x map[string]interface{}
	err = json.Unmarshal(d, &x)
	assert.NoError(t, err)

	var a []string
	for k := range x {
		a = append(a, k)
	}
	return a
}

func deleteCondition(t *testing.T, id string) {
	resp, err := piazza.Delete("http://localhost:12342/v1/conditions/" + id)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func getAlerts(t *testing.T) []Alert {
	resp, err := http.Get("http://localhost:12342/v1/alerts")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	d, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var x map[string]Alert
	err = json.Unmarshal(d, &x)
	assert.NoError(t, err)

	var a []Alert
	for i := range x {
		a = append(a, x[i])
	}
	return a
}

//---------------------------------------------------------------------------

func (suite *AlerterTester) TestConditions() {
	t := suite.T()

	var c1 Condition
	c1.Title = "c1"
	c1.Type = "=type="
	c1.UserID = "=userid="
	c1.Date = "=date="
	c1ID := postCondition(t, c1)
	assert.Equal(t, c1ID, "1")

	var c2 Condition
	c2.Title = "c2"
	c2.Type = "=type="
	c2.UserID = "=userid="
	c2.Date = "=date="
	c2ID := postCondition(t, c2)
	assert.Equal(t, c2ID, "2")

	cs := getConditions(t)
	assert.Len(t, cs, 2)
	assert.Contains(t, cs, "1")
	assert.Contains(t, cs, "2")

	ok := getCondition(t, "1")
	assert.True(t, ok)

	deleteCondition(t, "1")

	ok = getCondition(t, "1")
	assert.True(t, !ok)

	deleteCondition(t, "2")

	cs = getConditions(t)
	assert.Len(t, cs, 0)
}

func (suite *AlerterTester) TestEvents() {
	t := suite.T()

	var e1 Event
	e1.Type = EventDataIngested
	e1.Date = "22 Jan 2016"
	e1.Data = nil
	e1ID := postEvent(t, e1)
	assert.Equal(t, "E1", e1ID)

	var e2 Event
	e2.Type = EventDataAccessed
	e2.Date = "22 Jan 2016"
	e2.Data = nil
	e2ID := postEvent(t, e2)
	assert.Equal(t, "E2", e2ID)

	es := getEvents(t)
	assert.Len(t, es, 2)
	assert.Contains(t, es, "E1")
	assert.Contains(t, es, "E2")
}

func (suite *AlerterTester) TestTriggering() {
	t := suite.T()

	cs := getConditions(t)
	assert.Len(t, cs, 0)

	var rawC3 Condition
	rawC3.Title = "cond1 title"
	rawC3.Description = "cond1 descr"
	rawC3.Type = EventDataIngested
	rawC3.UserID = "user1"
	rawC3.Date = time.Now().String()
	c3ID := postCondition(t, rawC3)
	assert.Equal(t, "3", c3ID)

	var rawC4 Condition
	rawC4.Title = "cond2 title"
	rawC4.Description = "cond2 descr"
	rawC4.Type = EventDataAccessed
	rawC4.UserID = "user2"
	rawC4.Date = time.Now().String()
	c4ID := postCondition(t, rawC4)
	assert.Equal(t, "4", c4ID)

	var rawC5 Condition
	rawC5.Title = "cond2 title"
	rawC5.Description = "cond2 descr"
	rawC5.Type = EventFoo
	rawC5.UserID = "user2"
	rawC5.Date = time.Now().String()
	c5ID := postCondition(t, rawC5)
	assert.Equal(t, "5", c5ID)

	cs = getConditions(t)
	assert.Len(t, cs, 3)

	var e3 Event
	e3.Type = EventDataAccessed
	e3.Date = time.Now().String()
	e3.Data = map[string]string{"file": "111.tif"}
	e3ID := postEvent(t, e3)
	assert.Equal(t, "E3", e3ID)

	var e4 Event
	e4.Type = EventDataIngested
	e4.Date = time.Now().String()
	e4.Data = map[string]string{"file": "111.tif"}
	e4ID := postEvent(t, e4)
	assert.Equal(t, "E4", e4ID)

	var e5 Event
	e5.Type = EventBar
	e5.Date = time.Now().String()
	e5ID := postEvent(t, e5)
	assert.Equal(t, "E5", e5ID)

	as := getAlerts(t)
	assert.Len(t, as, 2)

	// TODO: dependent on order of returned results
	if as[0].ID == "A1" {
		assert.Equal(t, "A1", as[0].ID)
		assert.Equal(t, "E3", as[0].EventID)
		assert.Equal(t, "4", as[0].ConditionID)
		assert.Equal(t, "A2", as[1].ID)
		assert.Equal(t, "E4", as[1].EventID)
		assert.Equal(t, "3", as[1].ConditionID)
	} else {
		assert.Equal(t, "A2", as[0].ID)
		assert.Equal(t, "E4", as[0].EventID)
		assert.Equal(t, "3", as[0].ConditionID)
		assert.Equal(t, "A1", as[1].ID)
		assert.Equal(t, "E3", as[1].EventID)
		assert.Equal(t, "4", as[1].ConditionID)
	}
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
