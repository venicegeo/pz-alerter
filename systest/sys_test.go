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

package workflow_systest

import (
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

func sleep() {
	time.Sleep(1 * time.Second)
}

type WorkflowTester struct {
	suite.Suite
	client        *workflow.Client
	url           string
	apiKey        string
	uniq          string
	eventTypeId   piazza.Ident
	eventTypeName string
	triggerName   string
	serviceId     string
}

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

func (suite *WorkflowTester) xTest00Init() {
	//t := suite.T()
	//assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()
}

func (suite *WorkflowTester) Test01RegisterService() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	jsn :=
		`{
		"url" : "http://pzsvc-hello.int.geointservices.io",
		"contractUrl" : "http://pzsvc-hello.int.geointservices.io",
		"method" : "GET",
		"resourceMetadata" : {
			"name" : "HELLO World Test",
			"description" : "Hello world test",
			"classType" : "U"
		}
	}`
	_ = jsn
	jsn2 := map[string]interface{}{
		"serviceId":   "0bcc6896-642e-4a30-a01e-6bd0467b57ba",
		"url":         "http://pzsvc-hello.int.geointservices.io",
		"contractUrl": "http://pzsvc-hello.int.geointservices.io",
		"method":      "GET",
		"resourceMetadata": map[string]interface{}{
			"name":        "HELLO World Test",
			"description": "Hello world test",
			"classType":   "U",
		},
	}

	url := strings.Replace(suite.url, "workflow", "gateway", 1)
	h := piazza.Http{
		BaseUrl:    url,
		ApiKey:     suite.apiKey,
		Preflight:  piazza.SimplePreflight,
		Postflight: piazza.SimplePostflight,
	}
	//body := strings.NewReader(jsn)

	obj := map[string]interface{}{}
	resp, err := h.Post("/service", jsn2, &obj)
	assert.NoError(err)
	assert.NotNil(resp)

	//raw, err := ioutil.ReadAll(resp.Body)
	//assert.NoError(err)
	log.Printf("%#v", obj)
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) xTest02PostEventType() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	eventType := &workflow.EventType{
		Name: suite.eventTypeName,
		Mapping: map[string]elasticsearch.MappingElementTypeName{
			"alpha": elasticsearch.MappingElementTypeString,
			"beta":  elasticsearch.MappingElementTypeInteger,
		},
	}

	ack, err := client.PostEventType(eventType)
	assert.NoError(err)
	assert.NotNil(ack)

	suite.eventTypeId = ack.EventTypeId
}

func (suite *WorkflowTester) xTest03GetEventType() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	items, err := client.GetAllEventTypes()
	assert.NoError(err)
	assert.True(len(*items) > 1)

	item, err := client.GetEventType((*items)[0].EventTypeId)
	assert.NoError(err)
	assert.NotNil(item)
	assert.EqualValues((*items)[0].EventTypeId, item.EventTypeId)
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) xTest04PostTrigger() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	trigger := &workflow.Trigger{
		Title:   suite.triggerName,
		Enabled: true,

		Condition: workflow.Condition{
			EventTypeIds: []piazza.Ident{suite.eventTypeId},
			Query: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"num": 17,
					},
				},
			},
		},
		Job: workflow.Job{
			CreatedBy: "test",
			JobType: map[string]interface{}{
				"type": "execute-service",
				"data": map[string]interface{}{
					// "dataInputs": map[string]interface{},
					// "dataOutput": map[string]interface{},
					"serviceId": suite.serviceId,
				},
			},
		},
	}

	ack, err := client.PostTrigger(trigger)
	assert.NoError(err)
	assert.NotNil(ack)
}

func (suite *WorkflowTester) xTest04Gets() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	items, err := client.GetAllEventTypes()
	assert.NoError(err)
	assert.True(len(*items) > 1)

	item, err := client.GetEventType((*items)[0].EventTypeId)
	assert.NoError(err)
	assert.NotNil(item)
	assert.EqualValues((*items)[0].EventTypeId, item.EventTypeId)
}

//---------------------------------------------------------------------

func (suite *WorkflowTester) xTest99Admin() {
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
