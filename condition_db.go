package main

import (
	"encoding/json"
	piazza "github.com/venicegeo/pz-gocommon"
	"gopkg.in/olivere/elastic.v3"
)

type ConditionDB struct {
	//data   map[string]Condition
	es *piazza.ElasticSearch
	index  string
}

func newConditionDB(es *piazza.ElasticSearch, index string) (*ConditionDB, error) {
	db := new(ConditionDB)
	db.es = es
	db.index = index

	err := es.MakeIndex(index)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *ConditionDB) write(condition *Condition) error {

	id := newConditionID()
	condition.ID = id

	_, err := db.es.Client.Index().
		Index(db.index).
		Type("condition").
		Id(condition.ID).
		BodyJson(condition).
		Do()
	if err != nil {
		return err
	}

	err = db.es.Flush(db.index)
	if err != nil {
		return err
	}

	return nil
}

func (db *ConditionDB) update(condition *Condition) bool {
/**	_, ok := db.data[condition.ID]
	if ok {
		db.data[condition.ID] = *condition
		return true
	}
	**/
	return false
}

func (db *ConditionDB) readByID(id string) (*Condition, error) {
	termQuery := elastic.NewTermQuery("id", id)
	searchResult, err := db.es.Client.Search().
		Index(db.index).
		Query(termQuery).
		Do()

	if err != nil {
		return nil, err
	}

	for _, hit := range searchResult.Hits.Hits {
		var a Condition
		err := json.Unmarshal(*hit.Source, &a)
		if err != nil {
			return nil, err
		}
		return &a, nil
	}

	return nil, nil
}

func (db *ConditionDB) deleteByID(id string) (bool, error) {
	res, err := db.es.Client.Delete().
		Index(db.index).
		Type("condition").
		Id(id).
		Do()
	if err != nil {
		return false, err
	}

	err = db.es.Flush(db.index)
	if err != nil {
		return false, err
	}

	return res.Found, nil
}

func (db *ConditionDB) getAll() (map[string]Condition, error) {

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

	m := make(map[string]Condition)

	for _, hit := range searchResult.Hits.Hits {
		var t Condition
		err := json.Unmarshal(*hit.Source, &t)
		if err != nil {
			return nil, err
		}
		m[t.ID] = t
	}

	return m, nil
}
