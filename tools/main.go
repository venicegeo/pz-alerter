package main

import (
	"log"
	"math"
	"sync"

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	"github.com/venicegeo/pz-workflow/workflow"
)

var client *workflow.Client

const pageSize = 1000

type GetNumObjectsF func() (int, error)
type DeletePageOfObjectsF func(int, int) error

func main() {
	var err error

	client, err = makeClient()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("*** ALERTS ***")
	err = deleteAllObjects(client.GetNumAlerts, deletePageOfAlerts)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("*** TRIGGERS ***")
	err = deleteAllObjects(client.GetNumTriggers, deletePageOfTriggers)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("*** EVENTS ***")
	err = deleteAllEvents(client.GetNumEvents, deletePageOfEvents)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("*** EVENT TYPES ***")
	err = deleteAllEventTypes(client.GetNumEventTypes, deletePageOfEventTypes)
	if err != nil {
		log.Fatal(err)
	}
}

func deleteAllObjects(getNumObjectsF GetNumObjectsF, deletePageOfObjectsF DeletePageOfObjectsF) error {
	for {
		count, err := getNumObjectsF()
		if err != nil {
			return err
		}
		log.Printf("Num objects: %d", count)
		if count < pageSize {
			break
		}

		// let's keep a few around, just for kicks
		count -= pageSize

		numThreads := int(math.Ceil(float64(count / pageSize)))

		var wg sync.WaitGroup
		wg.Add(numThreads)

		for id := 0; id < numThreads; id++ {
			go func(id int) {
				defer wg.Done()
				_ = deletePageOfObjectsF(pageSize, id)
			}(id)
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

	url := "https://" + apiServer
	url = "https://pz-workflow.int.geointservices.io"

	log.Printf("Url: %s", url)
	log.Printf("Key: %s", apiKey)

	client, err := workflow.NewClient2(url, apiKey)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func deletePageOfAlerts(perPage int, page int) error {
	id := page

	alerts, err := client.GetAllAlerts(perPage, page)
	if err != nil {
		return err
	}

	tot := len(*alerts)
	//log.Printf("Got %d alerts", tot)
	if tot != perPage {
		return nil
	}

	for i, alert := range *alerts {
		if i > 0 && i%100 == 0 {
			log.Printf("[%d] Deleted %d alerts", id, i)
		}

		err = client.DeleteAlert(alert.AlertID)
		if err != nil {
			// ignore err cases for now
			//log.Printf("error %#v", err)
		}
	}
	log.Printf("[%d] Deleted all %d alerts", id, tot)

	return nil
}

func deletePageOfEvents(perPage int, page int) error {
	id := page

	events, err := client.GetAllEvents(perPage, page)
	if err != nil {
		return err
	}

	tot := len(*events)
	//log.Printf("Got %d events", tot)
	if tot != perPage {
		return nil
	}

	for i, event := range *events {
		if i > 0 && i%100 == 0 {
			log.Printf("[%d] Deleted %d events", id, i)
		}

		err = client.DeleteEvent(event.EventID)
		if err != nil {
			// ignore err cases for now
			//log.Printf("error %#v", err)
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
	//log.Printf("Got %d eventtypes", tot)
	if tot != perPage {
		return nil
	}

	for i, eventtype := range *eventtypes {
		if i > 0 && i%100 == 0 {
			log.Printf("[%d] Deleted %d eventtypes", id, i)
		}

		err = client.DeleteEventType(eventtype.EventTypeID)
		if err != nil {
			// ignore err cases for now
			//log.Printf("error %#v", err)
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
	//log.Printf("Got %d triggers", tot)
	if tot != perPage {
		return nil
	}

	for i, trigger := range *triggers {
		if i > 0 && i%100 == 0 {
			log.Printf("[%d] Deleted %d triggers", id, i)
		}

		err = client.DeleteTrigger(trigger.TriggerID)
		if err != nil {
			// ignore err cases for now
			//log.Printf("error %#v", err)
		}
	}
	log.Printf("[%d] Deleted all %d triggers", id, tot)

	return nil
}
