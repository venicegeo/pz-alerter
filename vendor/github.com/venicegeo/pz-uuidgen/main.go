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

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	pzsyslog "github.com/venicegeo/pz-gocommon/syslog"
	pzuuidgen "github.com/venicegeo/pz-uuidgen/uuidgen"
)

func main() {

	required := []piazza.ServiceName{
		piazza.PzElasticSearch,
		piazza.PzLogger,
	}

	sys, err := piazza.NewSystemConfig(piazza.PzUuidgen, required)
	if err != nil {
		log.Fatal(err)
	}

	logUrl, err := sys.GetURL(piazza.PzLogger)
	if err != nil {
		log.Fatal(err)
	}
	logWriter, err := pzsyslog.NewHttpWriter(logUrl, "")
	if err != nil {
		log.Fatal(err)
	}
	logger := pzsyslog.NewLogger(logWriter, "pz-uuidgen")

	service := &pzuuidgen.Service{}
	err = service.Init(sys, logger)
	if err != nil {
		log.Fatal(err)
	}

	server := &pzuuidgen.Server{}
	server.Init(service)

	genericServer := &piazza.GenericServer{Sys: sys}
	err = genericServer.Configure(server.Routes)
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
