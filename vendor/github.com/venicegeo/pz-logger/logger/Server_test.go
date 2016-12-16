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
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	"github.com/venicegeo/pz-gocommon/syslog"
)

//---------------------------------------------------------------------

type LoggerTester struct {
	suite.Suite

	esi    elasticsearch.IIndex
	server *piazza.GenericServer

	client    *Client
	syslogger *syslog.Logger
}

func (suite *LoggerTester) setupFixture() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	// make ES index
	{
		suite.esi = elasticsearch.NewMockIndex("loggertest$")
		err = suite.esi.Create("")
		assert.NoError(err)
	}

	// make SystemConfig
	var sys *piazza.SystemConfig
	{
		required := []piazza.ServiceName{}
		sys, err = piazza.NewSystemConfig(piazza.PzLogger, required)
		assert.NoError(err)
	}

	// make backend DB writer
	backendWriter := syslog.NewElasticWriter(suite.esi, LogSchema)

	// make service, server, and generic server
	{
		logWriters := []syslog.Writer{backendWriter}
		auditWriters := []syslog.Writer{}

		service := &Service{}
		err = service.Init(sys, logWriters, auditWriters, suite.esi)
		assert.NoError(err)

		server := &Server{}
		server.Init(service)

		suite.server = &piazza.GenericServer{Sys: sys}

		err = suite.server.Configure(server.Routes)
		assert.NoError(err)

		_, err = suite.server.Start()
		assert.NoError(err)
	}

	// make the client
	var client *Client
	{
		client, err = NewClient(sys)
		assert.NoError(err)
		suite.client = client
	}

	suite.syslogger = syslog.NewLogger(client, "loggertesterapp")
}

func (suite *LoggerTester) teardownFixture() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	// stop server
	{
		err = suite.server.Stop()
		assert.NoError(err)
	}

	// close index
	{
		err = suite.esi.Close()
		assert.NoError(err)

		err = suite.esi.Delete()
		assert.NoError(err)
	}
}

func TestRunSuite(t *testing.T) {
	s := &LoggerTester{}
	suite.Run(t, s)
}

//---------------------------------------------------------------------

func sleep() {
	time.Sleep(1 * time.Second)
}

func (suite *LoggerTester) getLastMessage() string {
	t := suite.T()
	assert := assert.New(t)

	format := piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   "", // ignored by MockClient
		SortBy:  "", // ignored by MockClient
	}
	ms, count, err := suite.client.GetMessages(&format, nil)
	assert.NoError(err)
	assert.True(len(ms) > 0)
	assert.True(count >= len(ms))

	return ms[len(ms)-1].String()
}

//---------------------------------------------------------------------

func (suite *LoggerTester) Test00Time() {
	t := suite.T()
	assert := assert.New(t)

	a := "2006-01-02T15:04:05+07:00"
	b, err := time.Parse(time.RFC3339, a)
	assert.NoError(err)
	c := b.Format(time.RFC3339)
	assert.EqualValues(a, c)
}

func (suite *LoggerTester) Test01Version() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	version, err := suite.client.GetVersion()
	assert.NoError(err)
	assert.EqualValues("1.0.0", version.Version)
	_, _, _, err = piazza.HTTP(piazza.GET, fmt.Sprintf("localhost:%s/version", piazza.LocalPortNumbers[piazza.PzLogger]), piazza.NewHeaderBuilder().AddJsonContentType().GetHeader(), nil)
	assert.NoError(err)
}

func (suite *LoggerTester) Test02Admin() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	stats, err := suite.client.GetStats()
	assert.NoError(err, "GetFromAdminStats")
	assert.NotNil(stats)

	_, _, _, err = piazza.HTTP(piazza.GET, fmt.Sprintf("localhost:%s/admin/stats", piazza.LocalPortNumbers[piazza.PzLogger]), piazza.NewHeaderBuilder().AddJsonContentType().GetHeader(), nil)
	assert.NoError(err)

}

