package main

import (
	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-alerter/server"
	"log"
	"os"
)

var pzService *piazza.PzService


func Main(done chan bool, local bool) int {

	var err error

	config, err := piazza.GetConfig("pz-alerter", local)
	if err != nil {
		log.Fatal(err)
		return 1
	}

	err = config.RegisterServiceWithDiscover()
	if err != nil {
		log.Fatal(err)
		return 1
	}

	pzService, err = piazza.NewPzService(config, false)
	if err != nil {
		//pzService.Fatal(err)
		log.Fatal(err)
		return 1
	}

	err = pzService.WaitForService("pz-logger", 1000)
	if err != nil {
		//pzService.Fatal(err)
		log.Fatal(err)
		return 1
	}

	err = pzService.WaitForService("pz-uuidgen", 1000)
	if err != nil {
		//pzService.Fatal(err)
		log.Fatal(err)
		return 1
	}

	if done != nil {
		done <- true
	}

	err = server.RunAlertServer(pzService)
	if err != nil {
		//pzService.Fatal(err)
		log.Fatal(err)
		return 1
	}

	// not reached
	return 1
}

func main() {
	local := piazza.IsLocalConfig()
	os.Exit(Main(nil, local))
}
