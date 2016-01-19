package main

import (
	"encoding/json"
	//"errors"
	"gopkg.in/olivere/elastic.v3"
	//"log"
)

type ConditionDB struct {
	//data   map[string]Condition
	client *elastic.Client
	index  string
}

func newConditionDB(client *elastic.Client, index string) (*ConditionDB, error) {
	db := new(ConditionDB)
	//db.data = make(map[string]Condition)
	db.client = client
	db.index = index

	err := makeESIndex(client, index)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *ConditionDB) write(condition *Condition) error {
	_, err := db.client.Index().
		Index(db.index).
		Type("condition").
		Id(condition.ID).
		BodyJson(condition).
		Do()
	if err != nil {
		panic(err)
	}

	// TODO: how often should we do this?
	_, err = db.client.Flush().Index(db.index).Do()
	if err != nil {
		panic(err)
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
	searchResult, err := db.client.Search().
		Index(db.index).  // search in index "twitter"
		Query(termQuery). // specify the query
		Do()              // execute

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
	res, err := db.client.Delete().
		Index(db.index).
		Type("condition").
		Id(id).
		Do()
	if err != nil {
		return false, err
	}


	// TODO: how often should we do this?
	_, err = db.client.Flush().Index(db.index).Do()
	if err != nil {
		return false, err
	}

	return res.Found, nil
}

func (db *ConditionDB) getAll() (map[string]Condition, error) {

	// search for everything
	// TODO: there's a GET call for this?
	searchResult, err := db.client.Search().
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
