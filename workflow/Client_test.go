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

package workflow

import (
	"log"
	"time"

	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/logger"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/uuidgen"
)

type ClientTester struct {
	suite.Suite
	logger     loggerPkg.IClient
	uuidgenner uuidgenPkg.IUuidGenService
	client     *Client
	sys        *piazza.SystemConfig
}

func (suite *ClientTester) SetupSuite() {
	assertNoData(suite.T(), suite.client)
}

func (suite *ClientTester) TearDownSuite() {
	assertNoData(suite.T(), suite.client)
}

//---------------------------------------------------------------------------

func (suite *ClientTester) Test11Admin() {
	t := suite.T()
	assert := assert.New(t)

	client := suite.client

	log.Printf("AdminStats:")
	_, err := client.GetFromAdminStats()
	assert.NoError(err)
}

func (suite *ClientTester) Test12AlertResource() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	var err error

	a1 := Alert{TriggerId: "dummyT1", EventId: "dummyE1"}
	id, err := client.PostOneAlert(&a1)
	assert.NoError(err)

	sleep()

	alerts, err := client.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 1)
	assert.EqualValues(id, (*alerts)[0].AlertId)
	assert.EqualValues("dummyT1", (*alerts)[0].TriggerId)
	assert.EqualValues("dummyE1", (*alerts)[0].EventId)

	alert, err := client.GetOneAlert(id)
	assert.NoError(err)
	assert.EqualValues(id, alert.AlertId)

	alert, err = client.GetOneAlert("nosuchalert1")
	assert.Error(err)

	err = client.DeleteOneAlert("nosuchalert2")
	assert.Error(err)

	err = client.DeleteOneAlert(id)
	assert.NoError(err)

	sleep()

	alert, err = client.GetOneAlert(id)
	assert.Error(err)
	assert.Nil(alert)

	alerts, err = client.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 0)
}

func (suite *ClientTester) Test13EventResource() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	var err error

	mapping := map[string]elasticsearch.MappingElementTypeName{
		"myint": elasticsearch.MappingElementTypeString,
		"mystr": elasticsearch.MappingElementTypeString,
	}
	eventTypeName := "mytype"
	eventType := &EventType{Name: eventTypeName, Mapping: mapping}
	etID, err := client.PostOneEventType(eventType)
	assert.NoError(err)
	defer func() {
		err = client.DeleteOneEventType(etID)
		assert.NoError(err)
	}()

	event := &Event{
		EventTypeId: etID,
		CreatedOn:   time.Now(),
		Data: map[string]interface{}{
			"myint": 17,
			"mystr": "quick",
		},
	}
	eID, err := client.PostOneEvent(eventTypeName, event)
	assert.NoError(err)

	defer func() {
		err = client.DeleteOneEvent(eventTypeName, eID)
		assert.NoError(err)
	}()

	//events, err := workflow.GetAllEvents("")
	//assert.NoError(err)
	//assert.Len(*events, 1)
	//assert.EqualValues(eID, (*events)[0].ID)

	sleep()

	tmp, err := client.GetOneEvent(eventTypeName, eID)
	assert.NoError(err)
	assert.EqualValues(eID, tmp.EventId)
}

func (suite *ClientTester) Test14EventTypeResource() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	var err error

	mapping := map[string]elasticsearch.MappingElementTypeName{
		"myint": elasticsearch.MappingElementTypeString,
		"mystr": elasticsearch.MappingElementTypeString,
	}
	eventType := &EventType{Name: "typnam", Mapping: mapping}
	id, err := client.PostOneEventType(eventType)
	defer func() {
		err = client.DeleteOneEventType(id)
		assert.NoError(err)
	}()

	sleep()

	eventTypes, err := client.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*eventTypes, 1)
	assert.EqualValues(id, (*eventTypes)[0].EventTypeId)

	tmp, err := client.GetOneEventType(id)
	assert.NoError(err)
	assert.EqualValues(id, tmp.EventTypeId)
}

