package main

import (
	"encoding/json"
	"gopkg.in/olivere/elastic.v3"
	"log"
	"fmt"
)

type EventDB struct {
	//	data map[string]Event
}

func newEventDB() *EventDB {
	db := new(EventDB)
	//	db.data = make(map[string]Event)
	return db
}

func (db *EventDB) write(event *Event) error {

	////////////////////////////////////////////////////////////////////
	// Add a document to the index
	type Tweet struct {
		User string `json:"user"`
		Message  string `json:"message"`
	}

	// Create a client
	client, err := elastic.NewClient()
	if err != nil {
		// Handle error
	}

	// Delete the index, just in case
	_, err = esClient.DeleteIndex("twitter").Do()
	if err != nil {
		// Handle error
		panic(err)
	}
	// Create an index
	_, err = client.CreateIndex("twitter").Do()
	if err != nil {
		// Handle error
		panic(err)
	}

	tweet := Tweet{User: "olivere", Message: "Take Five"}
	_, err = client.Index().
	Index("twitter").
	Type("tweet").
	Id("1").
	BodyJson(tweet).
	Do()
	if err != nil {
		// Handle error
		panic(err)
	}

_, err = client.Flush().Index("twitter").Do()
if err != nil {
panic(err)
}
	// Search with a term query
	termQuery := elastic.NewTermQuery("user", "olivere")
	searchResult, err := client.Search().
	Index("twitter").   // search in index "twitter"
	Query(termQuery).   // specify the query
	Sort("user", true). // sort by "user" field, ascending
	From(0).Size(10).   // take documents 0-9
	Pretty(true).       // pretty print request and response JSON
	Do()                // execute
	if err != nil {
		// Handle error
		panic(err)
	}

	fmt.Printf("Found a total of %d tweets\n", searchResult.TotalHits())
	////////////////////////////////////////////////////////////////////

	event.ID = newEventID()
	//	db.data[event.ID] = *event
	event.Tag = "ttaagg"

	// Add a document to the index
	_, err = esClient.Index().
		Index("events").
		Type("event").
		Id(event.ID).
		BodyString("{\"id\":\"22\"}").
		Do()
	if err != nil {
		// Handle error
		panic(err)
	}

_, err = client.Flush().Index("events").Do()
if err != nil {
panic(err)
}
	return nil
}

func (db *EventDB) getAll() map[string]Event {
	//return db.data

	// Search with a term query
	termQuery := elastic.NewTermQuery("id", "22")
	searchResult, err := esClient.Search().
		Index("events").  // search in index "twitter"
		Query(termQuery). // specify the query
		//Sort("user", true). // sort by "user" field, ascending
		//From(0).Size(10).   // take documents 0-9
		//Pretty(true).       // pretty print request and response JSON
		Do() // execute
	if err != nil {
		// Handle error
		panic(err)
	}

	m := make(map[string]Event)

	if searchResult.Hits != nil {
		log.Printf("Found a total of %d events\n", searchResult.Hits.TotalHits)

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			// hit.Index contains the name of the index

			// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
			var t Event
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				// Deserialization failed
			}

			// Work with tweet
			log.Printf("event: %v\n", t)

			m[t.ID] = t
		}
	} else {
		// No hits
		log.Print("Found no events\n")
	}

	return m
}
