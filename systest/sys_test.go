// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workflowsystest

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	"github.com/venicegeo/pz-workflow/workflow"
)

type WorkflowTester struct {
	suite.Suite
	client        *workflow.Client
	url           string
	apiKey        string
	uniq          string
	eventTypeID   piazza.Ident
	eventTypeName string
	triggerName   string
	triggerID     piazza.Ident
	serviceID     piazza.Ident
	eventIDYes    piazza.Ident
	eventIDNo     piazza.Ident
	alertID       piazza.Ident
	jobID         piazza.Ident
	dataID        piazza.Ident
}

var mapType = map[string]interface{}{}
var stringType = "string!"

func (suite *WorkflowTester) setupFixture() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	suite.url = "https://pz-workflow.int.geointservices.io"

	suite.apiKey, err = piazza.GetApiKey("int")
	assert.NoError(err)

	client, err := workflow.NewClient2(suite.url, suite.apiKey)
	assert.NoError(err)
	suite.client = client

	suite.uniq = "systest$" + strconv.Itoa(time.Now().Nanosecond())
	suite.eventTypeName = suite.uniq + "-eventtype"
	suite.triggerName = suite.uniq + "-trigger"
}

func (suite *WorkflowTester) teardownFixture() {
}

func TestRunSuite(t *testing.T) {
	s := &WorkflowTester{}
	suite.Run(t, s)
}

func (suite *WorkflowTester) Test00Init() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	version, err := client.GetVersion()
	assert.NoError(err)
	assert.EqualValues("1.0.0", version.Version)
}

func (suite *WorkflowTester) Test01RegisterService() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	body := map[string]interface{}{
		"url":         "http://pzsvc-hello.int.geointservices.io/hello",
		"contractUrl": "http://pzsvc-hello.int.geointservices.io/contract",
		"method":      "POST",
		"resourceMetadata": map[string]interface{}{
			"name":        "Hello World test",
			"description": "This is the test of Hello World",
			"classType":   "U",
		},
	}

	url := strings.Replace(suite.url, "workflow", "gateway", 1)
	h := piazza.Http{
		BaseUrl: url,
		ApiKey:  suite.apiKey,
		//Preflight:  piazza.SimplePreflight,
		//Postflight: piazza.SimplePostflight,
	}

	obj := map[string]interface{}{}
	code, err := h.Post("/service", body, &obj)
	assert.NoError(err)
	assert.Equal(201, code)
	assert.NotNil(obj)

	assert.IsType(mapType, obj["data"])
	data := obj["data"].(map[string]interface{})
	assert.IsType(stringType, data["serviceId"])
	serviceID := data["serviceId"].(string)
	assert.NotEmpty(serviceID)

	suite.serviceID = piazza.Ident(serviceID)
	log.Printf("ServiceId: %s", suite.serviceID)
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) Test02PostEventType() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	eventType := &workflow.EventType{
		Name: suite.eventTypeName,
		Mapping: map[string]interface{}{
			"alpha": elasticsearch.MappingElementTypeString,
			"beta":  elasticsearch.MappingElementTypeInteger,
		},
	}

	ack, err := client.PostEventType(eventType)
	assert.NoError(err)
	assert.NotNil(ack)

	suite.eventTypeID = ack.EventTypeID
	log.Printf("EventTypeId: %s", suite.eventTypeID)
}

func (suite *WorkflowTester) Test03GetEventType() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	items, err := client.GetAllEventTypes()
	assert.NoError(err)
	assert.True(len(*items) > 1)

	item, err := client.GetEventType(suite.eventTypeID)
	assert.NoError(err)
	assert.NotNil(item)
	assert.EqualValues(suite.eventTypeID, item.EventTypeID)
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) Test04PostTrigger() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	//suite.eventTypeId = "77bbe4c6-b1ac-4bbb-8e86-0f6e6a731c39"
	//suite.serviceId = "61985d9c-d4d0-45d9-a655-7dcf2dc08fad"

	client := suite.client

	trigger := &workflow.Trigger{
		Name:    suite.triggerName,
		Enabled: true,
		Condition: workflow.Condition{
			EventTypeIDs: []piazza.Ident{suite.eventTypeID},
			Query: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"data.beta": 17,
					},
				},
			},
		},
		Job: workflow.JobRequest{
			CreatedBy: "test",
			JobType: workflow.JobType{
				Type: "execute-service",
				Data: map[string]interface{}{
					"dataInputs": map[string]interface{}{
						"": map[string]interface{}{
							"content":  `{"name":"ME", "count":5}`,
							"type":     "body",
							"mimeType": "application/json",
						},
					},
					"dataOutput": [](map[string]interface{}){
						{
							"mimeType": "application/json",
							"type":     "text",
						},
					},
					"serviceId": suite.serviceID,
				},
			},
		},
	}

	ack, err := client.PostTrigger(trigger)
	assert.NoError(err)
	assert.NotNil(ack)

	suite.triggerID = ack.TriggerID
	log.Printf("TriggerId: %s", suite.triggerID)
}

