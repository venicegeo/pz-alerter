package client

import (
	"errors"
	piazza "github.com/venicegeo/pz-gocommon"
	"fmt"
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
)

type PzAlerterClient struct {
	url string
}

func NewPzAlerterClient(sys *piazza.System) (*PzAlerterClient, error) {
	c := new(PzAlerterClient)
	c.url = fmt.Sprintf("http://%s/v1", sys.Config.ServerAddress)

	return c, nil
}

func (c *PzAlerterClient) PostToEvents(event *Event) (*AlerterIdResponse, error) {
	body, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.url + "/events", piazza.ContentTypeJSON, bytes.NewBuffer(body))
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

func (c *PzAlerterClient) GetFromEvents() (*EventList, error) {
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
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	return &x, nil
}

func (c *PzAlerterClient) GetFromAlerts() (*AlertList, error) {
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
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	return &x, nil
}

func (c *PzAlerterClient) PostToConditions(cond *Condition) (*AlerterIdResponse, error) {
	body, err := json.Marshal(cond)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.url + "/conditions", piazza.ContentTypeJSON, bytes.NewBuffer(body))
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

func (c *PzAlerterClient) GetFromConditions() (*ConditionList, error) {
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
	err = json.Unmarshal(d, &x)
	if err != nil {
		return nil, err
	}

	return &x, nil
}

func (c *PzAlerterClient) GetFromCondition(id string) (*Condition, error) {
	resp, err := http.Get(c.url + "/conditions/" + id)
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

func (c *PzAlerterClient) DeleteOfCondition(id string) error {
	resp, err := piazza.Delete(c.url + "/conditions/" + id)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}


func (c *PzAlerterClient) GetFromAdminStats() (*AlerterAdminStats, error) {

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

func (c *PzAlerterClient) GetFromAdminSettings() (*AlerterAdminSettings, error) {

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

func (c *PzAlerterClient) PostToAdminSettings(settings *AlerterAdminSettings) error {

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url + "/admin/settings", piazza.ContentTypeJSON, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
