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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/venicegeo/pz-gocommon"
	logger "github.com/venicegeo/pz-logger/client"
	"github.com/venicegeo/pz-workflow/common"
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

func (c *PzWorkflowService) PostEventType(eventType *common.EventType) (common.Ident, error) {

	body, err := json.Marshal(*eventType)
	if err != nil {
		return common.NoIdent, err
	}

	resp, err := http.Post(c.url+"/eventtypes", piazza.ContentTypeJSON, bytes.NewBuffer(body))
	if err != nil {
		return common.NoIdent, common.NewErrorFromHttp(resp)
	}

	if resp.StatusCode != http.StatusCreated {
		return common.NoIdent, common.NewErrorFromHttp(resp)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return common.NoIdent, err
	}

	result := new(common.WorkflowIdResponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		return common.NoIdent, err
	}

	return result.ID, nil
}

func (c *PzWorkflowService) PostTrigger(trigger *common.Trigger) (common.Ident, error) {
	body, err := json.Marshal(trigger)
	if err != nil {
		return common.NoIdent, err
	}

	resp, err := http.Post(c.url+"/triggers", piazza.ContentTypeJSON, bytes.NewBuffer(body))
	if err != nil {
		return common.NoIdent, common.NewErrorFromHttp(resp)
	}
	if resp.StatusCode != http.StatusCreated {
		return common.NoIdent, common.NewErrorFromHttp(resp)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return common.NoIdent, err
	}

	result := new(common.WorkflowIdResponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		return common.NoIdent, err
	}

	return result.ID, nil
}

func (c *PzWorkflowService) PostEvent(eventTypeName string, event *common.Event) (common.Ident, error) {

	body, err := json.Marshal(event)
	if err != nil {
		return common.NoIdent, err
	}

	resp, err := http.Post(c.url+"/events/"+eventTypeName, piazza.ContentTypeJSON, bytes.NewBuffer(body))
	if err != nil {
		return common.NoIdent, common.NewErrorFromHttp(resp)
	}

	if resp.StatusCode != http.StatusCreated {
		return common.NoIdent, common.NewErrorFromHttp(resp)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return common.NoIdent, err
	}

	result := new(common.WorkflowIdResponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		return common.NoIdent, err
	}

	return result.ID, nil
}

