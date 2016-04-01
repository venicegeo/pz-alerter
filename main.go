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

	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
	"github.com/venicegeo/pz-workflow/server"
)

func main() {

	required := []piazza.ServiceName{
		piazza.PzElasticSearch,
		piazza.PzLogger,
		piazza.PzGateway,
	}

	sys, err := piazza.NewSystemConfig(piazza.PzWorkflow, required, false)
	if err != nil {
		log.Fatal(err)
	}

	theLogger, err := loggerPkg.NewMockLoggerService(sys)
	if err != nil {
		log.Fatal(err)
	}
	var tmp = theLogger
	clogger := loggerPkg.NewCustomLogger(&tmp, piazza.PzWorkflow, sys.Address)

	uuidgen, err := uuidgenPkg.NewMockUuidGenService(sys)
	if err != nil {
		log.Fatal(err)
	}

	es, err := elasticsearch.NewClient(sys)
	if err != nil {
		log.Fatal(err)
	}

	clogger.Info("pz-workflow starting...")

	// start server
	routes, err := server.CreateHandlers(sys, clogger, uuidgen, es)
	if err != nil {
		log.Fatal(err)
	}

	_ = sys.StartServer(routes)

	done := sys.StartServer(routes)

	err = <-done
	if err != nil {
		log.Fatal(err)
	}
}
