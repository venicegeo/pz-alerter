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
	"net/http"
	"strings"

	"github.com/venicegeo/pz-gocommon/gocommon"
	loggerpkg "github.com/venicegeo/pz-logger/logger"
)

type Client struct {
	url    string
	logger loggerpkg.IClient
}

func NewClient(sys *piazza.SystemConfig, logger loggerpkg.IClient) (*Client, error) {

	var err error

	err = sys.WaitForService(piazza.PzWorkflow)
	if err != nil {
		return nil, err
	}

	url, err := sys.GetURL(piazza.PzWorkflow)
	if err != nil {
		return nil, err
	}

	service := &Client{
		url:    url,
		logger: logger,
	}

	service.logger.Info("Client started")

	return service, nil
}

func NewClient2(url string, apiKey string) (*Client, error) {

	var err error

	loggerURL := strings.Replace(url, "workflow", "logger", 1)
	logger, err := loggerpkg.NewClient2(loggerURL, apiKey)
	if err != nil {
		return nil, err
	}

	service := &Client{
		url:    url,
		logger: logger,
	}

	return service, nil
}

//------------------------------------------------------------------------------

func (c *Client) getObject(endpoint string, out interface{}) error {

	h := piazza.Http{BaseUrl: c.url}
	resp := h.PzGet(endpoint)
	if resp.IsError() {
		return resp.ToError()
	}

	if resp.StatusCode != http.StatusOK {
		return resp.ToError()
	}

	err := resp.ExtractData(out)
	return err
}

func (c *Client) postObject(obj interface{}, endpoint string, out interface{}) error {
	h := piazza.Http{BaseUrl: c.url}

	resp := h.PzPost(endpoint, obj)
	if resp.IsError() {
		return resp.ToError()
	}

	if resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusOK {
		return resp.ToError()
	}

	err := resp.ExtractData(out)
	return err
}

func (c *Client) putObject(obj interface{}, endpoint string, out interface{}) error {
	h := piazza.Http{BaseUrl: c.url}

	resp := h.PzPut(endpoint, obj)
	if resp.IsError() {
		return resp.ToError()
	}

	if resp.StatusCode != http.StatusOK {
		return resp.ToError()
	}

	err := resp.ExtractData(out)
	return err
}

func (c *Client) deleteObject(endpoint string) error {
	h := piazza.Http{BaseUrl: c.url}
	resp := h.PzDelete(endpoint)
	if resp.IsError() {
		return resp.ToError()
	}
	if resp.StatusCode != http.StatusOK {
		return resp.ToError()
	}

	return nil
}

//------------------------------------------------------------------------------

func (c *Client) GetVersion() (*piazza.Version, error) {
	out := &piazza.Version{}
	err := c.getObject("/version", out)
	return out, err
}

//------------------------------------------------------------------------------

func (c *Client) GetEventType(id piazza.Ident) (*EventType, error) {
	out := &EventType{}
	err := c.getObject("/eventType/"+id.String(), out)
	return out, err
}

func (c *Client) GetEventTypeByName(name string) (*EventType, error) {
	out := &EventType{}
	err := c.getObject("/eventType?name="+name, out)
	return out, err
}

func (c *Client) GetAllEventTypes() (*[]EventType, error) {
	out := &[]EventType{}
	err := c.getObject("/eventType", out)
	return out, err
}

func (c *Client) PostEventType(eventType *EventType) (*EventType, error) {
	out := &EventType{}
	err := c.postObject(eventType, "/eventType", out)
	return out, err
}

func (c *Client) QueryEventTypes(query map[string]interface{}) (*[]EventType, error) {
	out := &[]EventType{}
	err := c.postObject(query, "/eventType/query", out)
	return out, err
}

func (c *Client) PutEventType(eventType *EventType) (*EventType, error) {
	out := &EventType{}
	err := c.putObject(eventType, "/eventType", out)
	return out, err
}

func (c *Client) DeleteEventType(id piazza.Ident) error {
	err := c.deleteObject("/eventType/" + id.String())
	return err
}

//------------------------------------------------------------------------------

func (c *Client) GetEvent(id piazza.Ident) (*Event, error) {
	out := &Event{}
	err := c.getObject("/event/"+id.String(), out)
	return out, err
}

