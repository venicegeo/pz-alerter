package main

import (
	"encoding/json"
	//"io/ioutil"
	"fmt"
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	piazza "github.com/venicegeo/pz-gocommon"
	"net/http"
	"testing"
	"time"
	//pztesting "github.com/venicegeo/pz-gocommon/testing"
	"bytes"
	"io/ioutil"
)

type AlerterTester struct {
	suite.Suite
}

func (suite *AlerterTester) SetupSuite() {
	setup("12342")
}

func (suite *AlerterTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func TestRunSuite(t *testing.T) {
	s := new(AlerterTester)
	suite.Run(t, s)
}

//---------------------------------------------------------------------------

func setup(port string) {
	s := fmt.Sprintf("-discovery http://localhost:3000 -port %s", port)

	go main2(s)

	time.Sleep(250 * time.Millisecond)
}

//---------------------------------------------------------------------------

func makeEvent(t *testing.T, name string) string {
	m := Event{Type: "=type=", Date: "=date="}
	return makeRawEvent(t, m)
}

func makeRawEvent(t *testing.T, event Event) string {

	data, err := json.Marshal(event)
	assert.NoError(t, err)

	body := bytes.NewBuffer(data)

	resp, err := http.Post("http://localhost:12342/events", piazza.ContentTypeJSON, body)
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
	resp, err := http.Get("http://localhost:12342/events")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	d, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var x map[string]interface{}
	err = json.Unmarshal(d, &x)
	assert.NoError(t, err)

	a := make([]string, 0)
	for k := range x {
		a = append(a, k)
	}
	return a
}

//---------------------------------------------------------------------------

func makeCondition(t *testing.T, title string) string {
	m := Condition{
		Title:     title,
		Type:      "=type=",
		UserID:    "=userid=",
		Date: "=date="}

	return makeRawCondition(t, m)
}

func makeRawCondition(t *testing.T, cond Condition) string {
	data, err := json.Marshal(cond)
	assert.NoError(t, err)

	body := bytes.NewBuffer(data)

	resp, err := http.Post("http://localhost:12342/conditions", piazza.ContentTypeJSON, body)
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
	resp, err := http.Get("http://localhost:12342/conditions/" + id)
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
	resp, err := http.Get("http://localhost:12342/conditions")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	d, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var x map[string]interface{}
	err = json.Unmarshal(d, &x)
	assert.NoError(t, err)

	a := make([]string, 0)
	for k := range x {
		a = append(a, k)
	}
	return a
}

func deleteCondition(t *testing.T, id string) {
	resp, err := piazza.Delete("http://localhost:12342/conditions/" + id)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func getAlerts(t *testing.T) []Alert {
	resp, err := http.Get("http://localhost:12342/alerts")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	d, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var x map[string]Alert
	err = json.Unmarshal(d, &x)
	assert.NoError(t, err)

	a := make([]Alert, 0)
	for i := range x {
		a = append(a, x[i])
	}
	return a
}

//---------------------------------------------------------------------------

func (suite *AlerterTester) TestConditions() {
	t := suite.T()

	c1 := makeCondition(t, "c1")
	assert.Equal(t, "1", c1)

	c2 := makeCondition(t, "c2")
	assert.Equal(t, "2", c2)

	cs := getConditions(t)
	assert.Len(t, cs, 2)
	assert.Contains(t, cs, c1)
	assert.Contains(t, cs, c2)

	ok := getCondition(t, c1)
	assert.True(t, ok)

	deleteCondition(t, c1)

	ok = getCondition(t, c1)
	assert.True(t, !ok)

	deleteCondition(t, c2)

	cs = getConditions(t)
	assert.Len(t, cs, 0)
}

func (suite *AlerterTester) TestEvents() {
	t := suite.T()

	e1 := makeEvent(t, "e1")
	assert.Equal(t, "1", e1)

	e2 := makeEvent(t, "e2")
	assert.Equal(t, "2", e2)

	es := getEvents(t)
	assert.Len(t, es, 2)
	assert.Contains(t, es, e1)
	assert.Contains(t, es, e2)
}

func (suite *AlerterTester) TestTriggering() {
	t := suite.T()

	rawC1 := Condition{
		Title:       "cond1 title",
		Description: "cond1 descr",
		Type:        EventDataIngested,
		UserID:      "user1",
		Date:   time.Now().String(),
	}

	rawC2 := Condition{
		Title:       "cond2 title",
		Description: "cond2 descr",
		Type:        EventDataAccessed,
		UserID:      "user2",
		Date:   time.Now().String(),
	}

	rawC3 := Condition{
		Title:       "cond2 title",
		Description: "cond2 descr",
		Type:        EventFoo,
		UserID:      "user2",
		Date:   time.Now().String(),
	}

	c1 := makeRawCondition(t, rawC1)
	assert.Equal(t, "3", c1)

	c2 := makeRawCondition(t, rawC2)
	assert.Equal(t, "4", c2)

	c3 := makeRawCondition(t, rawC3)
	assert.Equal(t, "5", c3)

	rawE1 := Event{
		Type: EventDataAccessed,
		Date: time.Now().String(),
		Data: map[string]string{"file": "111.tif"},
	}

	rawE2 := Event{
		Type: EventDataIngested,
		Date: time.Now().String(),
		Data: map[string]string{"file": "111.tif"},
	}

	rawE3 := Event{
		Type: EventBar,
		Date: time.Now().String(),
	}

	e1 := makeRawEvent(t, rawE1)
	assert.Equal(t, "3", e1)

	e2 := makeRawEvent(t, rawE2)
	assert.Equal(t, "4", e2)

	e3 := makeRawEvent(t, rawE3)
	assert.Equal(t, "5", e3)

	as := getAlerts(t)
	assert.Len(t, as, 2)
	assert.Equal(t, "1", as[0].ID)
	assert.Equal(t, "3", as[0].EventID)
	assert.Equal(t, "4", as[0].ConditionID)
	assert.Equal(t, "2", as[1].ID)
	assert.Equal(t, "4", as[1].EventID)
	assert.Equal(t, "3", as[1].ConditionID)
}
