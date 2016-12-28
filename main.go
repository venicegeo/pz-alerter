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
	sys, logger, uuidgen := makeSystem()

	log.Printf("pz-workflow starting...")

	indices := makeIndexes(sys)

	workflowServer := makeWorkflow(sys, indices, logger, uuidgen)

	serverLoop(sys, workflowServer)
}

func makeWorkflow(sys *piazza.SystemConfig,
	indices []*elasticsearch.Index,
	logger *pzsyslog.Logger,
	uuidgen *pzuuidgen.Client) *pzworkflow.Server {
	workflowService := &pzworkflow.Service{}
	err := workflowService.Init(
		sys,
		logger,
		uuidgen,
		indices[0],
		indices[1],
		indices[2],
		indices[3],
		indices[4],
		indices[5])
	if err != nil {
		log.Fatal(err)
	}

	err = workflowService.InitCron()
	if err != nil {
		log.Fatal(err)
	}

	workflowServer := &pzworkflow.Server{}
	err = workflowServer.Init(workflowService)
	if err != nil {
		log.Fatal(err)
	}

	return workflowServer
}

func makeSystem() (
	*piazza.SystemConfig,
	*pzsyslog.Logger,
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

	/*logUrl, err := sys.GetURL(piazza.PzLogger)
	if err != nil {
		log.Fatal(err)
	}
	logWriter, err := pzsyslog.NewHttpWriter(logUrl)
	if err != nil {
		log.Fatal(err)
	}*/
	logWriter := &NilWriter{}
	logger := pzsyslog.NewLogger(logWriter, "pz-uuidgen")

	uuidgen, err := pzuuidgen.NewClient(sys)
	if err != nil {
		log.Fatal(err)
	}

	return sys, logger, uuidgen
}

func serverLoop(sys *piazza.SystemConfig,
	workflowServer *pzworkflow.Server) {
	genericServer := piazza.GenericServer{Sys: sys}

	err := genericServer.Configure(workflowServer.Routes)
	if err != nil {
		log.Fatal(err)
	}

	done, err := genericServer.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = <-done
	if err != nil {
		log.Fatal(err)
	}
}

func makeIndexes(sys *piazza.SystemConfig) []*elasticsearch.Index {
	eventtypesIndex, err := elasticsearch.NewIndex(sys, "eventtypes003", pzworkflow.EventTypeIndexSettings)
	if err != nil {
		log.Fatal(err)
	}
	eventsIndex, err := elasticsearch.NewIndex(sys, "events004", pzworkflow.EventIndexSettings)
	if err != nil {
		log.Fatal(err)
	}
	triggersIndex, err := elasticsearch.NewIndex(sys, "triggers003", pzworkflow.TriggerIndexSettings)
	if err != nil {
		log.Fatal(err)
	}
	alertsIndex, err := elasticsearch.NewIndex(sys, "alerts003", pzworkflow.AlertIndexSettings)
	if err != nil {
		log.Fatal(err)
	}
	cronIndex, err := elasticsearch.NewIndex(sys, "crons003", pzworkflow.CronIndexSettings)
	if err != nil {
		log.Fatal(err)
	}
	testElasticsearchIndex, err := elasticsearch.NewIndex(sys, "testelasticsearch003", pzworkflow.TestElasticsearchSettings)
	if err != nil {
		log.Fatal(err)
	}

	ret := []*elasticsearch.Index{
		eventtypesIndex, eventsIndex, triggersIndex,
		alertsIndex, cronIndex, testElasticsearchIndex,
	}
	return ret
}

//-------------------------------
// NilWriter doesn't do anything
type NilWriter struct {
}

func (*NilWriter) Write(*pzsyslog.Message) error {
	return nil
}

func (*NilWriter) Close() error {
	return nil
}
