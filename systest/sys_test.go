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
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	//pzsyslog "github.com/venicegeo/pz-gocommon/syslog"
	pzworkflow "github.com/venicegeo/pz-workflow/workflow"
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

func (suite *WorkflowTester) postToGateway(endpoint string, body interface{}, obj *map[string]interface{}) (int, error) {
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
func (suite *WorkflowTester) putToGateway(endpoint string, body interface{}, obj *map[string]interface{}) (int, error) {
	h := piazza.Http{
		BaseUrl: suite.url,
		ApiKey:  suite.apiKey,
		//Preflight:  piazza.SimplePreflight,
		//Postflight: piazza.SimplePostflight,
	}
	return h.Put(endpoint, body, obj)
}
func (suite *WorkflowTester) deleteFromGateway(endpoint string, obj *map[string]interface{}) (int, error) {
	h := piazza.Http{
		BaseUrl: suite.url,
		ApiKey:  suite.apiKey,
		//Preflight:  piazza.SimplePreflight,
		//Postflight: piazza.SimplePostflight,
	}
	return h.Delete(endpoint, obj)
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

	log.Println("URL", suite.url)
	obj := map[string]interface{}{}
	code, err := suite.postToGateway("/service", body, &obj)
	assert.NoError(err)
	assert.Equal(201, code)
	assert.NotNil(obj)

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

	eventType := &pzworkflow.EventType{
		Name: suite.eventTypeName,
		Mapping: map[string]interface{}{
			"alpha": elasticsearch.MappingElementTypeString,
			"beta":  elasticsearch.MappingElementTypeInteger,
		},
	}

	obj := map[string]interface{}{}
	code, err := suite.postToGateway("/eventType", eventType, &obj)
	assert.NoError(err)
	assert.Equal(201, code)
	assert.NotNil(obj)

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

//---------------------------------------------------------------------

func (suite *WorkflowTester) Test04PostTrigger() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	//suite.eventTypeId = "77bbe4c6-b1ac-4bbb-8e86-0f6e6a731c39"
	//suite.serviceId = "61985d9c-d4d0-45d9-a655-7dcf2dc08fad"

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

	obj := map[string]interface{}{}
	code, err := suite.postToGateway("/trigger", trigger, &obj)
	assert.NoError(err)
	assert.Equal(201, code)
	assert.NotNil(obj)

	data := obj["data"].(map[string]interface{})
	assert.IsType(stringType, data["triggerId"])
	triggerID := data["triggerId"].(string)
	assert.NotEmpty(triggerID)

	suite.triggerID = piazza.Ident(triggerID)
	log.Printf("TriggerId: %s", suite.triggerID)
}

func (suite *WorkflowTester) Test05GetTrigger() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	obj := map[string]interface{}{}
	code, err := suite.getFromGateway("/trigger?size=100", &obj)
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
	code, err = suite.postToGateway("/trigger/query", query, &obj2)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj2)

	assert.IsType(listType, obj2["data"])
	data2 := obj2["data"].([]interface{})
	assert.True(len(data2) > 1)

	obj3 := map[string]interface{}{}
	code, err = suite.getFromGateway("/trigger/"+string(suite.triggerID), &obj3)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj3)

	assert.IsType(mapType, obj3["data"])
	data3 := obj3["data"].(map[string]interface{})
	assert.NoError(err)
	assert.NotNil(data3)
	assert.EqualValues(suite.triggerID, data3["triggerId"])
}

