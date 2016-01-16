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

func makeAlert(t *testing.T, name string) string {
	m := Alert{Name: name, Condition: "c"}

	data, err := json.Marshal(m)
	assert.NoError(t, err)

	body := bytes.NewBuffer(data)

	resp, err := http.Post("http://localhost:12342/alerts", piazza.ContentTypeJSON, body)
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

func getAlert(t *testing.T, id string) bool {
	resp, err := http.Get("http://localhost:12342/alerts/" + id)
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

func getAlerts(t *testing.T) []string {
	resp, err := http.Get("http://localhost:12342/alerts")
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

func deleteAlert(t *testing.T, id string) {
	resp, err := piazza.Delete("http://localhost:12342/alerts/" + id)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestOkay(t *testing.T) {
	setup("12342")

	a1 := makeAlert(t, "a")
	assert.Equal(t, "1", a1)

	a2 := makeAlert(t, "b")
	assert.Equal(t, "2", a2)

	as := getAlerts(t)
	assert.Len(t, as, 2)
	assert.Contains(t, as, a1)
	assert.Contains(t, as, a2)

	ok := getAlert(t, a1)
	assert.True(t, ok)

	deleteAlert(t, a1)

	ok = getAlert(t, a1)
	assert.True(t, !ok)

	deleteAlert(t, a2)

	as = getAlerts(t)
	assert.Len(t, as, 0)

//
	e1 := makeEvent(t, "x")
	assert.Equal(t, "1", e1)

	e2 := makeEvent(t, "y")
	assert.Equal(t, "2", e2)

	es := getEvents(t)
	assert.Len(t, es, 2)
	assert.Contains(t, es, e1)
	assert.Contains(t, es, e2)
}