func (suite *ClientTester) Test15One() {

	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	var err error
	var eventTypeName = "EventTypeA"

	var etID Ident
	{
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num": elasticsearch.MappingElementTypeInteger,
			"str": elasticsearch.MappingElementTypeString,
		}

		eventType := &EventType{Name: eventTypeName, Mapping: mapping}

		etID, err = client.PostOneEventType(eventType)
		assert.NoError(err)

		defer func() {
			err := client.DeleteOneEventType(etID)
			assert.NoError(err)
		}()
	}

	sleep()

	var tID Ident
	{
		x1 := &Trigger{
			Title: "the x1 trigger",
			Condition: Condition{
				EventTypeIds: []Ident{etID},
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"num": 17,
						},
					},
				},
			},
			Job: Job{
				Username: "test",
				JobType: map[string]interface{}{
					"type": "execute-service",
					"data": map[string]interface{}{
						// "dataInputs": map[string]interface{},
						// "dataOutput": map[string]interface{},
						"serviceId": "ddd5134",
					},
				},
			},
		}

		tID, err = client.PostOneTrigger(x1)
		assert.NoError(err)

		defer func() {
			err := client.DeleteOneTrigger(tID)
			assert.NoError(err)
		}()
	}

	var e1ID Ident
	{
		// will cause trigger t1ID
		e1 := &Event{
			EventTypeId: etID,
			CreatedOn:   time.Now(),
			Data: map[string]interface{}{
				"num":      17,
				"str":      "quick",
				"userName": "my-api-key-38n987",
			},
		}

		e1ID, err = client.PostOneEvent(eventTypeName, e1)
		assert.NoError(err)

		defer func() {
			err := client.DeleteOneEvent(eventTypeName, e1ID)
			assert.NoError(err)
		}()
	}

	sleep()

	var e2ID Ident
	{
		// will cause no triggers
		e2 := &Event{
			EventTypeId: etID,
			CreatedOn:   time.Now(),
			Data: map[string]interface{}{
				"num": 18,
				"str": "brown",
			},
		}

		e2ID, err = client.PostOneEvent(eventTypeName, e2)
		assert.NoError(err)

		defer func() {
			err := client.DeleteOneEvent(eventTypeName, e2ID)
			assert.NoError(err)
		}()
	}
	sleep()

	//{
	//	ary, err := client.GetAllEvents("")
	//	assert.NoError(err)
	//	assert.Len(*ary, 2)
	//}

	var aID Ident
	{
		if MOCKING {
			t.Skip("Skipping test, because mocking")
		}
		alerts, err := client.GetAllAlerts()
		assert.NoError(err)
		assert.Len(*alerts, 1)
		alert0 := (*alerts)[0]
		assert.EqualValues(e1ID, alert0.EventId)
		assert.EqualValues(tID, alert0.TriggerId)

		aID = alert0.AlertId

		defer func() {
			err := client.DeleteOneAlert(aID)
			assert.NoError(err)
		}()
	}
}

func (suite *ClientTester) Test16TriggerResource() {
	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	var err error

	mapping := map[string]elasticsearch.MappingElementTypeName{
		"myint": elasticsearch.MappingElementTypeString,
		"mystr": elasticsearch.MappingElementTypeString,
	}
	eventType := &EventType{Name: "typnam", Mapping: mapping}
	etID, err := client.PostOneEventType(eventType)

	defer func() {
		err = client.DeleteOneEventType(etID)
		assert.NoError(err)
	}()

	t1 := Trigger{
		Title: "the x1 trigger",
		Condition: Condition{
			EventTypeIds: []Ident{etID},
			Query: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"myint": 17,
					},
				},
			},
		},
		Job: Job{
			Username: "test",
			JobType: map[string]interface{}{
				"type": "execute-service",
				"data": map[string]interface{}{
					// "dataInputs": map[string]interface{},
					// "dataOutput": map[string]interface{},
					"serviceId": "ddd5134",
				},
			},
		},
	}
	t1ID, err := client.PostOneTrigger(&t1)
	assert.NoError(err)

	defer func() {
		err = client.DeleteOneTrigger(t1ID)
		assert.NoError(err)
	}()

	sleep()

	tmp, err := client.GetOneTrigger(t1ID)
	assert.NoError(err)
	assert.EqualValues(t1ID, tmp.TriggerId)

	triggers, err := client.GetAllTriggers()
	assert.NoError(err)
	assert.Len(*triggers, 1)
	assert.EqualValues(t1ID, (*triggers)[0].TriggerId)
}

