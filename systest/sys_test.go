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
	//"errors"
	"fmt"
	"log"
	"strconv"
	//"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	//pzsyslog "github.com/venicegeo/pz-gocommon/syslog"
)

type WorkflowTester struct {
	suite.Suite
	url           string
	apiKey        string
	apiServer     string
	uniq          string
	eventTypeID   piazza.Ident
	eventTypeName string
	triggerName   string
	triggerID     piazza.Ident
	serviceID     piazza.Ident
	eventIDYes    piazza.Ident
	eventIDNo     piazza.Ident
	repeatID      piazza.Ident
	alertID       piazza.Ident
	jobID         piazza.Ident
	dataID        piazza.Ident
}

var mapType = map[string]interface{}{}
var listType = []interface{}{}
var stringType = "string!"

const goodBeta = 17
const goodAlpha = "quick brown fox"

func (suite *WorkflowTester) setupFixture() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	suite.apiServer, err = piazza.GetApiServer()
	assert.NoError(err)

	suite.url, err = piazza.GetPiazzaUrl()
	assert.NoError(err)

	suite.apiKey, err = piazza.GetApiKey(suite.apiServer)
	assert.NoError(err)

	//logWriter := &pzsyslog.NilWriter{}
	//auditWriter := &pzsyslog.NilWriter{}
	//logger := pzsyslog.NewLogger(logWriter, auditWriter, "pz-workflow/systest")

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

func (suite *WorkflowTester) postToGateway(endpoint string, body map[string]interface{}, obj map[string]interface{}) (int, error) {
	h := piazza.Http{
		BaseUrl: suite.url,
		ApiKey:  suite.apiKey,
		//Preflight:  piazza.SimplePreflight,
		//Postflight: piazza.SimplePostflight,
	}
	return h.Post(endpoint, body, obj)
}

func (suite *WorkflowTester) getFromGateway(endpoint string, obj *map[string]interface{}) (int, error) {
	h := piazza.Http{
		BaseUrl: suite.url,
		ApiKey:  suite.apiKey,
		//Preflight:  piazza.SimplePreflight,
		//Postflight: piazza.SimplePostflight,
	}
	return h.Get(endpoint, obj)
}

func (suite *WorkflowTester) Test01RegisterService() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	helloUrl, err := piazza.GetPiazzaServiceUrl(piazza.PzsvcHello)
	assert.NoError(err)

	body := map[string]interface{}{
		"url":            helloUrl + "/hello",
		"contractUrl":    helloUrl + "/contract",
		"method":         "POST",
		"isAsynchronous": "false",
		"resourceMetadata": map[string]interface{}{
			"name":        "Hello World test",
			"description": "This is the test of Hello World",
			"classType": map[string]interface{}{
				"classification": "UNCLASSIFIED",
			},
		},
	}

	fmt.Println("URL", suite.url)
	obj := map[string]interface{}{}
	code, err := suite.postToGateway("/service", body, &obj)
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

	body := map[string]interface{}{
		"name": suite.eventTypeName,
		"mapping": map[string]interface{}{
			"alpha": elasticsearch.MappingElementTypeString,
			"beta":  elasticsearch.MappingElementTypeInteger,
		},
	}
	obj := map[string]interface{}{}
	code, err := suite.postToGateway("/eventType", body, &obj)
	assert.NoError(err)
	assert.Equal(201, code)
	assert.NotNil(obj)

	assert.IsType(mapType, obj["data"])
	data := obj["data"].(map[string]interface{})
	assert.IsType(stringType, data["eventTypeId"])
	eventTypeID := data["eventTypeId"].(string)
	assert.NotEmpty(eventTypeID)

	suite.eventTypeID = piazza.Ident(eventTypeID)
	log.Printf("EventTypeId: %s", suite.eventTypeID)
}