func (c *Client) GetAllEvents() (*[]Event, error) {
	out := &[]Event{}
	err := c.getObject("/event", out)
	return out, err
}

func (c *Client) GetAllEventsByEventType(eventTypeID piazza.Ident) (*[]Event, error) {
	out := &[]Event{}
	err := c.getObject("/event?eventTypeId="+eventTypeID.String(), out)
	return out, err
}

func (c *Client) PostEvent(event *Event) (*Event, error) {
	out := &Event{}
	err := c.postObject(event, "/event", out)
	return out, err
}

func (c *Client) QueryEvents(query map[string]interface{}) (*[]Event, error) {
	out := &[]Event{}
	err := c.postObject(query, "/event/query", out)
	return out, err
}

func (c *Client) PutEvent(event *Event) (*Event, error) {
	out := &Event{}
	err := c.putObject(event, "/event", out)
	return out, err
}

func (c *Client) DeleteEvent(id piazza.Ident) error {
	err := c.deleteObject("/event/" + id.String())
	return err
}

//------------------------------------------------------------------------------

func (c *Client) GetTrigger(id piazza.Ident) (*Trigger, error) {
	out := &Trigger{}
	err := c.getObject("/trigger/"+id.String(), out)
	return out, err
}

func (c *Client) GetAllTriggers() (*[]Trigger, error) {
	out := &[]Trigger{}
	err := c.getObject("/trigger", &out)
	return out, err
}

func (c *Client) PostTrigger(trigger *Trigger) (*Trigger, error) {
	out := &Trigger{}
	err := c.postObject(trigger, "/trigger", out)
	return out, err
}

func (c *Client) QueryTriggers(query map[string]interface{}) (*[]Trigger, error) {
	out := &[]Trigger{}
	err := c.postObject(query, "/trigger/query", out)
	return out, err
}

func (c *Client) PutTrigger(trigger *Trigger) (*Trigger, error) {
	out := &Trigger{}
	err := c.putObject(trigger, "/trigger", out)
	return out, err
}

func (c *Client) DeleteTrigger(id piazza.Ident) error {
	err := c.deleteObject("/trigger/" + id.String())
	return err
}

//------------------------------------------------------------------------------

func (c *Client) GetAlert(id piazza.Ident) (*Alert, error) {
	out := &Alert{}
	err := c.getObject("/alert/"+id.String(), out)
	return out, err
}

func (c *Client) GetAlertByTrigger(id piazza.Ident) (*[]Alert, error) {
	out := &[]Alert{}
	err := c.getObject("/alert?triggerId="+id.String(), out)
	return out, err
}

func (c *Client) GetAllAlerts() (*[]Alert, error) {
	out := &[]Alert{}
	err := c.getObject("/alert", out)
	return out, err
}

func (c *Client) PostAlert(alert *Alert) (*Alert, error) {
	out := &Alert{}
	err := c.postObject(alert, "/alert", out)
	return out, err
}

func (c *Client) QueryAlerts(query map[string]interface{}) (*[]Alert, error) {
	out := &[]Alert{}
	err := c.postObject(query, "/alert/query", out)
	return out, err
}

func (c *Client) PutAlert(alert *Alert) (*Alert, error) {
	out := &Alert{}
	err := c.putObject(alert, "/alert", out)
	return out, err
}

func (c *Client) DeleteAlert(id piazza.Ident) error {
	err := c.deleteObject("/alert/" + id.String())
	return err
}

//------------------------------------------------------------------------------

func (c *Client) TestElasticsearchGetVersion() (*string, error) {
	ss := ""
	s := &ss
	err := c.getObject("/_test/elasticsearch/version", s)
	return s, err
}

func (c *Client) TestElasticsearchGetOne(id piazza.Ident) (*TestElasticsearchBody, error) {
	out := &TestElasticsearchBody{}
	err := c.getObject("/_test/elasticsearch/data/"+id.String(), out)
	return out, err
}

func (c *Client) TestElasticsearchPost(body *TestElasticsearchBody) (*TestElasticsearchBody, error) {
	out := &TestElasticsearchBody{}
	err := c.postObject(body, "/_test/elasticsearch/data", out)
	return out, err
}

//------------------------------------------------------------------------------

func (c *Client) GetStats() (*Stats, error) {
	out := &Stats{}
	err := c.getObject("/admin/stats", out)
	return out, err

}
