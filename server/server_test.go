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
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
	"github.com/venicegeo/pz-workflow/common"
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
	var tmp loggerPkg.ILoggerService = theLogger
	clogger := loggerPkg.NewCustomLogger(&tmp, piazza.PzWorkflow, config.GetAddress())

	theUuidgen, err := uuidgenPkg.NewMockUuidGenService(sys)
	if err != nil {
		log.Fatal(err)
	}

	es, err := elasticsearch.NewElasticsearchClient(sys, true)
	if err != nil {
		log.Fatal(err)
	}
	sys.Services[piazza.PzElasticSearch] = es

	routes, err := CreateHandlers(sys, clogger, theUuidgen)
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
	if err != nil {
		log.Printf("--- %s: %s", path, string(data))
	}
	assert.NoError(err)

	return result
}

func (suite *ServerTester) Delete(path string) {
	t := suite.T()
	assert := assert.New(t)

	resp, err := piazza.HTTPDelete(suite.url + path)
	assert.NoError(err)
	assert.NotNil(resp)
}

//---------------------------------------------------------------------------

var eventTypeName = "MYTYPE"

func (suite *ServerTester) getAllTypes() []common.EventType {
	t := suite.T()
	assert := assert.New(t)
	var err error

	resp := suite.Get("/eventtypes")
	var typs []common.EventType
	err = common.SuperConvert(resp, &typs)
	assert.NoError(err)

	return typs
}

func (suite *ServerTester) postOneType() common.Ident {
	t := suite.T()
	assert := assert.New(t)
	var err error

	mapping := map[string]elasticsearch.MappingElementTypeName{
		"num": elasticsearch.MappingElementTypeInteger,
	}
	eventType := &common.EventType{Name: eventTypeName, Mapping: mapping}

	resp := suite.Post("/eventtypes", eventType)
	resp2 := &common.WorkflowIdResponse{}
	err = common.SuperConvert(resp, resp2)
	assert.NoError(err)

	return resp2.ID
}

func (suite *ServerTester) getOneType(id common.Ident) common.EventType {
	t := suite.T()
	assert := assert.New(t)
	var err error

	resp := suite.Get("/eventtypes/" + string(id))
	resp2 := common.EventType{}
	err = common.SuperConvert(resp, &resp2)
	assert.NoError(err)
	return resp2
}

func (suite *ServerTester) deleteOneType(id common.Ident) {
	suite.Delete("/eventtypes/" + string(id))
}

//---------------------------------------------------------------------------

func (suite *ServerTester) getAllEvents() []common.Event {
	t := suite.T()
	assert := assert.New(t)
	var err error

	resp := suite.Get("/events")
	var events []common.Event
	err = common.SuperConvert(resp, &events)
	assert.NoError(err)

	return events
}

func (suite *ServerTester) postOneEvent() common.Ident {
	t := suite.T()
	assert := assert.New(t)
	var err error

	tid := suite.postOneType()

	event := &common.Event{
		EventTypeId: tid,
		Date:        time.Now(),
		Data: map[string]interface{}{
			"num": 17,
		},
	}

	resp := suite.Post("/events/"+eventTypeName, event)

	resp2 := &common.WorkflowIdResponse{}
	err = common.SuperConvert(resp, resp2)
	assert.NoError(err)

	return resp2.ID
}

func (suite *ServerTester) getOneEvent(id common.Ident) common.Event {
	t := suite.T()
	assert := assert.New(t)
	var err error

	resp := suite.Get("/events/" + eventTypeName + "/" + string(id))
	resp2 := common.Event{}
	err = common.SuperConvert(resp, &resp2)
	assert.NoError(err)
	return resp2
}

func (suite *ServerTester) deleteOneEvent(id common.Ident) {
	suite.Delete("/events/" + eventTypeName + "/" + string(id))
}

//---------------------------------------------------------------------------

func (suite *ServerTester) getAllTriggers() []common.Trigger {
	t := suite.T()
	assert := assert.New(t)
	var err error

	resp := suite.Get("/triggers")
	var triggers []common.Trigger
	err = common.SuperConvert(resp, &triggers)
	assert.NoError(err)

	return triggers
}

func (suite *ServerTester) postOneTrigger() common.Ident {
	t := suite.T()
	assert := assert.New(t)
	var err error

	tid := suite.postOneType()

	trigger := &common.Trigger{
		Title: "MY TRIGGER TITLEr",
		Condition: common.Condition{
			EventId: tid,
			Query: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"num": 17,
					},
				},
			},
		},
		Job: common.Job{
			Task: "MY TASK",
		},
	}

	resp := suite.Post("/triggers", trigger)
	resp2 := &common.WorkflowIdResponse{}
	err = common.SuperConvert(resp, resp2)
	assert.NoError(err)

	return resp2.ID
}

func (suite *ServerTester) getOneTrigger(id common.Ident) common.Trigger {
	t := suite.T()
	assert := assert.New(t)
	var err error

	resp := suite.Get("/triggers/" + string(id))
	resp2 := common.Trigger{}
	err = common.SuperConvert(resp, &resp2)
	assert.NoError(err)
	return resp2
}

