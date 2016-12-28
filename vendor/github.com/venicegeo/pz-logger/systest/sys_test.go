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

package systest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/gocommon"
	pzsyslog "github.com/venicegeo/pz-gocommon/syslog"
)

func sleep() {
	time.Sleep(1 * time.Second)
}

type LoggerTester struct {
	suite.Suite
	writer     pzsyslog.Writer
	httpWriter *pzsyslog.HttpWriter
	logger     *pzsyslog.Logger
	url        string
	apiKey     string
	apiServer  string
}

func (suite *LoggerTester) setupFixture() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	suite.apiServer, err = piazza.GetApiServer()
	assert.NoError(err)

	i := strings.Index(suite.apiServer, ".")
	assert.NotEqual(1, i)
	host := "pz-logger" + suite.apiServer[i:]
	suite.url = "https://" + host

	suite.apiKey, err = piazza.GetApiKey(suite.apiServer)
	assert.NoError(err)

	suite.httpWriter, err = pzsyslog.NewHttpWriter(suite.url)
	suite.writer = suite.httpWriter
	assert.NoError(err)
	suite.logger = pzsyslog.NewLogger(suite.writer, "loggersystesterapp")
}

func (suite *LoggerTester) teardownFixture() {
}

func TestRunSuite(t *testing.T) {
	s := &LoggerTester{}
	suite.Run(t, s)
}

func (suite *LoggerTester) verifyMessageExists(expected *pzsyslog.Message) bool {
	t := suite.T()
	assert := assert.New(t)

	format := &piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   piazza.SortOrderDescending,
		SortBy:  "createdOn",
	}
	ms, _, err := suite.httpWriter.GetMessages(format, nil)
	fmt.Println("====\n", ms, "\n=====")
	assert.NoError(err)
	assert.Len(ms, format.PerPage)

	for _, m := range ms {
		if m.String() == expected.String() {
			return true
		}
	}
	return false
}

func (suite *LoggerTester) getVersion() (*piazza.Version, error) {
	h := &piazza.Http{BaseUrl: suite.url}

	jresp := h.PzGet("/version")
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

func (suite *LoggerTester) getStats(output interface{}) error {
	h := &piazza.Http{BaseUrl: suite.url}

	jresp := h.PzGet("/admin/stats")
	if jresp.IsError() {
		return jresp.ToError()
	}

	return jresp.ExtractData(output)
}

func (suite *LoggerTester) Test00Version() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	version, err := suite.getVersion()
	assert.NoError(err)
	assert.EqualValues("1.0.0", version.Version)
}

func (suite *LoggerTester) Test01RawGet() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	resp, err := http.Get(suite.url + "/message?perPage=13&page=&0")
	assert.NoError(err)
	assert.True(resp.ContentLength >= 0)
	if resp.ContentLength == -1 {
		assert.FailNow("bonk")
	}
	assert.True(resp.ContentLength > 0)

	raw := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, raw)
	defer func() {
		err = resp.Body.Close()
		assert.NoError(err)
	}()
	if err != nil && err != io.EOF {
		assert.NoError(err)
	}

	assert.Equal(200, resp.StatusCode)
}

func (suite *LoggerTester) Test02RawPost() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	jsn := `
	{
		"address":"XXX",
		"createdOn":"2016-07-22T16:44:58.065583138-04:00",
		"message":"XXX",
		"service":"XXX",
		"severity":"XXX"
	}`
	reader := bytes.NewReader([]byte(jsn))

	resp, err := http.Post(suite.url+"/message",
		piazza.ContentTypeJSON, reader)
	assert.NoError(err)

	raw := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, raw)
	defer func() {
		err = resp.Body.Close()
		assert.NoError(err)
	}()
	if err != nil && err != io.EOF {
		assert.NoError(err)
	}

	//err = json.Unmarshal(raw, output)
	//assett.NoError(err)

	assert.Equal(200, resp.StatusCode)
}

func (suite *LoggerTester) Test03Get() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	format := &piazza.JsonPagination{
		PerPage: 12,
		Page:    0,
		Order:   piazza.SortOrderAscending,
		SortBy:  "createdOn",
	}
	ms, _, err := suite.httpWriter.GetMessages(format, nil)
	assert.NoError(err)
	assert.Len(ms, format.PerPage)
}

func (suite *LoggerTester) Test04Post() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	var err error

	key := time.Now().String()

	data := &pzsyslog.Message{
		Facility:    1,
		Version:     1,
		Process:     "pid1",
		Application: "log-tester",
		HostName:    "128.1.2.3",
		TimeStamp:   time.Now(),
		Severity:    pzsyslog.Notice,
		Message:     key,
	}

	err = suite.writer.Write(data)
	assert.NoError(err, "PostToMessages")

	sleep()

	assert.True(suite.verifyMessageExists(data))
}

