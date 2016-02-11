package client

import (
	"encoding/json"
	piazza "github.com/venicegeo/pz-gocommon"
)

type EventDB struct {
	es    *piazza.ElasticSearchService
	index string
}

func NewEventDB(es *piazza.ElasticSearchService, index string) (*EventDB, error) {
	db := &EventDB{es: es, index: index}

	err := es.DeleteIndex(index)
	if err != nil {
		return nil, err
	}

	err = es.CreateIndex(index)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *EventDB) Write(event *Event) error {
	id := NewEventID()
	event.ID = id

	_, err := db.es.PostData(db.index, "event", event.ID.String(), event)
	if err != nil {
		return err
	}

	err = db.es.FlushIndex(db.index)
	if err != nil {
		return err
	}

	return nil
}

func (db *EventDB) GetAll() (*EventList, error) {

	searchResult, err := db.es.SearchByMatchAll(db.index)
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

func (db *EventDB) DeleteByID(id string) (bool, error) {
	res, err := db.es.DeleteById(db.index, "event", id)
	if err != nil {
		return false, err
	}

	err = db.es.FlushIndex(db.index)
	if err != nil {
		return false, err
	}

	return res.Found, nil
}
