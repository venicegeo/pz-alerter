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
	"log"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	pzsyslog "github.com/venicegeo/pz-gocommon/syslog"
)

//---------------------------------------------------------------------------

type Kit struct {
	Service       *Service
	Server        *Server
	LogWriter     pzsyslog.Writer
	AuditWriter   pzsyslog.Writer
	Sys           *piazza.SystemConfig
	GenericServer *piazza.GenericServer
	Url           string
	done          chan error
	mocking       bool
	indices       *map[string]elasticsearch.IIndex
}

func NewKit(
	sys *piazza.SystemConfig,
	logWriter pzsyslog.Writer,
	auditWriter pzsyslog.Writer,
	mocking bool,
) (*Kit, error) {

	var err error

	kit := &Kit{}
	kit.Service = &Service{}
	kit.LogWriter = logWriter
	kit.AuditWriter = auditWriter
	kit.Sys = sys
	kit.mocking = mocking

	if kit.mocking {
		kit.indices = kit.makeMockIndices()
	} else {
		kit.indices = kit.makeIndices(sys)
	}

	err = kit.Service.Init(sys, logWriter, auditWriter, kit.indices)
	if err != nil {
		return nil, err
	}

	if !kit.mocking {
		err = kit.Service.InitCron()
		if err != nil {
			log.Fatal(err)
		}
	}

	kit.Server = &Server{}
	err = kit.Server.Init(kit.Service)
	if err != nil {
		return nil, err
	}

	kit.GenericServer = &piazza.GenericServer{Sys: kit.Sys}
	err = kit.GenericServer.Configure(kit.Server.Routes)
	if err != nil {
		return nil, err
	}

	kit.Url = piazza.DefaultProtocol + "://" + kit.GenericServer.Sys.BindTo

	return kit, nil
}

func (kit *Kit) Start() error {
	var err error
	kit.done, err = kit.GenericServer.Start()
	return err
}

func (kit *Kit) Wait() error {
	return <-kit.done
}

func (kit *Kit) Stop() error {
	err := kit.GenericServer.Stop()
	if err != nil {
		return err
	}

	if kit.mocking {
		indices := *kit.indices

		err = indices[keyEventTypes].Delete()
		if err != nil {
			return err
		}

		err = indices[keyEvents].Delete()
		if err != nil {
			return err
		}

		err = indices[keyTriggers].Delete()
		if err != nil {
			return err
		}

		err = indices[keyAlerts].Delete()
		if err != nil {
			return err
		}
	}

	return nil
}

func (kit *Kit) makeMockIndices() *map[string]elasticsearch.IIndex {

	indices := &map[string]elasticsearch.IIndex{
		keyEventTypes:        elasticsearch.NewMockIndex(keyEventTypes),
		keyEvents:            elasticsearch.NewMockIndex(keyEvents),
		keyTriggers:          elasticsearch.NewMockIndex(keyTriggers),
		keyAlerts:            elasticsearch.NewMockIndex(keyAlerts),
		keyCrons:             elasticsearch.NewMockIndex(keyCrons),
		keyTestElasticsearch: elasticsearch.NewMockIndex(keyTestElasticsearch),
	}
	return indices
}

func (kit *Kit) makeIndices(sys *piazza.SystemConfig) *map[string]elasticsearch.IIndex {
	var err error

	indices := make(map[string]elasticsearch.IIndex)

	if indices[keyEventTypes], err = elasticsearch.NewIndex(sys, "eventtypes003", EventTypeIndexSettings); err != nil {
		log.Fatal(err)
	}
	if indices[keyEvents], err = elasticsearch.NewIndex(sys, "events004", EventIndexSettings); err != nil {
		log.Fatal(err)
	}
	if indices[keyTriggers], err = elasticsearch.NewIndex(sys, "triggers003", TriggerIndexSettings); err != nil {
		log.Fatal(err)
	}
	if indices[keyAlerts], err = elasticsearch.NewIndex(sys, "alerts003", AlertIndexSettings); err != nil {
		log.Fatal(err)
	}
	if indices[keyCrons], err = elasticsearch.NewIndex(sys, "crons003", CronIndexSettings); err != nil {
		log.Fatal(err)
	}
	if indices[keyTestElasticsearch], err = elasticsearch.NewIndex(sys, "testelasticsearch003", TestElasticsearchSettings); err != nil {
		log.Fatal(err)
	}

	return &indices
}