func (suite *LoggerTester) xTest05PostHelpers() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	uniq := time.Now().String()
	//TODO client.Info(uniq)

	sleep()

	{
		format := &piazza.JsonPagination{
			PerPage: 100,
			Page:    0,
			Order:   piazza.SortOrderDescending,
			SortBy:  "createdOn",
		}
		ms, _, err := suite.httpWriter.GetMessages(format, nil)
		assert.NoError(err)
		assert.True(len(ms) <= format.PerPage)

		ok := false
		for _, m := range ms {
			if m.Message == uniq {
				ok = true
				break
			}
		}
		assert.True(ok)
	}
}

func (suite *LoggerTester) Test06Admin() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	output := map[string]interface{}{}
	err := suite.getStats(&output)
	assert.NoError(err)
	assert.NotZero(output["numMessages"])
}

func (suite *LoggerTester) Test07Pagination() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	format := &piazza.JsonPagination{
		PerPage: 10,
		Page:    0,
		SortBy:  "createdOn",
		Order:   piazza.SortOrderAscending,
	}
	params := &piazza.HttpQueryParams{}

	// check per-page
	{
		format.PerPage = 17
		ms, _, err := suite.httpWriter.GetMessages(format, params)
		assert.NoError(err)
		assert.Len(ms, 17)
	}

	// check sort order
	{
		format.PerPage = 10
		format.Order = piazza.SortOrderAscending
		ms, _, err := suite.httpWriter.GetMessages(format, params)
		assert.NoError(err)
		last := len(ms) - 1
		assert.True(last <= 9)

		// we can't check "before", because two createdOn's might be the same
		isBefore := ms[0].TimeStamp.Before(ms[last].TimeStamp)
		isEqual := ms[0].TimeStamp.Equal(ms[last].TimeStamp)
		assert.True(isBefore || isEqual)

		format.Order = piazza.SortOrderDescending
		ms, _, err = suite.httpWriter.GetMessages(format, params)
		assert.NoError(err)
		last = len(ms) - 1
		assert.True(last <= 9)

		isAfter := ms[0].TimeStamp.After(ms[last].TimeStamp)
		isEqual = ms[0].TimeStamp.Equal(ms[last].TimeStamp)
		assert.True(isAfter || isEqual)
	}

	// check sort-by
	{
		format.Order = piazza.SortOrderAscending
		format.SortBy = "severity"
		format.PerPage = 100
		format.Page = 0
		ms, _, err := suite.httpWriter.GetMessages(format, params)
		if err != nil {
			panic(88)
		}
		assert.NoError(err)

		last := len(ms) - 1
		for i := 0; i < last; i++ {
			a, b := string(ms[i].Severity), string(ms[i+1].Severity)
			isBefore := (a <= b)
			assert.True(isBefore)
		}
	}
}

func (suite *LoggerTester) xTest08Params() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	uniqService := strconv.Itoa(time.Now().Nanosecond())
	uniqDebug := strconv.Itoa(time.Now().Nanosecond() * 3)
	//uniqError := strconv.Itoa(time.Now().Nanosecond() * 5)
	//uniqFatal := strconv.Itoa(time.Now().Nanosecond() * 7)

	now := time.Now()
	sec3 := time.Second * 3
	tstart := now.Add(-sec3)

	//TODO client.Debug(uniqDebug)
	//TODO client.Error(uniqError)
	//TODO client.Fatal(uniqFatal)

	sleep()

	format := &piazza.JsonPagination{
		PerPage: 256,
		Page:    0,
		Order:   piazza.SortOrderDescending,
		SortBy:  "createdOn",
	}

	// test date range params
	{
		tend := now.Add(sec3)

		params := &piazza.HttpQueryParams{}
		params.AddTime("after", tstart)
		params.AddTime("before", tend)

		msgs, cnt, err := suite.httpWriter.GetMessages(format, params)

		assert.NoError(err)
		assert.True(cnt >= 3)
		assert.True(len(msgs) >= 3)
	}

	// test service param
	{
		params := &piazza.HttpQueryParams{}
		params.AddString("service", uniqService)

		msgs, _, err := suite.httpWriter.GetMessages(format, params)

		assert.NoError(err)
		assert.Len(msgs, 3)
	}

	// test contains param
	{
		params := &piazza.HttpQueryParams{}
		params.AddString("contains", uniqDebug)

		msgs, _, err := suite.httpWriter.GetMessages(format, params)

		assert.NoError(err)
		assert.True(len(msgs) == 1)
	}
}
