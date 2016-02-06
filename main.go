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

	logger, err := loggerPkg.NewPzLoggerService(sys)
	if err != nil {
		log.Fatal(err)
	}

	uuidgenner, err := uuidgenPkg.NewPzUuidGenService(sys)
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
