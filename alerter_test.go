package main

import (
	"encoding/json"
	//"io/ioutil"
	"fmt"
	assert "github.com/stretchr/testify/assert"
	piazza "github.com/venicegeo/pz-gocommon"
	"net/http"
	"testing"
	"time"
	//pztesting "github.com/venicegeo/pz-gocommon/testing"
	"bytes"
	"io/ioutil"
)

// @TODO: need to automate call to setup() and/or kill thread after each test
func setup(port string) {
	s := fmt.Sprintf("-discovery http://localhost:3000 -port %s", port)

	go main2(s)

	time.Sleep(250 * time.Millisecond)
}

//---------------------------------------------------------------------------

func makeEvent(t *testing.T, name string) string {
	m := Event{Condition: "c"}

	data, err := json.Marshal(m)
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
	for k := range(x)  {
		a = append(a, k)
	}
	return a
}

//---------------------------------------------------------------------------

func makeCondition(t *testing.T, name string) string {
	m := Condition{Name: name, Condition: "c"}

	data, err := json.Marshal(m)
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

	var x map[string]string
	err = json.Unmarshal(d, &x)
	assert.NoError(t, err)

	assert.Equal(t, id, x["id"])

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
	for k := range(x)  {
		a = append(a, k)
	}
	return a
}

func deleteCondition(t *testing.T, id string) {
	resp, err := piazza.Delete("http://localhost:12342/conditions/" + id)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestOkay(t *testing.T) {
	setup("12342")

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

//
	e1 := makeEvent(t, "e1")
	assert.Equal(t, "1", e1)

	e2 := makeEvent(t, "e2")
	assert.Equal(t, "2", e2)

	es := getEvents(t)
	assert.Len(t, es, 2)
	assert.Contains(t, es, e1)
	assert.Contains(t, es, e2)
}
