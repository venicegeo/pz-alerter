package main

import (
	"encoding/json"
	"gopkg.in/olivere/elastic.v3"
)

type EventDB struct {
	client *elastic.Client
	index  string
}

func newEventDB(client *elastic.Client, index string) (*EventDB, error) {
	db := &EventDB{client: client, index: index}

	err := makeESIndex(client, index)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *EventDB) write(event *Event) error {

	_, err := db.client.Index().
		Index("events").
		Type("event").
		Id(event.ID).
		BodyJson(event).
		Do()
	if err != nil {
		panic(err)
	}

	// TODO: how often should we do this?
	_, err = db.client.Flush().Index("events").Do()
	if err != nil {
		panic(err)
	}

	return nil
}

func (db *EventDB) getAll() (map[string]Event, error) {

	// search for everything
	// TODO: there's a GET call for this?
	searchResult, err := db.client.Search().
		Index("events").
		Query(elastic.NewMatchAllQuery()).
		Sort("id", true).
		Do()
	if err != nil {
		return nil, err
	}

	m := make(map[string]Event)

	for _, hit := range searchResult.Hits.Hits {
		var t Event
		err := json.Unmarshal(*hit.Source, &t)
		if err != nil {
			return nil, err
		}
		m[t.ID] = t
	}

	return m, nil
}
