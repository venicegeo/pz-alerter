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
	pzlogger "github.com/venicegeo/pz-logger/logger"
	pzuuidgen "github.com/venicegeo/pz-uuidgen/uuidgen"
	pzworkflow "github.com/venicegeo/pz-workflow/workflow"
)

func main() {

	required := []piazza.ServiceName{
		piazza.PzElasticSearch,
		piazza.PzLogger,
		piazza.PzUuidgen,
		piazza.PzKafka,
	}

	sys, err := piazza.NewSystemConfig(piazza.PzWorkflow, required)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := pzlogger.NewClient(sys)
	if err != nil {
		log.Fatal(err)
	}

	//	uuidgen, err := uuidgenPkg.NewMockUuidGenService(sys)
	uuidgen, err := pzuuidgen.NewClient(sys)

	if err != nil {
		log.Fatal(err)
	}

	eventtypesIndex, err := elasticsearch.NewIndex(sys, "eventtypes", EventTypeIndexSettings)
	if err != nil {
		log.Fatal(err)
	}
	eventsIndex, err := elasticsearch.NewIndex(sys, "events", EventIndexSettings)
	if err != nil {
		log.Fatal(err)
	}
	triggersIndex, err := elasticsearch.NewIndex(sys, "triggers", TriggerIndexSettings)
	if err != nil {
		log.Fatal(err)
	}
	alertsIndex, err := elasticsearch.NewIndex(sys, "alerts", AlertIndexSettings)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("pz-workflow starting...")

	workflowService := &pzworkflow.WorkflowService{}
	err = workflowService.Init(sys, logger, uuidgen, eventtypesIndex, eventsIndex, triggersIndex, alertsIndex)
	if err != nil {
		log.Fatal(err)
	}
	workflowServer := &pzworkflow.WorkflowServer{}
	err = workflowServer.Init(workflowService)
	if err != nil {
		log.Fatal(err)
	}

	genericServer := piazza.GenericServer{Sys: sys}

	err = genericServer.Configure(workflowServer.Routes)
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
