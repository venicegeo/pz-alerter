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

	config, err := piazza.NewConfig("pz-alerter", mode)
	if err != nil {
		log.Fatal(err)
	}

	sys, err := piazza.NewSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := loggerPkg.NewPzLoggerService(sys, true)
	if err != nil {
		log.Fatal(err)
	}

	uuidgenner, err := uuidgenPkg.NewPzUuidGenService(sys, true)
	if err != nil {
		log.Fatal(err)
	}

	err = server.RunAlertServer(sys, logger, uuidgenner)
	if err != nil {
		log.Fatal(err)
	}

	// not reached
	log.Fatal("not reached")
}
