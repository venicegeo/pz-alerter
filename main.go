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

package main

import (
	"log"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	pzsyslog "github.com/venicegeo/pz-gocommon/syslog"
	pzuuidgen "github.com/venicegeo/pz-uuidgen/uuidgen"
	pzworkflow "github.com/venicegeo/pz-workflow/workflow"
)

func main() {
	log.Printf("pz-workflow starting...")

	sys, logWriter, auditWriter, uuidgen := makeClients()

	kit, err := pzworkflow.NewKit(sys, logWriter, auditWriter, uuidgen, false)
	if err != nil {
		log.Fatal(err)
	}

	err = kit.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = kit.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func makeClients() (
	*piazza.SystemConfig,
	pzsyslog.Writer,
	pzsyslog.Writer,
	*pzuuidgen.Client) {

	required := []piazza.ServiceName{
		piazza.PzElasticSearch,
		piazza.PzLogger,
		piazza.PzUuidgen,
		piazza.PzKafka,
		piazza.PzServiceController,
		piazza.PzIdam,
	}

	sys, err := piazza.NewSystemConfig(piazza.PzWorkflow, required)
	if err != nil {
		log.Fatal(err)
	}

	loggerIndex, loggerType, auditType, err := pzsyslog.GetRequiredEnvVars()
	if err != nil {
		log.Fatal(err)
	}

	idx, err := elasticsearch.NewIndex(sys, loggerIndex, "")
	if err != nil {
		log.Fatal(err)
	}

	logWriter, auditWriter, err := pzsyslog.GetRequiredESIWriters(idx, loggerType, auditType)
	if err != nil {
		log.Fatal(err)
	}

	stdOutWriter := pzsyslog.StdoutWriter{}

	url, err := sys.GetURL(piazza.PzUuidgen)
	if err != nil {
		log.Fatal(err)
	}
	uuidgen, err := pzuuidgen.NewClient(url, "")
	if err != nil {
		log.Fatal(err)
	}

	return sys, logWriter, pzsyslog.NewMultiWriter([]pzsyslog.Writer{auditWriter, &stdOutWriter}), uuidgen
}
