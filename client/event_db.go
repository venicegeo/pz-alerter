package client

import (
	"encoding/json"
	piazza "github.com/venicegeo/pz-gocommon"
	"gopkg.in/olivere/elastic.v2"
)

type EventDB struct {
	es    *piazza.ElasticSearchService
	index string
}

func NewEventDB(es *piazza.ElasticSearchService, index string) (*EventDB, error) {
	db := &EventDB{es: es, index: index}

	err := es.MakeIndex(index)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *EventDB) Write(event *Event) error {
	id := NewEventID()
	event.ID = id

	_, err := db.es.Client.Index().
		Index(db.index).
		Type("event").
		Id(event.ID.String()).
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

func (db *EventDB) GetAll() (*EventList, error) {

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

	m := EventList{}

	for _, hit := range searchResult.Hits.Hits {
		var t Event
		err := json.Unmarshal(*hit.Source, &t)
		if err != nil {
			return nil, err
		}
		m[t.ID] = t
	}

	return &m, nil
}