func (suite *WorkflowTester) Test05PutTrigger() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	obj := map[string]interface{}{}
	code, err := suite.getFromGateway("/trigger/"+string(suite.triggerID), &obj)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj)

	assert.IsType(mapType, obj["data"])
	data3 := obj["data"].(map[string]interface{})
	assert.NoError(err)
	assert.NotNil(data3)
	assert.EqualValues(suite.triggerID, data3["triggerId"])

	triggerUpdate := pzworkflow.TriggerUpdate{
		Enabled: true,
	}
	obj2 := map[string]interface{}{}
	code, err = suite.putToGateway("/trigger/"+string(suite.triggerID), triggerUpdate, &obj2)
	if err != nil {
		println(err.Error())
	}
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj2)
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) Test06PostEvent() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

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

	obj := map[string]interface{}{}
	code, err := suite.postToGateway("/event", eventY, &obj)
	//fmt.Println(obj) //The test often fails here due to elastic gateway timeout..
	assert.NoError(err)
	assert.Equal(201, code)
	assert.NotNil(obj)

	data := obj["data"].(map[string]interface{})
	assert.IsType(stringType, data["eventId"])
	eventIDYes := data["eventId"].(string)
	assert.NotEmpty(eventIDYes)

	suite.eventIDYes = piazza.Ident(eventIDYes)
	log.Printf("EventIdY: %s", suite.eventIDYes)

	obj2 := map[string]interface{}{}
	code, err = suite.postToGateway("/event", eventN, &obj2)
	assert.NoError(err)
	assert.Equal(201, code)
	assert.NotNil(obj2)

	data = obj2["data"].(map[string]interface{})
	assert.IsType(stringType, data["eventId"])
	eventIDNo := data["eventId"].(string)
	assert.NotEmpty(eventIDNo)

	suite.eventIDNo = piazza.Ident(eventIDNo)
	log.Printf("EventIdN: %s", suite.eventIDNo)
}

func (suite *WorkflowTester) Test07GetEvent() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	obj := map[string]interface{}{}
	code, err := suite.getFromGateway("/event?size=100", &obj)
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
	code, err = suite.postToGateway("/event/query", query, &obj2)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj2)

	assert.IsType(listType, obj2["data"])
	data2 := obj2["data"].([]interface{})
	assert.True(len(data2) > 1)

	obj3 := map[string]interface{}{}
	code, err = suite.getFromGateway("/event/"+string(suite.eventIDYes), &obj3)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj3)

	assert.IsType(mapType, obj3["data"])
	data3 := obj3["data"].(map[string]interface{})
	assert.NoError(err)
	assert.NotNil(data3)
	assert.EqualValues(suite.eventIDYes, data3["eventId"])
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) Test09GetAlert() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	obj := map[string]interface{}{}
	code, err := suite.getFromGateway("/alert?size=100", &obj)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj)

	assert.IsType(listType, obj["data"])
	data := obj["data"].([]interface{})
	assert.True(len(data) > 1)

	obj2 := map[string]interface{}{}
	code, err = suite.getFromGateway("/alert?triggerId="+string(suite.triggerID), &obj2)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj2)

	assert.IsType(listType, obj["data"])
	data = obj["data"].([]interface{})
	assert.True(len(data) > 1)
	alert := data[0].(map[string]interface{})
	assert.EqualValues(suite.eventIDYes, alert["eventId"])
	suite.alertID = piazza.Ident(alert["alertId"].(string))
	log.Printf("AlertId: %s", suite.alertID)
	suite.jobID = piazza.Ident(alert["jobId"].(string))
	log.Printf("JobId: %s", suite.jobID)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	obj3 := map[string]interface{}{}
	code, err = suite.postToGateway("/alert/query", query, &obj3)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj3)

	assert.IsType(listType, obj3["data"])
	data2 := obj3["data"].([]interface{})
	assert.True(len(data2) > 1)

	obj4 := map[string]interface{}{}
	code, err = suite.getFromGateway("/alert/"+string(suite.alertID), &obj4)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj4)

	assert.IsType(mapType, obj4["data"])
	data3 := obj4["data"].(map[string]interface{})
	assert.NoError(err)
	assert.NotNil(data3)
	assert.EqualValues(suite.alertID, data3["alertId"])
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) Test10GetJob() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	h := piazza.Http{
		BaseUrl: suite.url,
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

	h := piazza.Http{
		BaseUrl: suite.url,
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
/*
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
*/
func (suite *WorkflowTester) Test13RepeatingEvent() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	obj := map[string]interface{}{}
	code, err := suite.getFromGateway("/event?eventTypeId="+string(suite.eventTypeID), &obj)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj)

	assert.IsType(listType, obj["data"])
	data := obj["data"].([]interface{})
	assert.True(len(data) > 1)
	numEventsBefore := len(data)

	repeatingEvent := &pzworkflow.Event{
		EventTypeID: suite.eventTypeID,
		Data: map[string]interface{}{
			"beta":  goodBeta,
			"alpha": goodAlpha,
		},
		CronSchedule: "*/2 * * * * *",
	}

	obj2 := map[string]interface{}{}
	code, err = suite.postToGateway("/event", repeatingEvent, &obj2)
	assert.NoError(err)
	assert.Equal(201, code)
	assert.NotNil(obj2)

	data2 := obj2["data"].(map[string]interface{})
	assert.IsType(stringType, data2["eventId"])
	suite.repeatID = piazza.Ident(data2["eventId"].(string))
	log.Printf("RepeatId: %s", suite.repeatID)

	time.Sleep(20 * time.Second)

	obj3 := map[string]interface{}{}
	code, err = suite.deleteFromGateway("/event/"+string(suite.repeatID), &obj3)
	assert.Equal(200, code)
	assert.NoError(err)

	obj4 := map[string]interface{}{}
	code, err = suite.getFromGateway("/event?eventTypeId="+string(suite.eventTypeID), &obj4)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(obj4)

	assert.IsType(listType, obj4["data"])
	data = obj4["data"].([]interface{})
	assert.True(len(data) > 1)
	numEventsAfter := len(data)

	numEventsCreated := numEventsAfter - numEventsBefore
	log.Printf("Number of repeating events created: %d", numEventsCreated)
	assert.InDelta(10, numEventsCreated, 4.0)
}

