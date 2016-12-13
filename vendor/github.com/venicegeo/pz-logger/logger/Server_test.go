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
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	"github.com/venicegeo/pz-gocommon/syslog"
)

func sleep() {
	time.Sleep(1 * time.Second)
}

type LoggerTester struct {
	suite.Suite

	esi    elasticsearch.IIndex
	sys    *piazza.SystemConfig
	client IClient

	genericServer *piazza.GenericServer
}

func (suite *LoggerTester) SetupSuite() {}

func (suite *LoggerTester) TearDownSuite() {}

func (suite *LoggerTester) setupFixture() {
	t := suite.T()
	assert := assert.New(t)

	var required []piazza.ServiceName
	required = []piazza.ServiceName{}
	sys, err := piazza.NewSystemConfig(piazza.PzLogger, required)
	assert.NoError(err)
	suite.sys = sys

	esi := elasticsearch.NewMockIndex("loggertest$")
	suite.esi = esi

	client, err := NewMockClient()
	assert.NoError(err)
	suite.client = client

	logWriters := []syslog.Writer{&syslog.MessageWriter{}}
	auditWriters := []syslog.Writer{&syslog.MessageWriter{}}

	service := &Service{}
	err = service.Init(sys, logWriters, auditWriters, esi)
	assert.NoError(err)

	server := &Server{}
	server.Init(service)

	suite.genericServer = &piazza.GenericServer{Sys: sys}

	err = suite.genericServer.Configure(server.Routes)
	if err != nil {
		log.Fatal(err)
	}

	_, err = suite.genericServer.Start()
	assert.NoError(err)
}

func (suite *LoggerTester) teardownFixture() {
	err := suite.genericServer.Stop()
	if err != nil {
		panic(err)
	}

	err = suite.esi.Close()
	if err != nil {
		panic(err)
	}

	err = suite.esi.Delete()
	if err != nil {
		panic(err)
	}
}

func TestRunSuite(t *testing.T) {
	s := &LoggerTester{}
	suite.Run(t, s)
}

func (suite *LoggerTester) getLastMessage() string {
	t := suite.T()
	assert := assert.New(t)

	client := suite.client

	format := piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   piazza.SortOrderAscending,
		SortBy:  "createdOn",
	}
	ms, count, err := client.GetMessages(&format, nil)
	assert.NoError(err)
	assert.True(len(ms) > 0)
	assert.True(count >= len(ms))

	return ms[len(ms)-1].String()
}

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

	client := suite.client

	version, err := client.GetVersion()
	assert.NoError(err)
	assert.EqualValues("1.0.0", version.Version)
	_, _, _, err = piazza.HTTP(piazza.GET, fmt.Sprintf("localhost:%s/version", piazza.LocalPortNumbers[piazza.PzLogger]), piazza.NewHeaderBuilder().AddJsonContentType().GetHeader(), nil)
	assert.NoError(err)
}

func (suite *LoggerTester) Test02One() {
}

func (suite *LoggerTester) Test03Help() {
}

func (suite *LoggerTester) Test04Admin() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	_, err := client.GetStats()
	assert.NoError(err, "GetFromAdminStats")
	_, _, _, err = piazza.HTTP(piazza.GET, fmt.Sprintf("localhost:%s/admin/stats", piazza.LocalPortNumbers[piazza.PzLogger]), piazza.NewHeaderBuilder().AddJsonContentType().GetHeader(), nil)
	assert.NoError(err)

}

func (suite *LoggerTester) Test05Pagination() {
	/*
		t := suite.T()
		assert := assert.New(t)

		suite.setupFixture()
		defer suite.teardownFixture()

		client := suite.client

		client.SetService("myservice", "1.2.3.4")

		d := Message{
			Service:   "log-tester",
			Address:   "128.1.2.3",
			CreatedOn: time.Now(),
			Severity:  "Debug",
			Message:   "d",
		}
		i := Message{
			Service:   "log-tester",
			Address:   "128.1.2.3",
			CreatedOn: time.Now(),
			Severity:  "Info",
			Message:   "i",
		}
		w := Message{
			Service:   "log-tester",
			Address:   "128.1.2.3",
			CreatedOn: time.Now(),
			Severity:  "Warn",
			Message:   "w",
		}
		e := Message{
			Service:   "log-tester",
			Address:   "128.1.2.3",
			CreatedOn: time.Now(),
			Severity:  "Error",
			Message:   "e",
		}
		f := Message{
			Service:   "log-tester",
			Address:   "128.1.2.3",
			CreatedOn: time.Now(),
			Severity:  "Fatal",
			Message:   "f",
		}
		client.PostMessage(&d)
		client.PostMessage(&i)
		client.PostMessage(&w)
		client.PostMessage(&e)
		client.PostMessage(&f)
		sleep()

		format := piazza.JsonPagination{
			PerPage: 1,
			Page:    0,
			SortBy:  "createdOn",
			Order:   piazza.SortOrderDescending,
		}
		ms, count, err := client.GetMessages(&format, nil)
		assert.NoError(err)
		_, _, _, err = piazza.HTTP(piazza.GET, fmt.Sprintf("localhost:%s/message?page=0", piazza.LocalPortNumbers[piazza.PzLogger]), piazza.NewHeaderBuilder().AddJsonContentType().GetHeader(), nil)
		assert.NoError(err)

		assert.Len(ms, 1)
		assert.EqualValues(SeverityDebug, ms[0].Severity)
		assert.Equal(5, count)

		format = piazza.JsonPagination{
			PerPage: 5,
			Page:    0,
			SortBy:  "createdOn",
			Order:   piazza.SortOrderAscending,
		}
		ms, count, err = client.GetMessages(&format, nil)
		assert.NoError(err)
		assert.Len(ms, 5)
		assert.EqualValues(SeverityFatal, ms[4].Severity)
		assert.Equal(5, count)

		format = piazza.JsonPagination{
			PerPage: 3,
			Page:    1,
			SortBy:  "createdOn",
			Order:   piazza.SortOrderDescending,
		}
		ms, count, err = client.GetMessages(&format, nil)
		assert.NoError(err)
		assert.Len(ms, 2)

		assert.EqualValues(SeverityError, ms[1].Severity)
		assert.EqualValues(SeverityFatal, ms[0].Severity)
		assert.Equal(5, count)
	*/
}

