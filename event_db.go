package main

import (
	"encoding/json"
	"gopkg.in/olivere/elastic.v3"
	piazza "github.com/venicegeo/pz-gocommon"
)

type EventDB struct {
	es *piazza.ElasticSearch
	index  string
}

func newEventDB(es *piazza.ElasticSearch, index string) (*EventDB, error) {
	db := &EventDB{es: es, index: index}

	err := es.MakeIndex(index)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *EventDB) write(event *Event) error {
	id := newEventID()
	event.ID = id

	_, err := db.es.Client.Index().
		Index(db.index).
		Type("event").
		Id(event.ID).
		BodyJson(event).
		Do()
	if err != nil {
		panic(err)
	}

	err = db.es.Flush(db.index)
	if err != nil {
		return err
	}

	return nil
}

func (db *EventDB) getAll() (map[string]Event, error) {

	// search for everything
	// TODO: there's a GET call for this?
	searchResult, err := db.es.Client.Search().
		Index(db.index).
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
