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
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"

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
	var pwd, esURL string
	{
		pwd, err = os.Getwd()
		if err != nil {
			log.Fatalln(err)
		}
		esURL, err = sys.GetURL(piazza.PzElasticSearch)
		if err != nil {
			log.Fatalln(err)
		}
	}
	keyToScript := map[string]string{
		keyEventTypes:        "000-CreateEventTypeIndex.sh",
		keyEvents:            "000-CreateEventIndex.sh",
		keyTriggers:          "000-CreateTriggerIndex.sh",
		keyAlerts:            "000-CreateAlertIndex.sh",
		keyCrons:             "000-CreateCronIndex.sh",
		keyTestElasticsearch: "000-CreateTestESIndex.sh",
	}
	indices := make(map[string]elasticsearch.IIndex)

	for alias, script := range keyToScript {
		log.Println("Running", alias, "init script")
		outDat, err := exec.Command("bash", pwd+"/db/"+script, alias, esURL).Output()
		if err != nil {
			log.Fatalln(err)
		}
		if !(strings.HasSuffix(string(outDat), "Index already exists\n") || strings.HasSuffix(string(outDat), "Success!\n")) {
			log.Fatalln(errors.New(string(outDat)))
		}
		log.Println(string(outDat))
	}

	if indices[keyEventTypes], err = elasticsearch.NewIndex(sys, keyEventTypes, ""); err != nil {
		log.Fatalln(err)
	}
	if indices[keyEvents], err = elasticsearch.NewIndex(sys, keyEvents, ""); err != nil {
		log.Fatalln(err)
	}
	if indices[keyTriggers], err = elasticsearch.NewIndex(sys, keyTriggers, ""); err != nil {
		log.Fatalln(err)
	}
	if indices[keyAlerts], err = elasticsearch.NewIndex(sys, keyAlerts, ""); err != nil {
		log.Fatalln(err)
	}
	if indices[keyCrons], err = elasticsearch.NewIndex(sys, keyCrons, ""); err != nil {
		log.Fatalln(err)
	}
	if indices[keyTestElasticsearch], err = elasticsearch.NewIndex(sys, keyTestElasticsearch, ""); err != nil {
		log.Fatalln(err)
	}

	return &indices
}
