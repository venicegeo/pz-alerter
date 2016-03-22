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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/venicegeo/pz-gocommon"
	logger "github.com/venicegeo/pz-logger/client"
)

type PzWorkflowService struct {
	name    piazza.ServiceName
	address string
	url     string
	logger  *logger.CustomLogger
}

func NewPzWorkflowService(sys *piazza.System, address string) (*PzWorkflowService, error) {
	var _ piazza.IService = new(PzWorkflowService)

	var err error

	var x piazza.IService = sys.Services[piazza.PzLogger]
	var y logger.ILoggerService = x.(logger.ILoggerService)

	service := &PzWorkflowService{
		url:     fmt.Sprintf("http://%s/v1", address),
		name:    piazza.PzWorkflow,
		address: address,
		logger:  logger.NewCustomLogger(&y, piazza.PzLogger, address),
	}

	err = sys.WaitForService(service)
	if err != nil {
		return nil, err
	}

	sys.Services[piazza.PzWorkflow] = service

	service.logger.Info("PzWorkflowService started")

	return service, nil
}

func (c PzWorkflowService) GetName() piazza.ServiceName {
	return c.name
}

func (c PzWorkflowService) GetAddress() string {
	return c.address
}

//---------------------------------------------------------------------------

type HttpReturn struct {
	Code   int
	Status string
	Data   interface{}
	Err    error
}

func (c *PzWorkflowService) Post(path string, body interface{}) *HttpReturn {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return &HttpReturn{Err: err}
	}

	resp, err := http.Post(c.url+path, piazza.ContentTypeJSON, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return &HttpReturn{Code: resp.StatusCode, Status: resp.Status, Err: err}
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &HttpReturn{Err: err}
	}

	var result interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return &HttpReturn{Err: err}
	}

	return &HttpReturn{Code: resp.StatusCode, Status: resp.Status, Data: result}
}

func (c *PzWorkflowService) Get(path string) *HttpReturn {
	resp, err := http.Get(c.url + path)
	if err != nil {
		return &HttpReturn{Code: resp.StatusCode, Status: resp.Status, Err: err}
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &HttpReturn{Err: err}
	}

	var result interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return &HttpReturn{Err: err}
	}

	return &HttpReturn{Code: resp.StatusCode, Status: resp.Status, Data: result}
}

func (c *PzWorkflowService) Delete(path string) *HttpReturn {
	resp, err := piazza.HTTPDelete(c.url + path)
	if err != nil {
		return &HttpReturn{Code: resp.StatusCode, Status: resp.Status, Err: err}
	}
	return &HttpReturn{Code: resp.StatusCode, Status: resp.Status}
}

//---------------------------------------------------------------------------

func (c *PzWorkflowService) PostOneEventType(eventType *EventType) (Ident, error) {
	ret := c.Post("/eventtypes", eventType)
	if ret.Err != nil {
		return NoIdent, ret.Err
	}
	if ret.Code != http.StatusCreated {
		return NoIdent, errors.New(ret.Status)
	}

	resp2 := &WorkflowIdResponse{}
	err := SuperConvert(ret.Data, resp2)
	if err != nil {
		return NoIdent, err
	}

	return resp2.ID, nil
}

func (c *PzWorkflowService) GetAllEventTypes() (*[]EventType, error) {
	ret := c.Get("/eventtypes")
	if ret.Err != nil {
		return nil, ret.Err
	}
	if ret.Code != http.StatusOK {
		return nil, errors.New(ret.Status)
	}

	var typs []EventType
	err := SuperConvert(ret.Data, &typs)
	if err != nil {
		return nil, err
	}

	return &typs, nil
}

func (c *PzWorkflowService) GetOneEventType(id Ident) (*EventType, error) {
	ret := c.Get("/eventtypes/" + string(id))
	if ret.Err != nil {
		return nil, ret.Err
	}
	if ret.Code != http.StatusOK {
		return nil, errors.New(ret.Status)
	}

	resp2 := EventType{}
	err := SuperConvert(ret.Data, &resp2)
	if err != nil {
		return nil, err
	}

	return &resp2, nil
}

func (c *PzWorkflowService) DeleteOneEventType(id Ident) error {
	ret := c.Delete("/eventtypes/" + string(id))
	if ret.Err != nil {
		return ret.Err
	}
	if ret.Code != http.StatusOK {
		return errors.New(ret.Status)
	}

	return nil
}

//---------------------------------------------------------------------------

func (c *PzWorkflowService) PostOneEvent(eventTypeName string, event *Event) (Ident, error) {
	ret := c.Post("/events/"+eventTypeName, event)
	if ret.Err != nil {
		return NoIdent, ret.Err
	}
	if ret.Code != http.StatusCreated {
		return NoIdent, errors.New(ret.Status)
	}

	resp2 := &WorkflowIdResponse{}
	err := SuperConvert(ret.Data, resp2)
	if err != nil {
		return NoIdent, nil
	}

	return resp2.ID, nil
}

func (c *PzWorkflowService) GetAllEvents(typ string) (*[]Event, error) {
	ret := c.Get("/events")
	if ret.Err != nil {
		return nil, ret.Err
	}
	if ret.Code != http.StatusOK {
		return nil, errors.New(ret.Status)
	}

	var events []Event
	err := SuperConvert(ret.Data, &events)
	if err != nil {
		return nil, err
	}

	return &events, nil
}

