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
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"reflect"
	"regexp"

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
	(*indices)[keyEventTypes].SetMapping(EventTypeDBMapping, "{}")
	(*indices)[keyEvents].SetMapping(EventDBMapping, "{}")
	(*indices)[keyTriggers].SetMapping(TriggerDBMapping, "{}")
	(*indices)[keyAlerts].SetMapping(AlertDBMapping, "{}")
	(*indices)[keyCrons].SetMapping(CronDBMapping, "{}")
	(*indices)[keyTestElasticsearch].SetMapping(TestElasticsearchMapping, "{}")
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
	keyToType := map[string]string{
		keyEventTypes:        EventTypeDBMapping,
		keyEvents:            EventDBMapping,
		keyTriggers:          TriggerDBMapping,
		keyAlerts:            AlertDBMapping,
		keyCrons:             CronDBMapping,
		keyTestElasticsearch: TestElasticsearchMapping,
	}
	indices := make(map[string]elasticsearch.IIndex)

	type ScriptRes struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Mapping string `json:"mapping"`
	}

	format := func(dat []byte) []byte {
		re := regexp.MustCompile(`\r?\n?\t`)
		return []byte(re.ReplaceAllString(string(dat), ""))
	}

	for alias, script := range keyToScript {
		//	alias := keyEventTypes
		//	script := keyToScript[alias]
		log.Println("Running", alias, "init script")
		outDat, err := exec.Command("bash", pwd+"/db/"+script, alias, esURL).Output()
		if err != nil {
			log.Fatalln(err)
		}
		outDat = format(outDat)
		scriptRes := ScriptRes{}
		if err = json.Unmarshal(outDat, &scriptRes); err != nil {
			log.Fatalln(err)
		}
		if scriptRes.Status != "success" {
			log.Fatalf("Script failed: [%s]\n", scriptRes.Message)
		}
		if scriptRes.Message != "" {
			log.Println(scriptRes.Message)
		}
		if indices[alias], err = elasticsearch.NewIndex(sys, alias, ""); err != nil {
			log.Fatalln(err)
		}
		if scriptRes.Mapping != "" {
			inter, err := piazza.StructStringToInterface(scriptRes.Mapping)
			if err != nil {
				log.Fatalln(err)
			}
			var scriptMap, esMap map[string]interface{}
			var ok bool
			if scriptMap, ok = inter.(map[string]interface{}); !ok {
				log.Fatalf("Schema [%s] on alias [%s] in script is not type map[string]interface{}\n", alias, keyToType[alias])
			}
			if inter, err = indices[alias].GetMapping(keyToType[alias]); err != nil {
				log.Fatalln(err)
			}
			if esMap, ok = inter.(map[string]interface{}); !ok {
				log.Fatalf("Schema [%s] on alias [%s] on elasticsearch is not type map[string]interface{}\n", alias, keyToType[alias])
			}
			if !reflect.DeepEqual(scriptMap, esMap) {
				log.Fatalf("Schema [%s] on alias [%s] on elasticsearch does not match the mapping provided\n", alias, keyToType[alias])
			}
		}
	}

	return &indices
}