func (suite *WorkflowTester) Test03GetEventType() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	obj := map[string]interface{}{}
	code, err := suite.getFromGateway("/eventType?size=100", &obj)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj)

	assert.IsType(listType, obj["data"])
	data := obj["data"].([]interface{})
	assert.True(len(data) > 1)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	obj2 := map[string]interface{}{}
	code, err = suite.postToGateway("/eventType/query", query, &obj2)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj2)

	assert.IsType(listType, obj2["data"])
	data2 := obj2["data"].([]interface{})
	assert.True(len(data2) > 1)

	obj3 := map[string]interface{}{}
	code, err = suite.getFromGateway("/eventType/"+string(suite.eventTypeID), &obj3)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj3)

	assert.IsType(mapType, obj3["data"])
	data3 := obj3["data"].(map[string]interface{})
	assert.NoError(err)
	assert.NotNil(data3)
	assert.EqualValues(suite.eventTypeID, data3["eventTypeId"])
}
/*
//---------------------------------------------------------------------

func (suite *WorkflowTester) Test04PostTrigger() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	//suite.eventTypeId = "77bbe4c6-b1ac-4bbb-8e86-0f6e6a731c39"
	//suite.serviceId = "61985d9c-d4d0-45d9-a655-7dcf2dc08fad"

	client := suite.client

	trigger := &pzworkflow.Trigger{
		Name:        suite.triggerName,
		Enabled:     false,
		EventTypeID: suite.eventTypeID,
		Condition: map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"data.beta": 17,
				},
			},
		},
		Job: pzworkflow.JobRequest{
			CreatedBy: suite.apiKey,
			JobType: pzworkflow.JobType{
				Type: "execute-service",
				Data: map[string]interface{}{
					"dataInputs": map[string]interface{}{
						"": map[string]interface{}{
							"content":  `{"name":"$alpha", "count":$beta}`,
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
	str, _ := piazza.StructInterfaceToString(trigger)
	println(str)

	ack, err := client.PostTrigger(trigger)
	if err != nil {
		println(err.Error())
	}
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

	items, err := client.GetAllTriggers(100, 0)
	assert.NoError(err)
	assert.True(len(*items) > 1)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
	items, err = client.QueryTriggers(query)
	assert.NoError(err)
	assert.True(len(*items) > 1)

	item, err := client.GetTrigger(suite.triggerID)
	assert.NoError(err)
	assert.NotNil(item)
	assert.EqualValues(suite.triggerID, item.TriggerID)
}

func (suite *WorkflowTester) Test05PutTrigger() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	item, err := client.GetTrigger(suite.triggerID)
	assert.NoError(err)
	assert.NotNil(item)
	assert.EqualValues(suite.triggerID, item.TriggerID)

	triggerUpdate := pzworkflow.TriggerUpdate{
		Enabled: true,
	}

	err = client.PutTrigger(suite.triggerID, &triggerUpdate)
	assert.NoError(err)
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) Test06PostEvent() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	eventY := &pzworkflow.Event{
		EventTypeID: suite.eventTypeID,
		Data: map[string]interface{}{
			"beta":  goodBeta,
			"alpha": goodAlpha,
		},
	}

	eventN := &pzworkflow.Event{
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

	items, err := client.GetAllEvents(100, 0)
	assert.NoError(err)
	assert.True(len(*items) > 1)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
	items, err = client.QueryEvents(query)
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

	alert := &pzworkflow.Alert{
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

	items, err := client.GetAllAlerts(100, 0)
	assert.NoError(err)
	assert.True(len(*items) > 1)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
	items, err = client.QueryAlerts(query)
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
		"greeting": "Hello, %s!", 
		"countSquared": %d
	}`
	jsn = fmt.Sprintf(jsn, goodAlpha, goodBeta*goodBeta)
	assert.JSONEq(jsn, content)
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) Test12TestElasticsearch() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	version, err := client.TestElasticsearchGetVersion()
	assert.NoError(err)
	assert.EqualValues("2.2.0", *version)

	time.Sleep(3 * time.Second)

	var id piazza.Ident

	body := &pzworkflow.TestElasticsearchBody{Value: 17, ID: "19"}

	{
		retBody, err := client.TestElasticsearchPost(body)
		assert.NoError(err)
		assert.Equal(17, retBody.Value)
		assert.NotEmpty(retBody.ID)
		id = retBody.ID
	}
	time.Sleep(3 * time.Second)

	{
		retBody, err := client.TestElasticsearchGetOne(id)
		assert.NoError(err)
		assert.Equal(17, retBody.Value)
		assert.NotEmpty(retBody.ID)
	}
}

func (suite *WorkflowTester) Test13RepeatingEvent() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	allEvents, err := client.GetAllEventsByEventType(suite.eventTypeID)
	assert.NoError(err)
	numEventsBefore := len(*allEvents)

	repeatingEvent := &pzworkflow.Event{
		EventTypeID: suite.eventTypeID,
		Data: map[string]interface{}{
			"beta":  goodBeta,
			"alpha": goodAlpha,
		},
		CronSchedule: "*/ /*2 * * * * *",
	}

	ack, err := client.PostEvent(repeatingEvent)
	assert.NoError(err)
	assert.NotNil(ack)
	suite.repeatID = ack.EventID
	log.Printf("RepeatId: %s", suite.repeatID)

	time.Sleep(20 * time.Second)

	err = client.DeleteEvent(suite.repeatID)
	assert.NoError(err)

	allEvents, err = client.GetAllEventsByEventType(suite.eventTypeID)
	assert.NoError(err)
	numEventsAfter := len(*allEvents)

	numEventsCreated := numEventsAfter - numEventsBefore
	log.Printf("Number of repeating events created: %d", numEventsCreated)
	assert.InDelta(10, numEventsCreated, 4.0)
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
*/