package main

import (
	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-alerter/server"
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

	var logger loggerPkg.LoggerClient
	{
		logger, err = loggerPkg.NewPzLoggerClient(sys)
		if err != nil {
			log.Fatal(err)
		}
		err = sys.WaitForService("pz-logger", 1000)
		if err != nil {
			log.Fatal(err)
		}
	}

	var uuidgenner uuidgenPkg.UuidGenClient
	{
		uuidgenner, err = uuidgenPkg.NewPzUuidGenClient(sys)
		if err != nil {
			log.Fatal(err)
		}
		err = sys.WaitForService("pz-uuidgen", 1000)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = server.RunAlertServer(sys, logger, uuidgenner)
	if err != nil {
		log.Fatal(err)
	}

	// not reached
	log.Fatal("not reached")
}
