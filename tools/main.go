// Copyright 2017, RadiantBlue Technologies, Inc.
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
	"math"
	"sync"

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	pzsyslog "github.com/venicegeo/pz-gocommon/syslog"
	"github.com/venicegeo/pz-workflow/workflow"
)

var client *workflow.Client

const pageSize = 1000
const maxThreads = 10

type GetNumObjectsF func() (int, error)
type DeletePageOfObjectsF func(int, int) error

func main() {
	var err error

	client, err = makeClient()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		log.Printf("*** ALERTS ***")
		err = deleteAllObjects(client.GetNumAlerts, deletePageOfAlerts)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()
		log.Printf("*** TRIGGERS ***")
		err = deleteAllObjects(client.GetNumTriggers, deletePageOfTriggers)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()
		log.Printf("*** EVENTS ***")
		err = deleteAllObjects(client.GetNumEvents, deletePageOfEvents)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()
		log.Printf("*** EVENT TYPES ***")
		err = deleteAllObjects(client.GetNumEventTypes, deletePageOfEventTypes)
		if err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()
}

func deleteAllObjects(getNumObjectsF GetNumObjectsF, deletePageOfObjectsF DeletePageOfObjectsF) error {
	count, err := getNumObjectsF()
	if err != nil {
		return err
	}
	log.Printf("Num objects: %d", count)

	if count == 0 {
		return nil
	}

	numPages := int(math.Ceil(float64(count) / float64(pageSize)))
	log.Printf("Num pages: %d", numPages)

	delPage := func(wg *sync.WaitGroup, id int) {
		defer wg.Done()
		log.Printf("[%d] started", id)
		err = deletePageOfObjectsF(pageSize, id)
		log.Printf("[%d] ended: %s", id, err)
	}

	for i := 0; i < numPages; i += maxThreads {
		var wg sync.WaitGroup
		wg.Add(maxThreads)
		for j := 0; j < maxThreads; j++ {
			id := i + j
			if id > numPages {
				break
			}
			go delPage(&wg, id)
		}
		wg.Wait()
	}

	return nil
}

func makeClient() (*workflow.Client, error) {
	apiServer, err := piazza.GetApiServer()
	if err != nil {
		return nil, err
	}

	apiKey, err := piazza.GetApiKey(apiServer)
	if err != nil {
		return nil, err
	}

	url := "https://pz-workflow.int.geointservices.io"

	log.Printf("Url: %s", url)
	//log.Printf("Key: %s", apiKey)

	logger := pzsyslog.NewLogger(&pzsyslog.NilWriter{}, &pzsyslog.NilWriter{}, "pz-workflow/tool")
	theClient, err := workflow.NewClient(url, apiKey, logger)
	if err != nil {
		return nil, err
	}

	return theClient, nil
}

func deletePageOfAlerts(perPage int, page int) error {
	id := page

	alerts, err := client.GetAllAlerts(perPage, page)
	if err != nil {
		return err
	}

	tot := len(*alerts)
	log.Printf("[%d] got %d alerts", id, tot)

	for i, alert := range *alerts {
		if i > 0 && i%100 == 0 {
			log.Printf("[%d] deleted %d alerts", id, i)
		}

		err = client.DeleteAlert(alert.AlertID)
		if err != nil {
			log.Printf("[%d] error: %s", id, err)
			return err
		}
	}
	log.Printf("[%d] deleted all %d alerts", id, tot)

	return nil
}

func deletePageOfEvents(perPage int, page int) error {
	id := page

	events, err := client.GetAllEvents(perPage, page)
	if err != nil {
		return err
	}

	tot := len(*events)
	log.Printf("Got %d events", tot)

	for i, event := range *events {
		if i > 0 && i%100 == 0 {
			log.Printf("[%d] Deleted %d events", id, i)
		}

		err = client.DeleteEvent(event.EventID)
		if err != nil {
			// ignore err cases for now
			log.Printf("[%d] error: %s", id, err)
		}
	}
	log.Printf("[%d] Deleted all %d events", id, tot)

	return nil
}

func deletePageOfEventTypes(perPage int, page int) error {
	id := page

	eventtypes, err := client.GetAllEventTypes(perPage, page)
	if err != nil {
		return err
	}

	tot := len(*eventtypes)
	log.Printf("Got %d eventtypes", tot)

	for i, eventtype := range *eventtypes {
		if i > 0 && i%100 == 0 {
			log.Printf("[%d] Deleted %d eventtypes", id, i)
		}

		err = client.DeleteEventType(eventtype.EventTypeID)
		if err != nil {
			// ignore err cases for now
			log.Printf("[%d] error: %s", id, err)
		}
	}
	log.Printf("[%d] Deleted all %d eventtypes", id, tot)

	return nil
}

func deletePageOfTriggers(perPage int, page int) error {
	id := page

	triggers, err := client.GetAllTriggers(perPage, page)
	if err != nil {
		return err
	}

	tot := len(*triggers)
	log.Printf("[%d] got %d of %d triggers", id, tot, perPage)

	for i, trigger := range *triggers {
		if i > 0 && i%100 == 0 {
			log.Printf("[%d] Deleted %d triggers", id, i)
		}

		err = client.DeleteTrigger(trigger.TriggerID)
		if err != nil {
			// ignore err cases for now
			log.Printf("[%d] error: %s", id, err)
		} else {
			log.Printf("[%d] deleted trigger %d", id, i)
		}
	}
	log.Printf("[%d] Deleted all %d triggers", id, tot)

	return nil
}