func (suite *LoggerTester) Test03Pagination() {

	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	var err error

	syslogger := suite.syslogger

	err = syslogger.Debug("d")
	assert.NoError(err)
	err = syslogger.Info("i")
	assert.NoError(err)
	err = syslogger.Warning("w")
	assert.NoError(err)
	err = syslogger.Error("e")
	assert.NoError(err)
	err = syslogger.Fatal("f")
	assert.NoError(err)

	sleep()

	format := piazza.JsonPagination{
		PerPage: 1,
		Page:    0,
		SortBy:  "", // ignored by MockClient
		Order:   "", // ignored by MockClient
	}
	ms, count, err := suite.client.GetMessages(&format, nil)
	assert.NoError(err)
	_, _, _, err = piazza.HTTP(piazza.GET, fmt.Sprintf("localhost:%s/syslog?page=0", piazza.LocalPortNumbers[piazza.PzLogger]), piazza.NewHeaderBuilder().AddJsonContentType().GetHeader(), nil)
	assert.NoError(err)

	assert.Len(ms, 1)
	assert.EqualValues(syslog.Debug, ms[0].Severity)
	assert.Equal(5, count)

	format = piazza.JsonPagination{
		PerPage: 5,
		Page:    0,
		SortBy:  "", // ignored by MockClient
		Order:   "", // ignored by MockClient
	}
	ms, count, err = suite.client.GetMessages(&format, nil)
	assert.NoError(err)
	assert.Len(ms, 5)
	assert.EqualValues(syslog.Debug, ms[0].Severity)
	assert.EqualValues(syslog.Fatal, ms[4].Severity)
	assert.Equal(5, count)

	format = piazza.JsonPagination{
		PerPage: 3,
		Page:    1,
		SortBy:  "", // ignored by MockClient
		Order:   "", // ignored by MockClient
	}
	ms, count, err = suite.client.GetMessages(&format, nil)
	assert.NoError(err)
	assert.Len(ms, 2)

	assert.EqualValues(syslog.Error, ms[0].Severity)
	assert.EqualValues(syslog.Fatal, ms[1].Severity)
	assert.Equal(5, count)
}

// this test uses an ES query that is not supported under mocking
/*
func (suite *LoggerTester) Test04OtherParams() {

	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	//client.SetService("myservice", "1.2.3.4")

	sysloggerD := syslog.NewLogger(&syslog.MessageWriter{}, "Dispatcher")
	sysloggerJ := syslog.NewLogger(&syslog.MessageWriter{}, "JobManager")
	sysloggerU := syslog.NewLogger(&syslog.MessageWriter{}, "pz-uuidgen")

	sysloggerD.Info("Received Message to Relay on topic Request-Job with key f3b63085-b482-4ae8-8297-3c7d1fcfff7d")
	sysloggerJ.Info("Processed Update Status for Job 6d0ea538-4382-4ea5-9669-56519b8c8f58 with Status Success")
	sysloggerU.Info("generated 1: 09d4ec60-ea61-4066-8697-5568a47f84bf")
	sysloggerJ.Info("Handling Job with Topic Create-Job for Job ID 09d4ec60-ea61-4066-8697-5568a47f84b")
	sysloggerJ.Info("Handling Job with Topic Update-Job for Job ID be4ce034-1187-4a4f-95a9-a9c31826248b")

	sleep()

	format := piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   "",
		SortBy:  "",
	}

	params := piazza.HttpQueryParams{}
	params.AddString("service", "JobManager")
	params.AddString("contains", "Success")

	msgs, count, err := suite.client.GetMessages(&format, &params)
	assert.NoError(err)
	assert.Len(msgs, 1)
	assert.Equal(5, count)

}
*/