func (c *PzWorkflowService) GetOneEvent(eventTypeName string, id Ident) (*Event, error) {
	ret := c.Get("/events/" + eventTypeName + "/" + string(id))
	if ret.Err != nil {
		return nil, errors.New(ret.Status)
	}
	if ret.Code != http.StatusOK {
		return nil, errors.New(ret.Status)
	}

	resp2 := Event{}
	err := SuperConvert(ret.Data, &resp2)
	if err != nil {
		return nil, err
	}

	return &resp2, nil
}

func (c *PzWorkflowService) DeleteOneEvent(eventTypeName string, id Ident) error {
	ret := c.Delete("/events/" + eventTypeName + "/" + string(id))
	if ret.Err != nil {
		return ret.Err
	}
	if ret.Code != http.StatusOK {
		return errors.New(ret.Status)
	}

	return nil
}

//---------------------------------------------------------------------------

func (c *PzWorkflowService) PostOneTrigger(trigger *Trigger) (Ident, error) {
	ret := c.Post("/triggers", trigger)
	if ret.Err != nil {
		return NoIdent, ret.Err
	}
	if ret.Code != http.StatusCreated {
		return NoIdent, errors.New(ret.Status)
	}

	resp2 := &WorkflowIdResponse{}
	err := SuperConvert(ret.Data, resp2)
	if err != nil {
		return NoIdent, err
	}

	return resp2.ID, nil
}

func (c *PzWorkflowService) GetAllTriggers() (*[]Trigger, error) {
	ret := c.Get("/triggers")
	if ret.Err != nil {
		return nil, ret.Err
	}
	if ret.Code != http.StatusOK {
		return nil, errors.New(ret.Status)
	}

	var triggers []Trigger
	err := SuperConvert(ret.Data, &triggers)
	if err != nil {
		return nil, err
	}

	return &triggers, nil
}

func (c *PzWorkflowService) GetOneTrigger(id Ident) (*Trigger, error) {
	ret := c.Get("/triggers/" + string(id))
	if ret.Err != nil {
		return nil, ret.Err
	}
	if ret.Code != http.StatusOK {
		return nil, errors.New(ret.Status)
	}

	resp2 := Trigger{}
	err := SuperConvert(ret.Data, &resp2)
	if err != nil {
		return nil, err
	}

	return &resp2, nil
}

func (c *PzWorkflowService) DeleteOneTrigger(id Ident) error {
	ret := c.Delete("/triggers/" + string(id))
	if ret.Err != nil {
		return ret.Err
	}
	if ret.Code != http.StatusOK {
		return errors.New(ret.Status)
	}

	return nil
}

//---------------------------------------------------------------------------

func (c *PzWorkflowService) PostOneAlert(alert *Alert) (Ident, error) {
	ret := c.Post("/alerts", alert)
	if ret.Err != nil {
		return NoIdent, ret.Err
	}
	if ret.Code != http.StatusCreated {
		return NoIdent, errors.New(ret.Status)
	}

	resp2 := &WorkflowIdResponse{}
	err := SuperConvert(ret.Data, resp2)
	if err != nil {
		return NoIdent, err
	}

	return resp2.ID, nil
}

func (c *PzWorkflowService) GetAllAlerts() (*[]Alert, error) {
	ret := c.Get("/alerts")
	if ret.Err != nil {
		return nil, ret.Err
	}
	if ret.Code != http.StatusOK {
		return nil, errors.New(ret.Status)
	}

	var alerts []Alert
	err := SuperConvert(ret.Data, &alerts)
	if err != nil {
		return nil, err
	}

	return &alerts, nil
}

func (c *PzWorkflowService) GetOneAlert(id Ident) (*Alert, error) {
	ret := c.Get("/alerts/" + string(id))
	if ret.Err != nil {
		return nil, ret.Err
	}
	if ret.Code != http.StatusOK {
		return nil, errors.New(ret.Status)
	}

	resp2 := Alert{}
	err := SuperConvert(ret.Data, &resp2)
	if err != nil {
		return nil, err
	}

	return &resp2, nil
}

func (c *PzWorkflowService) DeleteOneAlert(id Ident) error {
	ret := c.Delete("/alerts/" + string(id))
	if ret.Err != nil {
		return ret.Err
	}
	if ret.Code != http.StatusOK {
		return errors.New(ret.Status)
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////

func (c *PzWorkflowService) GetFromAdminStats() (*WorkflowAdminStats, error) {

	resp, err := http.Get(c.url + "/admin/stats")
	if err != nil {
		return nil, NewErrorFromHttp(resp)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	stats := new(WorkflowAdminStats)
	err = json.Unmarshal(data, stats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (c *PzWorkflowService) GetFromAdminSettings() (*WorkflowAdminSettings, error) {

	resp, err := http.Get(c.url + "/admin/settings")
	if err != nil {
		return nil, NewErrorFromHttp(resp)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	settings := new(WorkflowAdminSettings)
	err = json.Unmarshal(data, settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (c *PzWorkflowService) PostToAdminSettings(settings *WorkflowAdminSettings) error {

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url+"/admin/settings", piazza.ContentTypeJSON, bytes.NewBuffer(data))
	if err != nil {
		return NewErrorFromHttp(resp)
	}

	if resp.StatusCode != http.StatusOK {
		return NewErrorFromHttp(resp)
	}

	return nil
}