func (suite *LoggerTester) Test06OtherParams() {
	/*
		t := suite.T()
		assert := assert.New(t)

		suite.setupFixture()
		defer suite.teardownFixture()

		client := suite.client

		client.SetService("myservice", "1.2.3.4")

		sometime := time.Now()

		var testData = []Message{
			{
				Address:   "gnemud7smfv/10.254.0.66",
				Message:   "Received Message to Relay on topic Request-Job with key f3b63085-b482-4ae8-8297-3c7d1fcfff7d",
				Service:   "Dispatcher",
				Severity:  "Info",
				CreatedOn: sometime,
			}, {
				Address:   "gnfbnqsn5m9/10.254.0.14",
				Message:   "Processed Update Status for Job 6d0ea538-4382-4ea5-9669-56519b8c8f58 with Status Success",
				Service:   "JobManager",
				Severity:  "Info",
				CreatedOn: sometime,
			}, {
				Address:   "0.0.0.0",
				Message:   "generated 1: 09d4ec60-ea61-4066-8697-5568a47f84bf",
				Service:   "pz-uuidgen",
				Severity:  "Info",
				CreatedOn: sometime,
			}, {
				Address:   "gnfbnqsn5m9/10.254.0.14",
				Message:   "Handling Job with Topic Create-Job for Job ID 09d4ec60-ea61-4066-8697-5568a47f84bf",
				Service:   "JobManager",
				Severity:  "Info",
				CreatedOn: sometime,
			}, {
				Address:   "gnfbnqsn5m9/10.254.0.14",
				Message:   "Handling Job with Topic Update-Job for Job ID be4ce034-1187-4a4f-95a9-a9c31826248b",
				Service:   "JobManager",
				Severity:  "Info",
				CreatedOn: sometime,
			},
		}

		for _, e := range testData {
			err := client.PostMessage(&e)
			assert.NoError(err)
		}
		httpMessage, _ := json.Marshal(testData[0])
		_, body, _, err := piazza.HTTP(piazza.POST, fmt.Sprintf("localhost:%s/message", piazza.LocalPortNumbers[piazza.PzLogger]), piazza.NewHeaderBuilder().AddJsonContentType().GetHeader(), bytes.NewReader(httpMessage))
		assert.NoError(err)
		println(string(body))
	*/
	/*
		sleep()

		format := piazza.JsonPagination{
			PerPage: 100,
			Page:    0,
			Order:   piazza.SortOrderDescending,
			SortBy:  "createdOn",
		}

		params := piazza.HttpQueryParams{}
		params.AddString("service", "JobManager")
		params.AddString("contains", "Success")

		msgs, count, err := client.GetMessages(&format, &params)
		assert.NoError(err)
		assert.Len(msgs, 1)
		assert.Equal(5, count)
	*/
}

func (suite *LoggerTester) Test07ConstructDsl() {
	t := suite.T()
	assert := assert.New(t)

	format := &piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   piazza.SortOrderDescending,
		SortBy:  "createdOn",
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
									"createdOn":{
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
		"sort":{"createdOn":"desc"}
	}`
	assert.JSONEq(expected, actual)
}

func (suite *LoggerTester) Test08Server() {
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

func (suite *LoggerTester) Test09GetMessagesErrors() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	format := piazza.JsonPagination{
		PerPage: 1,
		Page:    0,
		SortBy:  "id",
		Order:   piazza.SortOrderDescending,
	}
	_, _, err := client.GetMessages(&format, nil)
	assert.Error(err)

	format = piazza.JsonPagination{
		PerPage: 9999,
		Page:    9999,
		SortBy:  "createdOn",
		Order:   piazza.SortOrderDescending,
	}
	mssgs, count, err := client.GetMessages(&format, nil)
	assert.NoError(err)
	assert.Equal(0, count)
	assert.EqualValues([]syslog.Message{}, mssgs)
}

func (suite *LoggerTester) Test10Syslog() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	writer := &SyslogElkWriter{
		Client: suite.client,
	}
	syslogger := syslog.NewLogger(writer, "loggertester")

	{
		s := "The quick brown fox"
		syslogger.Warning(s)
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