func (suite *ServerTester) deleteOneTrigger(id common.Ident) {
	suite.Delete("/triggers/" + string(id))
}

//---------------------------------------------------------------------------

func (suite *ServerTester) getAllAlerts() []common.Alert {
	t := suite.T()
	assert := assert.New(t)
	var err error

	resp := suite.Get("/alerts")
	var alerts []common.Alert
	err = common.SuperConvert(resp, &alerts)
	assert.NoError(err)

	return alerts
}

func (suite *ServerTester) postOneAlert() common.Ident {
	t := suite.T()
	assert := assert.New(t)
	var err error

	eid := suite.postOneEvent()
	tid := suite.postOneTrigger()

	alert := &common.Alert{
		TriggerId: tid,
		EventId:   eid,
	}

	resp := suite.Post("/alerts", alert)
	resp2 := &common.WorkflowIdResponse{}
	err = common.SuperConvert(resp, resp2)
	assert.NoError(err)

	return resp2.ID
}

func (suite *ServerTester) getOneAlert(id common.Ident) common.Alert {
	t := suite.T()
	assert := assert.New(t)
	var err error

	resp := suite.Get("/alerts/" + string(id))
	resp2 := common.Alert{}
	err = common.SuperConvert(resp, &resp2)
	assert.NoError(err)
	return resp2
}

func (suite *ServerTester) deleteOneAlert(id common.Ident) {
	suite.Delete("/alerts/" + string(id))
}

//---------------------------------------------------------------------------
//---------------------------------------------------------------------------

func (suite *ServerTester) Test_1_EventType() {
	t := suite.T()
	assert := assert.New(t)

	typs := suite.getAllTypes()
	assert.Len(typs, 0)

	id := suite.postOneType()

	typs = suite.getAllTypes()
	assert.Len(typs, 1)

	typ := suite.getOneType(id)
	assert.EqualValues(string(id), string(typ.ID))

	suite.deleteOneType(id)

	typs = suite.getAllTypes()
	assert.Len(typs, 0)
}

func (suite *ServerTester) Test_2_Event() {
	t := suite.T()
	assert := assert.New(t)

	events := suite.getAllEvents()
	assert.Len(events, 0)

	id := suite.postOneEvent()

	events = suite.getAllEvents()
	assert.Len(events, 1)

	event := suite.getOneEvent(id)
	assert.EqualValues(string(id), string(event.ID))

	suite.deleteOneEvent(id)

	events = suite.getAllEvents()
	assert.Len(events, 0)
}

func (suite *ServerTester) Test_3_Trigger() {
	t := suite.T()
	assert := assert.New(t)

	triggers := suite.getAllTriggers()
	assert.Len(triggers, 0)

	id := suite.postOneTrigger()

	triggers = suite.getAllTriggers()
	assert.Len(triggers, 1)

	trigger := suite.getOneTrigger(id)
	assert.EqualValues(string(id), string(trigger.ID))

	suite.deleteOneTrigger(id)

	triggers = suite.getAllTriggers()
	assert.Len(triggers, 0)
}

func (suite *ServerTester) Test_4_Alert() {
	t := suite.T()
	assert := assert.New(t)

	alerts := suite.getAllAlerts()
	assert.Len(alerts, 0)

	id := suite.postOneAlert()

	alerts = suite.getAllAlerts()
	assert.Len(alerts, 1)

	alert := suite.getOneAlert(id)
	assert.EqualValues(string(id), string(alert.ID))

	suite.deleteOneAlert(id)

	alerts = suite.getAllAlerts()
	assert.Len(alerts, 0)
}

//---------------------------------------------------------------------------

func (suite *ServerTester) xTestAAAEmptyEventType() {

	t := suite.T()
	assert := assert.New(t)

	resp := suite.Get("/eventtypes")
	log.Printf("++ %#v ++", resp)
	resp2 := resp.([]interface{})
	assert.Len(resp2, 0)
}

