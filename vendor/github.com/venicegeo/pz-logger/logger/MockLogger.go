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
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	"github.com/venicegeo/pz-gocommon/syslog"
)

type MockLoggerKit struct {
	sys    *piazza.SystemConfig
	esi    elasticsearch.IIndex
	server *piazza.GenericServer

	url string

	SysLogger  *syslog.Logger
	esWriter   syslog.WriterReader
	httpWriter *syslog.HttpWriter
}

// MakeMockLogger starts a logger.Server, using a mocked ES backend.
// This is only used for testing.
func NewMockLoggerKit() (*MockLoggerKit, error) {
	var err error

	mock := &MockLoggerKit{}

	// make ES index
	{
		mock.esi = elasticsearch.NewMockIndex("loggertest$")
		err = mock.esi.Create("")
		if err != nil {
			return nil, err
		}
	}

	// make SystemConfig
	{
		required := []piazza.ServiceName{}
		mock.sys, err = piazza.NewSystemConfig(piazza.PzLogger, required)
		if err != nil {
			return nil, err
		}
	}

	// make backend DB writer
	mock.esWriter = syslog.NewElasticWriter(mock.esi, logSchema)

	// make service, server, and generic server
	{
		logWriters := []syslog.Writer{mock.esWriter}
		auditWriters := []syslog.Writer{}

		service := &Service{}
		err = service.Init(mock.sys, logWriters, auditWriters, mock.esi)
		if err != nil {
			return nil, err
		}

		server := &Server{}
		server.Init(service)

		mock.server = &piazza.GenericServer{Sys: mock.sys}

		err = mock.server.Configure(server.Routes)
		if err != nil {
			return nil, err
		}

		_, err = mock.server.Start()
		if err != nil {
			return nil, err
		}
	}

	mock.url = "http://" + mock.sys.BindTo

	// make the client's writer
	mock.httpWriter, err = syslog.NewHttpWriter(mock.url, "")
	if err != nil {
		return nil, err
	}

	// make syslog.Logger
	{
		mock.SysLogger = syslog.NewLogger(mock.httpWriter, "loggertesterapp")
	}

	return mock, nil
}

func (mock *MockLoggerKit) Close() error {
	var err error

	err = mock.httpWriter.Close()
	if err != nil {
		return err
	}

	// stop server
	err = mock.server.Stop()
	if err != nil {
		return err
	}

	err = mock.esWriter.Close()
	if err != nil {
		return err
	}

	// close index
	{
		err = mock.esi.Close()
		if err != nil {
			return err
		}

		err = mock.esi.Delete()
		if err != nil {
			return err
		}
	}

	return nil
}
