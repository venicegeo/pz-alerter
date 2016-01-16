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

func TestOkay(t *testing.T) {
	setup("12342")

	m := Alert{Name: "n", Condition: "c"}

	data, err := json.Marshal(m)
	assert.NoError(t, err)

	body := bytes.NewBuffer(data)

	resp, err := http.Post("http://localhost:12342/alerts", piazza.ContentTypeJSON, body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var s []byte
	s, _ = ioutil.ReadAll(resp.Body)
	t.Log(string(s))

	body = bytes.NewBuffer(data)
	resp, err = http.Post("http://localhost:12342/alerts", piazza.ContentTypeJSON, body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	s, _ = ioutil.ReadAll(resp.Body)
	t.Log(string(s))

	resp, err = http.Get("http://localhost:12342/alerts")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	s, _ = ioutil.ReadAll(resp.Body)
	t.Log(string(s))

	resp, err = http.Get("http://localhost:12342/alerts/1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	s, _ = ioutil.ReadAll(resp.Body)
	t.Log(string(s))

	resp, err = piazza.Delete("http://localhost:12342/alerts/1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get("http://localhost:12342/alerts/1")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	resp, err = piazza.Delete("http://localhost:12342/alerts/2")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get("http://localhost:12342/alerts")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	s, _ = ioutil.ReadAll(resp.Body)
	e := make([]byte, 2)
	e[0] = 0x7b ; e[1] = 0x7d
	assert.Equal(t, e, s[:2])
	t.Log(s)
}
