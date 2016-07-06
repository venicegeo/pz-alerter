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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/venicegeo/pz-gocommon/gocommon"
	logger "github.com/venicegeo/pz-logger/logger"
)

type Client struct {
	url    string
	logger logger.IClient
}

func NewClient(sys *piazza.SystemConfig,
	logger logger.IClient) (*Client, error) {

	var err error

	err = sys.WaitForService(piazza.PzWorkflow)
	if err != nil {
		return nil, err
	}

	url, err := sys.GetURL(piazza.PzWorkflow)

	service := &Client{
		url:    url,
		logger: logger,
	}

	service.logger.Info("Client started")

	return service, nil
}

//---------------------------------------------------------------------------

type HTTPReturn struct {
	StatusCode int
	Status     string
	Data       interface{}
}

func (resp *HTTPReturn) toError() error {
	s := fmt.Sprintf("%d: %s  {%v}", resp.StatusCode, resp.Status, resp.Data)
	return errors.New(s)
}

func (c *Client) Post(path string, body interface{}) (*HTTPReturn, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.url+path, piazza.ContentTypeJSON, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return &HTTPReturn{StatusCode: resp.StatusCode, Status: resp.Status, Data: result}, nil
}

func (c *Client) Get(path string) (*HTTPReturn, error) {
	resp, err := http.Get(c.url + path)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return &HTTPReturn{StatusCode: resp.StatusCode, Status: resp.Status, Data: result}, nil
}

func (c *Client) Delete(path string) (*HTTPReturn, error) {
	resp, err := piazza.HTTPDelete(c.url + path)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return &HTTPReturn{StatusCode: resp.StatusCode, Status: resp.Status, Data: result}, nil
}

//---------------------------------------------------------------------------

func (c *Client) PostOneEventType(eventType *EventType) (Ident, error) {
	resp, err := c.Post("/eventtypes", eventType)
	if err != nil {
		return NoIdent, err
	}
	if resp.StatusCode != http.StatusCreated {
		return NoIdent, resp.toError()
	}

	resp2 := &WorkflowIDResponse{}
	err = SuperConvert(resp.Data, resp2)
	if err != nil {
		return NoIdent, err
	}

	return resp2.ID, nil
}

func (c *Client) GetAllEventTypes() (*[]EventType, error) {
	resp, err := c.Get("/eventtypes")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.toError()
	}

	var typs []EventType
	err = SuperConvert(resp.Data, &typs)
	if err != nil {
		return nil, err
	}

	return &typs, nil
}

func (c *Client) GetOneEventType(id Ident) (*EventType, error) {
	resp, err := c.Get("/eventtypes/" + string(id))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.toError()
	}

	resp2 := EventType{}
	err = SuperConvert(resp.Data, &resp2)
	if err != nil {
		return nil, err
	}

	return &resp2, nil
}

func (c *Client) DeleteOneEventType(id Ident) error {
	resp, err := c.Delete("/eventtypes/" + string(id))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return resp.toError()
	}

	return nil
}

//---------------------------------------------------------------------------

func (c *Client) PostOneEvent(foo string, event *Event) (Ident, error) {
	resp, err := c.Post("/events", event)
	if err != nil {
		return NoIdent, err
	}
	if resp.StatusCode != http.StatusCreated {
		return NoIdent, resp.toError()
	}

	resp2 := &WorkflowIDResponse{}
	err = SuperConvert(resp.Data, resp2)
	if err != nil {
		return NoIdent, err
	}

	return resp2.ID, nil
}

func (c *Client) GetAllEvents(eventTypeName string) (*[]Event, error) {
	//if eventTypeName == "" {
	//	return nil, errors.New("GetAllEvents: type name required")
	//}

	url := "/events"
	if eventTypeName != "" {
		// url += "/" + eventTypeName
		url += "?eventTypeId=" + eventTypeName
	}

	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.toError()
	}

	var events []Event
	err = SuperConvert(resp.Data, &events)
	if err != nil {
		return nil, err
	}

	return &events, nil
}

func (c *Client) GetOneEvent(foo string, id Ident) (*Event, error) {
	resp, err := c.Get("/events/" + string(id))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.toError()
	}

	resp2 := Event{}
	err = SuperConvert(resp.Data, &resp2)
	if err != nil {
		return nil, err
	}

	return &resp2, nil
}

func (c *Client) DeleteOneEvent(foot string, id Ident) error {
	resp, err := c.Delete("/events/" + string(id))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return resp.toError()
	}

	return nil
}

//---------------------------------------------------------------------------

func (c *Client) PostOneTrigger(trigger *Trigger) (Ident, error) {
	resp, err := c.Post("/triggers", trigger)
	if err != nil {
		return NoIdent, err
	}
	if resp.StatusCode != http.StatusCreated {
		return NoIdent, resp.toError()
	}

	resp2 := &WorkflowIDResponse{}
	err = SuperConvert(resp.Data, resp2)
	if err != nil {
		return NoIdent, err
	}

	return resp2.ID, nil
}

func (c *Client) GetAllTriggers() (*[]Trigger, error) {
	resp, err := c.Get("/triggers")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.toError()
	}

	var triggers []Trigger
	err = SuperConvert(resp.Data, &triggers)
	if err != nil {
		return nil, err
	}

	return &triggers, nil
}

func (c *Client) GetOneTrigger(id Ident) (*Trigger, error) {
	resp, err := c.Get("/triggers/" + string(id))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.toError()
	}

	resp2 := Trigger{}
	err = SuperConvert(resp.Data, &resp2)
	if err != nil {
		return nil, err
	}

	return &resp2, nil
}

func (c *Client) DeleteOneTrigger(id Ident) error {
	resp, err := c.Delete("/triggers/" + string(id))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return resp.toError()
	}

	return nil
}

//---------------------------------------------------------------------------

func (c *Client) PostOneAlert(alert *Alert) (Ident, error) {
	resp, err := c.Post("/alerts", alert)
	if err != nil {
		return NoIdent, err
	}
	if resp.StatusCode != http.StatusCreated {
		return NoIdent, resp.toError()
	}

	resp2 := &WorkflowIDResponse{}
	err = SuperConvert(resp.Data, resp2)
	if err != nil {
		return NoIdent, err
	}

	return resp2.ID, nil
}

func (c *Client) GetAllAlerts() (*[]Alert, error) {
	resp, err := c.Get("/alerts")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.toError()
	}

	var alerts []Alert
	err = SuperConvert(resp.Data, &alerts)
	if err != nil {
		return nil, err
	}

	return &alerts, nil
}

func (c *Client) GetOneAlert(id Ident) (*Alert, error) {
	resp, err := c.Get("/alerts/" + string(id))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.toError()
	}

	resp2 := Alert{}
	err = SuperConvert(resp.Data, &resp2)
	if err != nil {
		return nil, err
	}

	return &resp2, nil
}

func (c *Client) DeleteOneAlert(id Ident) error {
	resp, err := c.Delete("/alerts/" + string(id))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return resp.toError()
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////

func (c *Client) GetFromAdminStats() (*WorkflowAdminStats, error) {

	resp, err := c.Get("/admin/stats")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.toError()
	}

	stats := &WorkflowAdminStats{}
	err = SuperConvert(resp.Data, stats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
