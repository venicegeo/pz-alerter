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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"sort"

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

	keyToPattern := map[string]string{
		keyEventTypes:        "EventType",
		keyEvents:            "Event",
		keyTriggers:          "Trigger",
		keyAlerts:            "Alert",
		keyCrons:             "Cron",
		keyTestElasticsearch: "TestES",
	}
	keyToScripts := map[string][]string{
		keyEventTypes:        []string{},
		keyEvents:            []string{},
		keyTriggers:          []string{},
		keyAlerts:            []string{},
		keyCrons:             []string{},
		keyTestElasticsearch: []string{},
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

	{
		files, err := ioutil.ReadDir(pwd + "/db/")
		if err != nil {
			log.Fatalln(err)
		}
		for _, f := range files {
			for alias, pattern := range keyToPattern {
				re, err := regexp.Compile(fmt.Sprintf(`^[0-9]{3}-%s.sh$`, pattern))
				if err != nil {
					log.Fatalln(err)
				}
				if re.MatchString(f.Name()) {
					keyToScripts[alias] = append(keyToScripts[alias], f.Name())
				}
			}
		}
		for _, scripts := range keyToScripts {
			sort.Strings(scripts)
		}
	}

	type ScriptRes struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Mapping string `json:"mapping"`
	}

	format := func(dat []byte) []byte {
		re := regexp.MustCompile(`\r?\n?\t`)
		return bytes.TrimPrefix([]byte(re.ReplaceAllString(string(dat), "")), []byte("\xef\xbb\xbf"))
	}

	log.Println("==========")
	for alias, scripts := range keyToScripts {
		log.Println(alias)
		for i, script := range scripts {
			log.Println(" ", script)
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
			if indices[alias], err = elasticsearch.NewIndex(sys, alias, ""); err != nil {
				log.Fatalln(err)
			}
			if scriptRes.Mapping != "" && i == len(scripts)-1 {
				inter, err := piazza.StructStringToInterface(scriptRes.Mapping)
				if err != nil {
					log.Fatalln(err)
				}
				var scriptMap, esMap map[string]interface{}
				var ok bool
				if scriptMap, ok = inter.(map[string]interface{}); !ok {
					log.Fatalf("Schema [%s] on alias [%s] in script is not type map[string]interface{}\n", keyToType[alias], alias)
				}
				if inter, err = indices[alias].GetMapping(keyToType[alias]); err != nil {
					log.Fatalln(err)
				}
				if esMap, ok = inter.(map[string]interface{}); !ok {
					log.Fatalf("Schema [%s] on alias [%s] on elasticsearch is not type map[string]interface{}\n", keyToType[alias], alias)
				}
				if !reflect.DeepEqual(scriptMap, esMap) {
					log.Fatalf("Schema [%s] on alias [%s] on elasticsearch does not match the mapping provided\n", keyToType[alias], alias)
				}
			}
		}
	}
	log.Println("==========")

	return &indices
}