func (c *PzWorkflowService) GetAllAlerts() (*[]common.Alert, error) {
	resp, err := http.Get(c.url + "/alerts")
	if err != nil {
		return nil, common.NewErrorFromHttp(resp)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, common.NewErrorFromHttp(resp)
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

func (c *PzWorkflowService) GetAllEventTypeNames() ([]string, error) {

	typs, err := esi.GetIndexTypes()
	if err != nil {
		return nil, common.NewErrorFromHttp(resp)
	}

	results := make([]string, 0)
	for _, typ := range typs {
		resp, err := http.Get(c.url + "/eventtypes/" + typ)
		if err != nil {
			return nil, common.NewErrorFromHttp(resp)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, common.NewErrorFromHttp(resp)
		}

		d, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var x []string
		if len(d) > 0 {
			err = json.Unmarshal(d, &x)
			if err != nil {
				return nil, err
			}
			for _, vv := range x {
				results = append(results, vv)
			}
		}
	}

	return results, nil
}

func (c *PzWorkflowService) GetAllEvents(typ string) (*[]common.Event, error) {
	if typ != "" {
		typ = "/" + typ
	}
	resp, err := http.Get(c.url + "/events" + typ)
	if err != nil {
		return nil, common.NewErrorFromHttp(resp)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, common.NewErrorFromHttp(resp)
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

func (c *PzWorkflowService) GetAllTriggers() (*[]common.Trigger, error) {
	resp, err := http.Get(c.url + "/triggers")
	if err != nil {
		return nil, common.NewErrorFromHttp(resp)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, common.NewErrorFromHttp(resp)
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

func (c *PzWorkflowService) DeleteEventType(id common.Ident) error {
	resp, err := piazza.HTTPDelete(c.url + "/eventtypes/" + id.String())
	if err != nil {
		return common.NewErrorFromHttp(resp)
	}
	if resp.StatusCode != http.StatusOK {
		return common.NewErrorFromHttp(resp)
	}
	return nil
}

func (c *PzWorkflowService) DeleteEvent(eventTypeName string, id common.Ident) error {
	resp, err := piazza.HTTPDelete(c.url + "/events/" + eventTypeName + "/" + id.String())
	if err != nil {
		return common.NewErrorFromHttp(resp)
	}
	if resp.StatusCode != http.StatusOK {
		return common.NewErrorFromHttp(resp)
	}
	return nil
}

func (c *PzWorkflowService) DeleteAlert(id common.Ident) error {
	resp, err := piazza.HTTPDelete(c.url + "/alerts/" + id.String())
	if err != nil {
		return common.NewErrorFromHttp(resp)
	}
	if resp.StatusCode != http.StatusOK {
		return common.NewErrorFromHttp(resp)
	}
	return nil
}

func (c *PzWorkflowService) DeleteTrigger(id common.Ident) error {
	resp, err := piazza.HTTPDelete(c.url + "/triggers/" + id.String())
	if err != nil {
		return common.NewErrorFromHttp(resp)
	}
	if resp.StatusCode != http.StatusOK {
		return common.NewErrorFromHttp(resp)
	}
	return nil
}

func (c *PzWorkflowService) PostAlert(event *common.Alert) (common.Ident, error) {

	body, err := json.Marshal(event)
	if err != nil {
		return common.NoIdent, err
	}

	resp, err := http.Post(c.url+"/alerts", piazza.ContentTypeJSON, bytes.NewBuffer(body))
	if err != nil {
		return common.NoIdent, common.NewErrorFromHttp(resp)
	}

	if resp.StatusCode != http.StatusCreated {
		return common.NoIdent, common.NewErrorFromHttp(resp)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return common.NoIdent, err
	}

	result := new(common.WorkflowIdResponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		return common.NoIdent, err
	}

	return result.ID, nil
}

func (c *PzWorkflowService) GetAlert(id common.Ident) (*common.Alert, error) {

	resp, err := http.Get(c.url + "/alerts/" + id.String())
	if err != nil {
		return nil, common.NewErrorFromHttp(resp)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, common.NewErrorFromHttp(resp)
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

	return &x, nil
}

func (c *PzWorkflowService) GetEventType(id common.Ident) (*common.EventType, error) {

	resp, err := http.Get(c.url + "/eventtypes/" + id.String())
	if err != nil {
		return nil, common.NewErrorFromHttp(resp)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, common.NewErrorFromHttp(resp)
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

	return &x, nil
}

func (c *PzWorkflowService) GetTrigger(id common.Ident) (*common.Trigger, error) {
	resp, err := http.Get(c.url + "/triggers/" + id.String())
	if err != nil {
		return nil, common.NewErrorFromHttp(resp)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, common.NewErrorFromHttp(resp)
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

	return &x, nil
}

func (c *PzWorkflowService) GetEvent(eventType string, id common.Ident) (*common.Event, error) {

	resp, err := http.Get(c.url + "/events/" + eventType + "/" + id.String())
	if err != nil {
		return nil, common.NewErrorFromHttp(resp)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, common.NewErrorFromHttp(resp)
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

	return &x, nil
}

//////////////////////////////////////////////////////////////////////////////

func (c *PzWorkflowService) GetFromAdminStats() (*common.WorkflowAdminStats, error) {

	resp, err := http.Get(c.url + "/admin/stats")
	if err != nil {
		return nil, common.NewErrorFromHttp(resp)
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
		return nil, common.NewErrorFromHttp(resp)
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
		return common.NewErrorFromHttp(resp)
	}

	if resp.StatusCode != http.StatusOK {
		return common.NewErrorFromHttp(resp)
	}

	return nil
}
