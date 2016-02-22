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

package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-workflow/common"
	"io/ioutil"
	"net/http"
)

type PzWorkflowService struct {
	name    piazza.ServiceName
	address string
	url     string
}

func NewPzWorkflowService(sys *piazza.System, address string) (*PzWorkflowService, error) {
	var _ common.IWorkflowService = new(PzWorkflowService)
	var _ piazza.IService = new(PzWorkflowService)

	var err error

	service := &PzWorkflowService{
		url:     fmt.Sprintf("http://%s/v1", address),
		name:    piazza.PzWorkflow,
		address: address,
	}

	err = sys.WaitForService(service)
	if err != nil {
		return nil, err
	}

	sys.Services[piazza.PzWorkflow] = service

	return service, nil
}

func (c PzWorkflowService) GetName() piazza.ServiceName {
	return c.name
}

func (c PzWorkflowService) GetAddress() string {
	return c.address
}

//////////////////////////////////////////////////////////////////////////////

func (c *PzWorkflowService) GetFromEventTypes() (*[]common.EventType, error) {
	resp, err := http.Get(c.url + "/eventtypes")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var x []common.EventType
	if len(d) > 0 {
		err = json.Unmarshal(d, &x)
		if err != nil {
			return nil, err
		}
	}

	return &x, nil
}

func (c *PzWorkflowService) GetFromEventType(id common.Ident) (*common.EventType, error) {

	resp, err := http.Get(c.url + "/eventtypes/" + id.String())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var x common.EventType
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	if id != x.ID {
		return nil, errors.New("internal error, eventtype ID mismatch")
	}

	return &x, nil
}

func (c *PzWorkflowService) PostToEventTypes(eventType *common.EventType) (*common.WorkflowIdResponse, error) {

	body, err := json.Marshal(eventType)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.url+"/eventtypes", piazza.ContentTypeJSON, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := new(common.WorkflowIdResponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *PzWorkflowService) DeleteOfEventType(id common.Ident) error {
	resp, err := piazza.HTTPDelete(c.url + "/eventtypes/" + id.String())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}


//////////////////////////////////////////////////////////////////////////////

func (c *PzWorkflowService) GetFromEvents() (*[]common.Event, error) {
	resp, err := http.Get(c.url + "/events")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var x []common.Event
	if len(d) > 0 {
		err = json.Unmarshal(d, &x)
		if err != nil {
			return nil, err
		}
	}

	return &x, nil
}

func (c *PzWorkflowService) GetFromEvent(id common.Ident) (*common.Event, error) {

	resp, err := http.Get(c.url + "/events/" + id.String())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var x common.Event
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	if id != x.ID {
		return nil, errors.New("internal error, event ID mismatch")
	}

	return &x, nil
}

func (c *PzWorkflowService) PostToEvents(event *common.Event) (*common.WorkflowIdResponse, error) {

	body, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.url+"/events", piazza.ContentTypeJSON, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := new(common.WorkflowIdResponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *PzWorkflowService) DeleteOfEvent(id common.Ident) error {
	resp, err := piazza.HTTPDelete(c.url + "/events/" + id.String())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////

func (c *PzWorkflowService) GetFromAlerts() (*[]common.Alert, error) {
	resp, err := http.Get(c.url + "/alerts")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var x []common.Alert
	if len(d) > 0 {
		err = json.Unmarshal(d, &x)
		if err != nil {
			return nil, err
		}
	}

	return &x, nil
}

func (c *PzWorkflowService) GetFromAlert(id common.Ident) (*common.Alert, error) {

	resp, err := http.Get(c.url + "/alerts/" + id.String())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var x common.Alert
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	if id != x.ID {
		return nil, errors.New("internal error, alert ID mismatch")
	}

	return &x, nil
}

func (c *PzWorkflowService) PostToAlerts(event *common.Alert) (*common.WorkflowIdResponse, error) {

	body, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.url+"/alerts", piazza.ContentTypeJSON, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := new(common.WorkflowIdResponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *PzWorkflowService) DeleteOfAlert(id common.Ident) error {
	resp, err := piazza.HTTPDelete(c.url + "/alerts/" + id.String())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////

func (c *PzWorkflowService) GetFromTriggers() (*[]common.Trigger, error) {
	resp, err := http.Get(c.url + "/triggers")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var x []common.Trigger
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	return &x, nil
}

func (c *PzWorkflowService) GetFromTrigger(id common.Ident) (*common.Trigger, error) {
	resp, err := http.Get(c.url + "/triggers/" + id.String())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var x common.Trigger
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	if id != x.ID {
		return nil, errors.New("internal error, trigger ID mismatch")
	}

	return &x, nil
}

func (c *PzWorkflowService) PostToTriggers(trigger *common.Trigger) (*common.WorkflowIdResponse, error) {
	body, err := json.Marshal(trigger)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.url+"/triggers", piazza.ContentTypeJSON, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := new(common.WorkflowIdResponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *PzWorkflowService) DeleteOfTrigger(id common.Ident) error {
	resp, err := piazza.HTTPDelete(c.url + "/triggers/" + id.String())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////

func (c *PzWorkflowService) GetFromAdminStats() (*common.WorkflowAdminStats, error) {

	resp, err := http.Get(c.url + "/admin/stats")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	stats := new(common.WorkflowAdminStats)
	err = json.Unmarshal(data, stats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (c *PzWorkflowService) GetFromAdminSettings() (*common.WorkflowAdminSettings, error) {

	resp, err := http.Get(c.url + "/admin/settings")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	settings := new(common.WorkflowAdminSettings)
	err = json.Unmarshal(data, settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (c *PzWorkflowService) PostToAdminSettings(settings *common.WorkflowAdminSettings) error {

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url+"/admin/settings", piazza.ContentTypeJSON, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