func (suite *ServerTester) xTestOne() {

	t := suite.T()
	assert := assert.New(t)

	var err error
	//var idResponse *common.WorkflowIdResponse

	var eventTypeName = "EventTypeA"

	var et1Id common.Ident
	{
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num":    elasticsearch.MappingElementTypeInteger,
			"str":    elasticsearch.MappingElementTypeString,
			"apiKey": elasticsearch.MappingElementTypeString,
			"jobId":  elasticsearch.MappingElementTypeString,
		}

		eventType := &common.EventType{Name: eventTypeName, Mapping: mapping}

		resp := suite.Post("/eventtypes", eventType)
		defer piazza.HTTPDelete("/eventtypes/" + string(eventType.ID))

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
				EventId: et1Id,
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"num": 17,
						},
					},
				},
			},
			Job: common.Job{
				//Task: "the x1 task",
				// Using a GetJob call as it is as close to a 'noop' as I could find.
				Task: `{"apiKey": "$apiKey", "jobType": {"type": "get", "jobId": "$jobId"}}`,
			},
		}

		resp := suite.Post("/triggers", x1)
		defer piazza.HTTPDelete("/triggers/" + string(x1.ID))
		resp2 := &common.WorkflowIdResponse{}
		err = common.SuperConvert(resp, resp2)
		assert.NoError(err)

		t1Id = resp2.ID
	}

	var e1Id common.Ident
	{
		// will cause trigger TRG1
		e1 := &common.Event{
			EventTypeId: et1Id,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num":    17,
				"str":    "quick",
				"apiKey": "my-api-key-38n987",
				"jobId":  "789a6531-85a9-4098-aa3c-e90d07d9b8a3",
			},
		}

		resp := suite.Post("/events/"+eventTypeName, e1)
		defer piazza.HTTPDelete("/events/" + eventTypeName + "/" + string(e1.ID))
		resp2 := &common.WorkflowIdResponse{}
		err = common.SuperConvert(resp, resp2)
		assert.NoError(err)

		e1Id = resp2.ID
	}

	{
		// will cause no triggers
		e1 := &common.Event{
			EventTypeId: et1Id,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": 18,
				"str": "brown",
				// Probably don't need the following as job shouldn't be executed.
				"apiKey": "my-api-key-38n987",
				"jobId":  "789a6531-85a9-4098-aa3c-e90d07d9b8a3",
			},
		}

		resp := suite.Post("/events/"+eventTypeName, e1)
		defer piazza.HTTPDelete("/events/" + eventTypeName + "/" + string(e1.ID))
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

func (suite *ServerTester) xTestEventMapping() {

	t := suite.T()
	assert := assert.New(t)

	var err error

	var eventTypeName1 = "Type1"
	var eventTypeName2 = "Type2"

	eventtypeF := func(typ string) common.Ident {
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num": elasticsearch.MappingElementTypeInteger,
		}

		eventType := &common.EventType{Name: typ, Mapping: mapping}

		resp1 := suite.Post("/eventtypes", eventType)
		defer piazza.HTTPDelete("/eventtypes/" + string(eventType.ID))

		resp2 := &common.WorkflowIdResponse{}
		err = common.SuperConvert(resp1, resp2)
		assert.NoError(err)

		resp3 := suite.Get("/eventtypes/" + string(resp2.ID))
		resp4 := &common.WorkflowIdResponse{}
		err = common.SuperConvert(resp3, resp4)
		assert.NoError(err)

		assert.EqualValues(resp4.ID, resp2.ID)

		return resp2.ID
	}

	eventF := func(typeId common.Ident, typ string, value int) common.Ident {
		e1 := &common.Event{
			EventTypeId: typeId,
			Date:        time.Now(),
			Data: map[string]interface{}{
				"num": value,
			},
		}

		resp1 := suite.Post("/events/"+typ, e1)
		defer piazza.HTTPDelete("/events/" + typ + "/" + string(e1.ID))
		resp2 := &common.WorkflowIdResponse{}
		err = common.SuperConvert(resp1, resp2)
		assert.NoError(err)

		resp3 := suite.Get("/events/" + typ + "/" + string(resp2.ID))
		resp4 := &common.WorkflowIdResponse{}
		err = common.SuperConvert(resp3, resp4)
		assert.NoError(err)

		assert.EqualValues(resp4.ID, resp2.ID)

		return resp2.ID
	}

	dump := func(typ string, expected int) {
		if typ != "" {
			typ = "/" + typ
		}
		x := suite.Get("/events" + typ)
		y := x.([]interface{})
		assert.Len(y, expected)
	}

	et1Id := eventtypeF(eventTypeName1)
	et2Id := eventtypeF(eventTypeName2)

	{
		resp := suite.Get("/eventtypes")
		log.Printf("++ %#v ++", resp)
		resp2 := resp.([]interface{})
		assert.Len(resp2, 2)
	}

	dump(eventTypeName1, 0)
	dump("", 0)

	{
		x := suite.Get("/eventtypes")
		y := x.([]interface{})
		assert.Len(y, 2)
	}

	{
		x := suite.Get("/eventtypes/" + string(et1Id))
		y := x.(map[string]interface{})
		assert.EqualValues(string(et1Id), string(y["id"].(string)))
	}

	e1Id := eventF(et1Id, eventTypeName1, 17)
	dump(eventTypeName1, 1)

	e2Id := eventF(et1Id, eventTypeName1, 18)
	dump(eventTypeName1, 2)

	e3Id := eventF(et2Id, eventTypeName2, 19)
	dump(eventTypeName2, 1)

	dump("", 3)

	suite.Delete("/events/" + eventTypeName1 + "/" + string(e1Id))
	suite.Delete("/events/" + eventTypeName1 + "/" + string(e2Id))
	suite.Delete("/events/" + eventTypeName2 + "/" + string(e3Id))

	suite.Delete("/eventtypes/" + string(et1Id))
	suite.Delete("/eventtypes/" + string(et2Id))
}

func (suite *ServerTester) xestTwo() {
	t := suite.T()
	assert := assert.New(t)
	assert.Equal(17, 10+7)
}