func (suite *ClientTester) Test17Triggering() {

	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	var err error

	//-----------------------------------------------------

	var etC, etD, etE Ident
	{
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num": elasticsearch.MappingElementTypeInteger,
			"str": elasticsearch.MappingElementTypeString,
		}
		eventTypeC := &EventType{Name: "EventType C", Mapping: mapping}
		eventTypeD := &EventType{Name: "EventType D", Mapping: mapping}
		eventTypeE := &EventType{Name: "EventType E", Mapping: mapping}
		etC, err = client.PostOneEventType(eventTypeC)
		assert.NoError(err)
		etD, err = client.PostOneEventType(eventTypeD)
		assert.NoError(err)
		etE, err = client.PostOneEventType(eventTypeE)
		assert.NoError(err)

		sleep()

		eventTypes, err := client.GetAllEventTypes()
		assert.NoError(err)
		assert.Len(*eventTypes, 3)
	}
	sleep()

	defer func() {
		client.DeleteOneEventType(etC)
		assert.NoError(err)
		client.DeleteOneEventType(etD)
		assert.NoError(err)
		client.DeleteOneEventType(etE)
		assert.NoError(err)
	}()

	////////////////

	var tA, tB Ident
	{
		t1 := &Trigger{
			Title: "Trigger A",
			Condition: Condition{
				EventTypeIds: []Ident{etC},
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"str": "quick",
						},
					},
				},
			},
			Job: Job{
				Username: "test",
				JobType: map[string]interface{}{
					"type": "execute-service",
					"data": map[string]interface{}{
						// "dataInputs": map[string]interface{},
						// "dataOutput": map[string]interface{},
						"serviceId": "ddd5134",
					},
				},
			},
		}
		tA, err = client.PostOneTrigger(t1)
		assert.NoError(err)
		defer func() {
			client.DeleteOneTrigger(tA)
			assert.NoError(err)
		}()

		t2 := &Trigger{
			Title: "Trigger B",
			Condition: Condition{
				EventTypeIds: []Ident{etD},
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"num": 18,
						},
					},
				},
			},
			Job: Job{
				Username: "test",
				JobType: map[string]interface{}{
					"type": "execute-service",
					"data": map[string]interface{}{
						// "dataInputs": map[string]interface{},
						// "dataOutput": map[string]interface{},
						"serviceId": "ddd5134",
					},
				},
			},
		}
		tB, err = client.PostOneTrigger(t2)
		assert.NoError(err)
		defer func() {
			client.DeleteOneTrigger(tB)
			assert.NoError(err)
		}()

		sleep()

		triggers, err := client.GetAllTriggers()
		assert.NoError(err)
		assert.Len(*triggers, 2)
	}

	var eF, eG, eH Ident
	{
		// will cause trigger TA
		e1 := Event{
			EventTypeId: etC,
			CreatedOn:   time.Now(),
			Data: map[string]interface{}{
				"num":      17,
				"str":      "quick",
				"userName": "my-api-key-38n987",
				"jobId":    "43688858-b6d4-4ef9-a98b-163e1980bba8",
			},
		}
		eF, err = client.PostOneEvent("EventType C", &e1)
		assert.NoError(err)
		defer func() {
			client.DeleteOneEvent("EventType C", eF)
			assert.NoError(err)
		}()

		// will cause trigger TB
		e2 := Event{
			EventTypeId: etD,
			CreatedOn:   time.Now(),
			Data: map[string]interface{}{
				"num":      18,
				"str":      "brown",
				"userName": "my-api-key-38n987",
				"jobId":    "43688858-b6d4-4ef9-a98b-163e1980bba8",
			},
		}
		eG, err = client.PostOneEvent("EventType D", &e2)
		assert.NoError(err)
		defer func() {
			client.DeleteOneEvent("EventType D", eG)
			assert.NoError(err)
		}()

		// will cause no triggers
		e3 := Event{
			EventTypeId: etE,
			CreatedOn:   time.Now(),
			Data: map[string]interface{}{
				"num": 19,
				"str": "fox",
			},
		}
		eH, err = client.PostOneEvent("EventType E", &e3)
		assert.NoError(err)
		defer func() {
			client.DeleteOneEvent("EventType E", eH)
			assert.NoError(err)
		}()
	}

	sleep()

	var aI, aJ Ident
	{
		if MOCKING {
			t.Skip("Skipping test, because mocking")
		}
		alerts, err := client.GetAllAlerts()
		assert.NoError(err)
		assert.Len(*alerts, 2)

		var alert0, alert1 *Alert
		if (*alerts)[0].EventId == eF {
			alert0 = &(*alerts)[0]
			alert1 = &(*alerts)[1]
		} else {
			alert0 = &(*alerts)[1]
			alert1 = &(*alerts)[0]
		}

		aI = alert0.AlertId
		aJ = alert1.AlertId

		assert.EqualValues(alert0.TriggerId, tA)
		assert.EqualValues(alert0.EventId, eF)
		assert.EqualValues(alert1.TriggerId, tB)
		assert.EqualValues(alert1.EventId, eG)

		defer func() {
			client.DeleteOneAlert(aI)
			assert.NoError(err)
		}()
		defer func() {
			client.DeleteOneAlert(aJ)
			assert.NoError(err)
		}()
	}
}
