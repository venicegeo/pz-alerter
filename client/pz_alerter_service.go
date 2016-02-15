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
	"io/ioutil"
	"net/http"
)

type PzAlerterService struct {
	name    piazza.ServiceName
	address string
	url     string
}

func NewPzAlerterService(sys *piazza.System, address string) (*PzAlerterService, error) {
	var _ IAlerterService = new(PzAlerterService)
	var _ piazza.IService = new(PzAlerterService)

	var err error

	service := &PzAlerterService{
		url:     fmt.Sprintf("http://%s/v1", address),
		name:    piazza.PzAlerter,
		address: address,
	}

	err = sys.WaitForService(service)
	if err != nil {
		return nil, err
	}

	sys.Services[piazza.PzAlerter] = service

	return service, nil
}

func (c PzAlerterService) GetName() piazza.ServiceName {
	return c.name
}

func (c PzAlerterService) GetAddress() string {
	return c.address
}

//////////////////////////////////////////////////////////////////////////////

func (c *PzAlerterService) PostToEvents(event *Event) (*AlerterIdResponse, error) {

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

	result := new(AlerterIdResponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *PzAlerterService) GetFromEvents() (*[]Event, error) {
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

	var x []Event
	if len(d) > 0 {
		err = json.Unmarshal(d, &x)
		if err != nil {
			return nil, err
		}
	}

	return &x, nil
}

func (c *PzAlerterService) DeleteOfEvent(id Ident) error {
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

func (c *PzAlerterService) GetFromAlerts() (*[]Alert, error) {
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

	var x []Alert
	if len(d) > 0 {
		err = json.Unmarshal(d, &x)
		if err != nil {
			return nil, err
		}
	}

	return &x, nil
}

func (c *PzAlerterService) PostToAlerts(event *Alert) (*AlerterIdResponse, error) {

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

	result := new(AlerterIdResponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *PzAlerterService) GetFromAlert(id Ident) (*Alert, error) {

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
	var x Alert
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	if id != x.ID {
		return nil, errors.New("internal error, alert ID mismatch")
	}

	return &x, nil
}

func (c *PzAlerterService) DeleteOfAlert(id Ident) error {
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

func (c *PzAlerterService) PostToTriggers(trigger *Trigger) (*AlerterIdResponse, error) {
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

	result := new(AlerterIdResponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *PzAlerterService) GetFromTriggers() (*[]Trigger, error) {
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

	var x []Trigger
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	return &x, nil
}

func (c *PzAlerterService) GetFromTrigger(id Ident) (*Trigger, error) {
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

	var x Trigger
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	if id != x.ID {
		return nil, errors.New("internal error, trigger ID mismatch")
	}

	return &x, nil
}

func (c *PzAlerterService) DeleteOfTrigger(id Ident) error {
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

func (c *PzAlerterService) GetFromAdminStats() (*AlerterAdminStats, error) {

	resp, err := http.Get(c.url + "/admin/stats")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	stats := new(AlerterAdminStats)
	err = json.Unmarshal(data, stats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (c *PzAlerterService) GetFromAdminSettings() (*AlerterAdminSettings, error) {

	resp, err := http.Get(c.url + "/admin/settings")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	settings := new(AlerterAdminSettings)
	err = json.Unmarshal(data, settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (c *PzAlerterService) PostToAdminSettings(settings *AlerterAdminSettings) error {

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
