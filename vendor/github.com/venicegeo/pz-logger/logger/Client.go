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

package logger

import (
	"errors"
	"fmt"

	"github.com/venicegeo/pz-gocommon/gocommon"
	syslog "github.com/venicegeo/pz-gocommon/syslog"
)

//---------------------------------------------------------------------

type Client struct {
	url string
	//apiKey         string
	serviceName    piazza.ServiceName
	serviceAddress string
	h              piazza.Http
}

//---------------------------------------------------------------------

func NewClient(sys *piazza.SystemConfig) (*Client, error) {
	var _ IClient = new(Client)

	var err error

	url, err := sys.GetURL(piazza.PzLogger)
	if err != nil {
		return nil, err
	}

	service := &Client{
		url:            url,
		serviceName:    sys.Name,
		serviceAddress: sys.Address,
		h: piazza.Http{
			BaseUrl: url,
			//ApiKey:  apiKey,
			//Preflight:  piazza.SimplePreflight,
			//Postflight: piazza.SimplePostflight,
		},
	}

	err = sys.WaitForService(piazza.PzLogger)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func NewClient2(url string, apiKey string) (*Client, error) {
	var _ IClient = new(Client)

	service := &Client{
		url:            url,
		serviceName:    "notset",
		serviceAddress: "0.0.0.0",
		h: piazza.Http{
			BaseUrl: url,
			ApiKey:  apiKey,
			//Preflight:  preflight,
			//Postflight: postflight,
		},
	}

	return service, nil
}

//---------------------------------------------------------------------

func (c *Client) GetVersion() (*piazza.Version, error) {
	jresp := c.h.PzGet("/version")
	if jresp.IsError() {
		return nil, jresp.ToError()
	}

	var version piazza.Version
	err := jresp.ExtractData(&version)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

//---------------------------------------------------------------------

func (c *Client) GetMessages(
	format *piazza.JsonPagination,
	params *piazza.HttpQueryParams) ([]Message, int, error) {

	formatString := format.String()
	paramString := params.String()

	var ext string
	if formatString != "" && paramString != "" {
		ext = "?" + formatString + "&" + paramString
	} else if formatString == "" && paramString != "" {
		ext = "?" + paramString
	} else if formatString != "" && paramString == "" {
		ext = "?" + formatString
	} else if formatString == "" && paramString == "" {
		ext = ""
	} else {
		return nil, 0, errors.New("Internal error: failed to parse query params")
	}

	endpoint := "/message" + ext

	jresp := c.h.PzGet(endpoint)
	if jresp.IsError() {
		return nil, 0, jresp.ToError()
	}

	var mssgs []Message
	err := jresp.ExtractData(&mssgs)
	if err != nil {
		return nil, 0, err
	}

	return mssgs, jresp.Pagination.Count, nil
}

func (c *Client) GetStats() (*Stats, error) {

	jresp := c.h.PzGet("/admin/stats")
	if jresp.IsError() {
		return nil, jresp.ToError()
	}

	stats := &Stats{}
	err := jresp.ExtractData(stats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

//---------------------------------------------------------------------

// PostMessage puts a new message into Elasticsearch.
func (c *Client) PostMessage(mssg *Message) error {

	err := mssg.Validate()
	if err != nil {
		return fmt.Errorf("message did not validate: %s", err.Error())
	}

	jresp := c.h.PzPost("/message", mssg)
	if jresp.IsError() {
		return jresp.ToError()
	}

	return nil
}

func (c *Client) SetService(name piazza.ServiceName, address string) {
	c.serviceName = name
	c.serviceAddress = address
}

type SyslogElkWriter struct {
	Client IClient
}

func (w *SyslogElkWriter) Write(mNew *syslog.Message) error {
	if w.Client == nil {
		return fmt.Errorf("Log writer client not set")
	}

	severity := SeverityInfo
	switch mNew.Severity {
	case syslog.Debug:
		severity = SeverityDebug
	case syslog.Informational:
		severity = SeverityInfo
	case syslog.Warning:
		severity = SeverityWarning
	case syslog.Error:
		severity = SeverityError
	case syslog.Fatal:
		severity = SeverityFatal
	}

	// translate syslog.Message to a logger.Message and the post it via the client
	mOld := &Message{
		Service:   piazza.ServiceName(mNew.Application),
		Address:   mNew.HostName,
		CreatedOn: mNew.TimeStamp,
		Severity:  severity,
		Message:   mNew.String(),
	}

	err := w.Client.PostMessage(mOld)
	if err != nil {
		return err
	}

	return nil
}
