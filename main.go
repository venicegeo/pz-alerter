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
	"os"

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	pzsyslog "github.com/venicegeo/pz-gocommon/syslog"
	pzworkflow "github.com/venicegeo/pz-workflow/workflow"
)

func main() {
	log.Printf("pz-workflow starting...")

	sys, logWriter, auditWriter := makeClients()

	pzPen := os.Getenv("PZ_PEN")
	if pzPen == "" {
		log.Fatal("Environment Variable PZ_PEN not found")
	}

	kit, err := pzworkflow.NewKit(sys, logWriter, auditWriter, false, pzPen)
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
	pzsyslog.Writer) {

	required := []piazza.ServiceName{
		piazza.PzElasticSearch,
		piazza.PzKafka,
		piazza.PzServiceController,
		piazza.PzIdam,
	}

	sys, err := piazza.NewSystemConfig(piazza.PzWorkflow, required)
	if err != nil {
		log.Fatal(err)
	}

	loggerIndex, err := pzsyslog.GetRequiredEnvVars()
	if err != nil {
		log.Fatal(err)
	}

	logWriter, auditWriter, err := pzsyslog.GetRequiredWriters(sys, loggerIndex)
	if err != nil {
		log.Fatal(err)
	}

	return sys, logWriter, auditWriter
}