func (suite *LoggerTester) Test05ConstructDsl() {
	t := suite.T()
	assert := assert.New(t)

	format := &piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   piazza.SortOrderDescending,
		SortBy:  "timeStamp",
	}

	startS := "2016-07-26T01:00:00.000Z"
	endS := "2016-07-26T02:00:00.000Z"

	startT, err := time.Parse(time.RFC3339, startS)
	assert.NoError(err)
	endT, err := time.Parse(time.RFC3339, endS)
	assert.NoError(err)

	params := &piazza.HttpQueryParams{}
	params.AddTime("before", startT)
	params.AddTime("after", endT)
	params.AddString("service", "myservice")
	params.AddString("contains", "mycontains")

	actual, err := createQueryDslAsString(format, params)
	assert.NoError(err)
	assert.NotEmpty(actual)

	expected := `
	{
		"from":0,
		"query": {
			"filtered":{ 
				"query":{
					"bool":{
						"must":
						[
							{
								"match":{"service":"myservice"}
							},
							{
								"multi_match":{
									"fields":["address", "message", "service", "severity"],
									"query":"mycontains"
								}
							},
							{
								"range":
								{
									"timeStamp":{
										"gte":"2016-07-26T02:00:00Z", "lte":"2016-07-26T01:00:00Z"
									}
								}
							}
						]
					}
				}
			}
		},
		"size":100, 
		"sort":{"timeStamp":"desc"}
	}`
	assert.JSONEq(expected, actual)
}

func (suite *LoggerTester) Test06Server() {
	t := suite.T()
	assert := assert.New(t)

	service := Service{origin: "yow"}

	resp := service.newInternalErrorResponse(fmt.Errorf("foo"))
	assert.Equal(http.StatusInternalServerError, resp.StatusCode)
	assert.Equal("foo", resp.Message)
	assert.Equal("yow", resp.Origin)

	resp = service.newBadRequestResponse(fmt.Errorf("bar"))
	assert.Equal(http.StatusBadRequest, resp.StatusCode)
	assert.Equal("bar", resp.Message)
	assert.Equal("yow", resp.Origin)
}

func (suite *LoggerTester) Test07GetMessagesErrors() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	format := piazza.JsonPagination{
		PerPage: 1,
		Page:    0,
		SortBy:  "d",
		Order:   "asc",
	}
	mssgs, count, err := suite.client.GetMessages(&format, nil)
	assert.NoError(err)
	assert.Equal(0, count)
	assert.EqualValues([]syslog.Message{}, mssgs)

	format = piazza.JsonPagination{
		PerPage: 9999,
		Page:    9999,
		SortBy:  "",
		Order:   "",
	}
	mssgs, count, err = suite.client.GetMessages(&format, nil)
	assert.NoError(err)
	assert.Equal(0, count)
	assert.EqualValues([]syslog.Message{}, mssgs)
}

func (suite *LoggerTester) Test08Syslog() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	var err error

	syslogger := suite.syslogger

	{
		s := "The quick brown fox"
		err = syslogger.Warning(s)
		assert.NoError(err)
		sleep()
		actual := suite.getLastMessage()
		assert.Contains(actual, s)
		pri := fmt.Sprintf("<%d>%d",
			8*syslog.DefaultFacility+syslog.Warning.Value(), syslog.DefaultVersion)
		assert.Contains(actual, pri)
	}

	{
		s := "The lazy dog"
		syslogger.Error(s)
		sleep()
		actual := suite.getLastMessage()
		assert.Contains(actual, s)
		pri := fmt.Sprintf("<%d>%d",
			8*syslog.DefaultFacility+syslog.Error.Value(), syslog.DefaultVersion)
		assert.Contains(actual, pri)
	}

	{
		stats, err := suite.client.GetStats()
		assert.NoError(err)
		assert.EqualValues(2, stats.NumMessages)
	}

	{
		syslogger.Audit("123", "login", "456", "789")
		sleep()
		actual := suite.getLastMessage()
		assert.Contains(actual, "login")
		pri := fmt.Sprintf("<%d>%d",
			8*syslog.DefaultFacility+syslog.Notice.Value(), syslog.DefaultVersion)
		assert.Contains(actual, pri)
	}
}