type multiError struct {
	errors []error
}

func (m *multiError) add(err error) {
	if err != nil {
		m.errors = append(m.errors, err)
	}
}
func (m *multiError) error() error {
	var str string
	for _, err := range m.errors {
		str += err.Error() + "\n"
	}
	if str == "" {
		return nil
	} else {
		return errors.New(str)
	}
}

func (suite *WorkflowTester) TestRemoveTrace() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	merr := multiError{}
	test := func(debug string, code int, err error) {
		merr.add(err)
		if code != 200 {
			merr.add(errors.New(fmt.Sprintf(debug+": Code %d is not 200", code)))
		}
	}

	code, err := suite.deleteFromGateway("/service/"+suite.serviceID.String()+"?softDelete=false", &map[string]interface{}{})
	test("deleting service", code, err)
	code, err = suite.deleteFromGateway("/data/"+suite.dataID.String(), &map[string]interface{}{})
	test("deleting data", code, err)
	//code, err = suite.deleteFromGateway("/trigger/"+suite.triggerID.String(), &map[string]interface{}{})
	//test("deleting trigger", code, err)
	code, err = suite.deleteFromGateway("/event/"+suite.eventIDNo.String(), &map[string]interface{}{})
	test("deleting event n", code, err)
	code, err = suite.deleteFromGateway("/event/"+suite.eventIDYes.String(), &map[string]interface{}{})
	test("deleting event y", code, err)
	{
		events := map[string]interface{}{}
		code, err = suite.getFromGateway("/event?perPage=100&eventTypeId="+string(suite.eventTypeID), &events)
		test("getting events from eventtype "+suite.eventTypeID.String(), code, err)
		assert.NotNil(events)
		assert.IsType(listType, events["data"])
		data := events["data"].([]interface{})
		e := map[string]interface{}{}
		for _, ev := range data {
			e = ev.(map[string]interface{})
			fmt.Println("Deleting event:", e["eventId"].(string))
			code, err = suite.deleteFromGateway("/event/"+e["eventId"].(string), &map[string]interface{}{})
			test("deleting event "+e["eventId"].(string), code, err)
		}
	}
	{
		triggers := map[string]interface{}{}
		//code, err := suite.getFromGateway("/trigger?eventTypeId="+string(suite.eventTypeID), &triggers)
		query, err := piazza.StructStringToInterface(fmt.Sprintf(`{"query": {"bool": {"must": [{"term":{"eventTypeId":"%s"}}]}}}`, suite.eventTypeID.String()))
		merr.add(err)
		code, err = suite.postToGateway("/trigger/query?perPage=100", query, &triggers)
		test("posting query", code, err)
		assert.IsType(listType, triggers["data"])
		data := triggers["data"].([]interface{})
		fmt.Println("Found", len(data), "triggers")
		t := map[string]interface{}{}
		for _, tr := range data {
			t = tr.(map[string]interface{})
			fmt.Println("Deleting trigger:", t["triggerId"].(string))
			code, err = suite.deleteFromGateway("/trigger/"+t["triggerId"].(string), &map[string]interface{}{})
			test("deleting trigger "+t["triggerId"].(string), code, err)
		}
	}
	code, err = suite.deleteFromGateway("/eventType/"+suite.eventTypeID.String(), &map[string]interface{}{})
	test("deleting eventtype", code, err)
	assert.NoError(merr.error())
}

//---------------------------------------------------------------------
/*
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
