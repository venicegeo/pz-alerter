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

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
	"github.com/venicegeo/pz-workflow/common"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

type ServerTester struct {
	suite.Suite
	sys *piazza.System
	url string
}

func (suite *ServerTester) SetupSuite() {
	t := suite.T()
	assert := assert.New(t)

	config, err := piazza.NewConfig(piazza.PzWorkflow, piazza.ConfigModeTest)
	if err != nil {
		log.Fatal(err)
	}

	sys, err := piazza.NewSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	theLogger, err := loggerPkg.NewMockLoggerService(sys)
	if err != nil {
		log.Fatal(err)
	}

	theUuidgen, err := uuidgenPkg.NewMockUuidGenService(sys)
	if err != nil {
		log.Fatal(err)
	}

	routes, err := CreateHandlers(sys, theLogger, theUuidgen)
	if err != nil {
		log.Fatal(err)
	}

	_ = sys.StartServer(routes)

	suite.sys = sys

	suite.url = fmt.Sprintf("http://%s/v1", sys.Config.GetBindToAddress())

	assert.Len(sys.Services, 4)
}

func (suite *ServerTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func TestRunSuite(t *testing.T) {
	s := new(ServerTester)
	suite.Run(t, s)
}

func (suite *ServerTester) Post(path string, body interface{}) interface{} {
	t := suite.T()
	assert := assert.New(t)

	bodyBytes, err := json.Marshal(body)
	assert.NoError(err)

	resp, err := http.Post(suite.url+path, piazza.ContentTypeJSON, bytes.NewBuffer(bodyBytes))
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusCreated, resp.StatusCode)

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	assert.NoError(err)

	var result interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(err)

	return result
}

func (suite *ServerTester) Get(path string) interface{} {
	t := suite.T()
	assert := assert.New(t)

	resp, err := http.Get(suite.url + path)
	assert.NoError(err)
	assert.NotNil(resp)

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	assert.NoError(err)

	var result interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(err)

	return result
}

//---------------------------------------------------------------------------

func (suite *ServerTester) TestOne() {

	t := suite.T()
	assert := assert.New(t)

	var err error
	//var idResponse *common.WorkflowIdResponse

	var eventTypeName = "EventTypeA"

	var et1Id common.Ident
	{
		mapping := map[string]piazza.MappingElementTypeName{
			"num": piazza.MappingElementTypeInteger,
			"str": piazza.MappingElementTypeString,
		}

		eventType := &common.EventType{Name: eventTypeName, Mapping: mapping}

		resp := suite.Post("/eventtypes", eventType)

		resp2 := &common.WorkflowIdResponse{}
		err = common.SuperConvert(resp, resp2)
		assert.NoError(err)

		et1Id = resp2.ID
	}

	var t1Id common.Ident
	{
		x1 := &common.Trigger{
			Title: "the x1 trigger",
			Condition: common.Condition{
				EventType: et1Id,
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"num": 17,
						},
					},
				},
			},
			Job: common.Job{
				Task: "the x1 task",
			},
		}

		resp := suite.Post("/triggers", x1)
		resp2 := &common.WorkflowIdResponse{}
		err = common.SuperConvert(resp, resp2)
		assert.NoError(err)

		t1Id = resp2.ID
	}

	var e1Id common.Ident
	{
		// will cause trigger TRG1
		e1 := &common.Event{
			EventType: et1Id,
			Date:      time.Now(),
			Data: map[string]interface{}{
				"num": 17,
				"str": "quick",
			},
		}

		resp := suite.Post("/events/"+eventTypeName, e1)
		resp2 := &common.WorkflowIdResponse{}
		err = common.SuperConvert(resp, resp2)
		assert.NoError(err)

		e1Id = resp2.ID
	}

	{
		// will cause no triggers
		e1 := &common.Event{
			EventType: et1Id,
			Date:      time.Now(),
			Data: map[string]interface{}{
				"num": 18,
				"str": "brown",
			},
		}

		resp := suite.Post("/events/"+eventTypeName, e1)
		resp2 := &common.WorkflowIdResponse{}
		err = common.SuperConvert(resp, resp2)
		assert.NoError(err)
	}

	{
		resp := suite.Get("/alerts")

		var alerts []common.Alert
		common.SuperConvert(resp, &alerts)
		assert.Len(alerts, 1)

		alert0 := alerts[0]
		assert.EqualValues(e1Id, alert0.EventId)
		assert.EqualValues(t1Id, alert0.TriggerId)
	}
}

func (suite *ServerTester) TestTwo() {
	t := suite.T()
	assert := assert.New(t)
	assert.Equal(17, 10 + 7)	
}