func (suite *WorkflowTester) Test05GetTrigger() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	items, err := client.GetAllTriggers()
	assert.NoError(err)
	assert.True(len(*items) > 1)

	item, err := client.GetTrigger(suite.triggerID)
	assert.NoError(err)
	assert.NotNil(item)
	assert.EqualValues(suite.triggerID, item.TriggerID)
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) Test06PostEvent() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	eventY := &workflow.Event{
		EventTypeID: suite.eventTypeID,
		Data: map[string]interface{}{
			"beta":  17,
			"alpha": "quick brown fox",
		},
	}

	eventN := &workflow.Event{
		EventTypeID: suite.eventTypeID,
		Data: map[string]interface{}{
			"beta":  71,
			"alpha": "lazy dog",
		},
	}

	ack, err := client.PostEvent(eventY)
	assert.NoError(err)
	assert.NotNil(ack)
	suite.eventIDYes = ack.EventID
	log.Printf("EventIdY: %s", suite.eventIDYes)

	ack, err = client.PostEvent(eventN)
	assert.NoError(err)
	assert.NotNil(ack)
	suite.eventIDNo = ack.EventID
	log.Printf("EventIdN: %s", suite.eventIDNo)
}

func (suite *WorkflowTester) Test07GetEvent() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	items, err := client.GetAllEvents()
	assert.NoError(err)
	assert.True(len(*items) > 1)

	item, err := client.GetEvent(suite.eventIDYes)
	assert.NoError(err)
	assert.NotNil(item)
	assert.EqualValues(suite.eventIDYes, item.EventID)
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) Test08PostAlert() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	alert := &workflow.Alert{
		TriggerID: "x",
		EventID:   "y",
		JobID:     "z",
	}

	ack, err := client.PostAlert(alert)
	assert.NoError(err)
	assert.NotNil(ack)
}

func (suite *WorkflowTester) Test09GetAlert() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	items, err := client.GetAllAlerts()
	assert.NoError(err)
	assert.True(len(*items) > 1)

	items, err = client.GetAlertByTrigger(suite.triggerID)
	assert.NoError(err)
	assert.NotNil(items)
	assert.Len(*items, 1)
	assert.EqualValues(suite.eventIDYes, (*items)[0].EventID)

	suite.alertID = (*items)[0].AlertID
	log.Printf("AlertId: %s", suite.alertID)

	suite.jobID = (*items)[0].JobID
	log.Printf("JobId: %s", suite.jobID)
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) Test10GetJob() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	//client := suite.client

	url := strings.Replace(suite.url, "workflow", "gateway", 1)
	h := piazza.Http{
		BaseUrl: url,
		ApiKey:  suite.apiKey,
		//Preflight:  piazza.SimplePreflight,
		//Postflight: piazza.SimplePostflight,
	}

	var data map[string]interface{}

	poll := func() (bool, error) {
		obj := map[string]interface{}{}
		code, err := h.Get("/job/"+string(suite.jobID), &obj)
		if err != nil {
			return false, err
		}
		if code != 200 {
			log.Printf("code is %d", code)
			return false, errors.New("code not 200")
		}
		if obj == nil {
			return false, errors.New("obj was nil")
		}

		var ok bool
		data, ok = obj["data"].(map[string]interface{})
		if !ok {
			return false, errors.New("obj[data] not a map")
		}

		status, ok := data["status"].(string)
		if !ok {
			return false, errors.New("obj[data][status] not a string")
		}

		if status != "Success" {
			return false, nil
		}

		return true, nil
	}

	ok, err := elasticsearch.PollFunction(poll)
	assert.NoError(err)
	assert.True(ok)

	result, ok := data["result"].(map[string]interface{})
	assert.True(ok)
	id, ok := result["dataId"].(string)
	assert.True(ok)

	suite.dataID = piazza.Ident(id)
	log.Printf("DataId: %s", suite.dataID)
}

func (suite *WorkflowTester) Test11GetData() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	//client := suite.client

	url := strings.Replace(suite.url, "workflow", "gateway", 1)
	h := piazza.Http{
		BaseUrl: url,
		ApiKey:  suite.apiKey,
		//Preflight:  piazza.SimplePreflight,
		//Postflight: piazza.SimplePostflight,
	}

	obj := map[string]interface{}{}
	code, err := h.Get("/data/"+string(suite.dataID), &obj)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj)

	var ok bool
	data, ok := obj["data"].(map[string]interface{})
	assert.True(ok)

	dataType, ok := data["dataType"].(map[string]interface{})
	assert.True(ok)
	content, ok := dataType["content"].(string)
	assert.True(ok)

	jsn := `{
		"greeting": "Hello, ME!", 
		"countSquared": 25
	}`
	assert.JSONEq(jsn, content)
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) Test99Admin() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	stats, err := client.GetStats()
	assert.NoError(err)

	assert.NotZero(stats.NumEventTypes)
	assert.NotZero(stats.NumEvents)
	assert.NotZero(stats.NumTriggers)
	assert.NotZero(stats.NumAlerts)
	assert.NotZero(stats.NumTriggeredJobs)
}
