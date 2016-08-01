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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	pzlogger "github.com/venicegeo/pz-logger/logger"
	pzuuidgen "github.com/venicegeo/pz-uuidgen/uuidgen"
)

type ClientTester struct {
	suite.Suite
	logger  pzlogger.IClient
	uuidgen pzuuidgen.IClient
	client  *Client
	sys     *piazza.SystemConfig
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

	_, err := client.GetStats()
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
	respAlert, err := client.PostAlert(&a1)
	id := respAlert.AlertId
	assert.NoError(err)

	sleep()

	alerts, err := client.GetAllAlerts()
	assert.NoError(err)
	assert.Len(*alerts, 1)
	assert.EqualValues(id, (*alerts)[0].AlertId)
	assert.EqualValues("dummyT1", (*alerts)[0].TriggerId)
	assert.EqualValues("dummyE1", (*alerts)[0].EventId)

	alert, err := client.GetAlert(id)
	assert.NoError(err)
	assert.EqualValues(id, alert.AlertId)

	alert, err = client.GetAlert("nosuchalert1")
	assert.Error(err)

	err = client.DeleteAlert("nosuchalert2")
	assert.Error(err)

	err = client.DeleteAlert(id)
	assert.NoError(err)

	sleep()

	alert, err = client.GetAlert(id)
	assert.Error(err)

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
	respEventType, err := client.PostEventType(eventType)
	etID := respEventType.EventTypeId
	assert.NoError(err)
	defer func() {
		err = client.DeleteEventType(etID)
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
	respEvent, err := client.PostEvent(event)
	eID := respEvent.EventId
	assert.NoError(err)

	defer func() {
		err = client.DeleteEvent(eID)
		assert.NoError(err)
	}()

	//events, err := workflow.GetAllEvents("")
	//assert.NoError(err)
	//assert.Len(*events, 1)
	//assert.EqualValues(eID, (*events)[0].ID)

	sleep()

	tmp, err := client.GetEvent(eID)
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
	respEventType, err := client.PostEventType(eventType)
	id := respEventType.EventTypeId
	defer func() {
		err = client.DeleteEventType(id)
		assert.NoError(err)
	}()

	sleep()

	eventTypes, err := client.GetAllEventTypes()
	assert.NoError(err)
	assert.Len(*eventTypes, 1)
	assert.EqualValues(id, (*eventTypes)[0].EventTypeId)

	tmp, err := client.GetEventType(id)
	assert.NoError(err)
	assert.EqualValues(id, tmp.EventTypeId)
}

func (suite *ClientTester) Test15One() {

	t := suite.T()
	assert := assert.New(t)
	client := suite.client

	assertNoData(suite.T(), client)
	defer assertNoData(suite.T(), client)

	var eventTypeName = "EventTypeA"

	var etID piazza.Ident
	{
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num": elasticsearch.MappingElementTypeInteger,
			"str": elasticsearch.MappingElementTypeString,
		}

		eventType := &EventType{Name: eventTypeName, Mapping: mapping}

		respEventType, err := client.PostEventType(eventType)
		etID = respEventType.EventTypeId
		assert.NoError(err)

		defer func() {
			err := client.DeleteEventType(etID)
			assert.NoError(err)
		}()
	}

	sleep()

	var tID piazza.Ident
	{
		x1 := &Trigger{
			Name: "the x1 trigger",
			Condition: Condition{
				EventTypeIds: []piazza.Ident{etID},
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"num": 17,
						},
					},
				},
			},
			Job: JobRequest{
				CreatedBy: "test",
				JobType: JobType{
					Type: "execute-service",
					Data: map[string]interface{}{
						// "dataInputs": map[string]interface{},
						// "dataOutput": map[string]interface{},
						"serviceId": "ddd5134",
					},
				},
			},
		}

		respTrigger, err := client.PostTrigger(x1)
		tID = respTrigger.TriggerId
		assert.NoError(err)

		defer func() {
			err := client.DeleteTrigger(tID)
			assert.NoError(err)
		}()
	}

	var e1ID piazza.Ident
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

		respEvent1, err := client.PostEvent(e1)
		e1ID = respEvent1.EventId
		assert.NoError(err)

		defer func() {
			err := client.DeleteEvent(e1ID)
			assert.NoError(err)
		}()
	}

	sleep()

	var e2ID piazza.Ident
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

		respEvent2, err := client.PostEvent(e2)
		e2ID = respEvent2.EventId
		assert.NoError(err)

		defer func() {
			err := client.DeleteEvent(e2ID)
			assert.NoError(err)
		}()
	}
	sleep()

	//{
	//	ary, err := client.GetAllEvents("")
	//	assert.NoError(err)
	//	assert.Len(*ary, 2)
	//}

	var aID piazza.Ident
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
			err := client.DeleteAlert(aID)
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
	respEventType, err := client.PostEventType(eventType)
	etID := respEventType.EventTypeId

	defer func() {
		err = client.DeleteEventType(etID)
		assert.NoError(err)
	}()

	t1 := Trigger{
		Name: "the x1 trigger",
		Condition: Condition{
			EventTypeIds: []piazza.Ident{etID},
			Query: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"myint": 17,
					},
				},
			},
		},
		Job: JobRequest{
			CreatedBy: "test",
			JobType: JobType{
				Type: "execute-service",
				Data: map[string]interface{}{
					// "dataInputs": map[string]interface{},
					// "dataOutput": map[string]interface{},
					"serviceId": "ddd5134",
				},
			},
		},
	}
	respTrigger, err := client.PostTrigger(&t1)
	t1ID := respTrigger.TriggerId
	assert.NoError(err)

	defer func() {
		err = client.DeleteTrigger(t1ID)
		assert.NoError(err)
	}()

	sleep()

	tmp, err := client.GetTrigger(t1ID)
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

	var etC, etD, etE piazza.Ident
	{
		mapping := map[string]elasticsearch.MappingElementTypeName{
			"num": elasticsearch.MappingElementTypeInteger,
			"str": elasticsearch.MappingElementTypeString,
		}
		eventTypeC := &EventType{Name: "EventType C", Mapping: mapping}
		eventTypeD := &EventType{Name: "EventType D", Mapping: mapping}
		eventTypeE := &EventType{Name: "EventType E", Mapping: mapping}
		respEventTypeC, err := client.PostEventType(eventTypeC)
		etC = respEventTypeC.EventTypeId
		assert.NoError(err)
		respEventTypeD, err := client.PostEventType(eventTypeD)
		etD = respEventTypeD.EventTypeId
		assert.NoError(err)
		respEventTypeE, err := client.PostEventType(eventTypeE)
		etE = respEventTypeE.EventTypeId
		assert.NoError(err)

		sleep()

		eventTypes, err := client.GetAllEventTypes()
		assert.NoError(err)
		assert.Len(*eventTypes, 3)
	}
	sleep()

	defer func() {
		client.DeleteEventType(etC)
		assert.NoError(err)
		client.DeleteEventType(etD)
		assert.NoError(err)
		client.DeleteEventType(etE)
		assert.NoError(err)
	}()

	////////////////

	var tA, tB piazza.Ident
	{
		t1 := &Trigger{
			Name: "Trigger A",
			Condition: Condition{
				EventTypeIds: []piazza.Ident{etC},
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"str": "quick",
						},
					},
				},
			},
			Job: JobRequest{
				CreatedBy: "test",
				JobType: JobType{
					Type: "execute-service",
					Data: map[string]interface{}{
						// "dataInputs": map[string]interface{},
						// "dataOutput": map[string]interface{},
						"serviceId": "ddd5134",
					},
				},
			},
		}
		respTriggerA, err := client.PostTrigger(t1)
		tA := respTriggerA.TriggerId
		assert.NoError(err)
		defer func() {
			client.DeleteTrigger(tA)
			assert.NoError(err)
		}()

		t2 := &Trigger{
			Name: "Trigger B",
			Condition: Condition{
				EventTypeIds: []piazza.Ident{etD},
				Query: map[string]interface{}{
					"query": map[string]interface{}{
						"match": map[string]interface{}{
							"num": 18,
						},
					},
				},
			},
			Job: JobRequest{
				CreatedBy: "test",
				JobType: JobType{
					Type: "execute-service",
					Data: map[string]interface{}{
						// "dataInputs": map[string]interface{},
						// "dataOutput": map[string]interface{},
						"serviceId": "ddd5134",
					},
				},
			},
		}
		respTriggerB, err := client.PostTrigger(t2)
		tB = respTriggerB.TriggerId
		assert.NoError(err)
		defer func() {
			client.DeleteTrigger(tB)
			assert.NoError(err)
		}()

		sleep()

		triggers, err := client.GetAllTriggers()
		assert.NoError(err)
		assert.Len(*triggers, 2)
	}

	var eF, eG, eH piazza.Ident
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
		respEventF, err := client.PostEvent(&e1)
		eF = respEventF.EventId
		assert.NoError(err)
		defer func() {
			client.DeleteEvent(eF)
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
		respEventG, err := client.PostEvent(&e2)
		eG = respEventG.EventId
		assert.NoError(err)
		defer func() {
			client.DeleteEvent(eG)
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
		respEventH, err := client.PostEvent(&e3)
		eH = respEventH.EventId
		assert.NoError(err)
		defer func() {
			client.DeleteEvent(eH)
			assert.NoError(err)
		}()
	}

	sleep()

	var aI, aJ piazza.Ident
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
			client.DeleteAlert(aI)
			assert.NoError(err)
		}()
		defer func() {
			client.DeleteAlert(aJ)
			assert.NoError(err)
		}()
	}
}
