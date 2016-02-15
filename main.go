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
	"github.com/venicegeo/pz-alerter/server"
	piazza "github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
	"log"
)

func main() {

	var mode piazza.ConfigMode = piazza.ConfigModeCloud
	if piazza.IsLocalConfig() {
		mode = piazza.ConfigModeLocal
	}

	config, err := piazza.NewConfig(piazza.PzAlerter, mode)
	if err != nil {
		log.Fatal(err)
	}

	sys, err := piazza.NewSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := loggerPkg.NewPzLoggerService(sys, sys.DiscoverService.GetDataForService(piazza.PzLogger).Host)
	if err != nil {
		log.Fatal(err)
	}

	uuidgenner, err := uuidgenPkg.NewPzUuidGenService(sys, sys.DiscoverService.GetDataForService(piazza.PzUuidgen).Host)
	if err != nil {
		log.Fatal(err)
	}

	routes, err := server.CreateHandlers(sys, logger, uuidgenner)
	if err != nil {
		log.Fatal(err)
	}

	if len(sys.Services) != 4 {
		log.Fatalf("internal error: services expected (%d) != actual (%d)", 4, len(sys.Services))
	}
	done := sys.StartServer(routes)

	err = <-done
	if err != nil {
		log.Fatal(err)
	}
}
