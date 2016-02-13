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

func (c *PzAlerterService) GetFromEvents() (*EventList, error) {
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

	var x EventList
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

func (c *PzAlerterService) GetFromAlerts() (*AlertList, error) {
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

	var x AlertList
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

func (c *PzAlerterService) PostToConditions(cond *Condition) (*AlerterIdResponse, error) {
	body, err := json.Marshal(cond)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.url+"/conditions", piazza.ContentTypeJSON, bytes.NewBuffer(body))
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

func (c *PzAlerterService) GetFromConditions() (*ConditionList, error) {
	resp, err := http.Get(c.url + "/conditions")
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

	var x ConditionList
	if len(d) > 0 {
		err = json.Unmarshal(d, &x)
		if err != nil {
			return nil, err
		}
	}

	return &x, nil
}

func (c *PzAlerterService) GetFromCondition(id Ident) (*Condition, error) {
	resp, err := http.Get(c.url + "/conditions/" + id.String())
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
	var x Condition
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	if id != x.ID {
		return nil, errors.New("internal error, condition ID mismatch")
	}

	return &x, nil
}

func (c *PzAlerterService) DeleteOfCondition(id Ident) error {
	resp, err := piazza.HTTPDelete(c.url + "/conditions/" + id.String())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

func (c *PzAlerterService) PostToActions(action *Action) (*AlerterIdResponse, error) {
	body, err := json.Marshal(action)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.url+"/actions", piazza.ContentTypeJSON, bytes.NewBuffer(body))
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

func (c *PzAlerterService) GetFromActions() (*ActionList, error) {
	resp, err := http.Get(c.url + "/actions")
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

	var x ActionList
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	return &x, nil
}

func (c *PzAlerterService) GetFromAction(id Ident) (*Action, error) {
	resp, err := http.Get(c.url + "/actions/" + id.String())
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

	var x Action
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	if id != x.ID {
		return nil, errors.New("internal error, action ID mismatch")
	}

	return &x, nil
}

func (c *PzAlerterService) DeleteOfAction(id Ident) error {
	resp, err := piazza.HTTPDelete(c.url + "/actions/" + id.String())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

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
