package main

import (
	"encoding/json"
	//"io/ioutil"
	"net/http"
	"testing"
	"time"
	"fmt"
	assert "github.com/stretchr/testify/assert"
	piazza "github.com/venicegeo/pz-gocommon"
	//pztesting "github.com/venicegeo/pz-gocommon/testing"
	"bytes"
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
}
