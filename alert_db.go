package main

import (
	"encoding/json"
	"gopkg.in/olivere/elastic.v3"
	"log"
)

//---------------------------------------------------------------------------

type AlertDB struct {
	client *elastic.Client
	index  string
}

func newAlertDB(client *elastic.Client, index string) (*AlertDB, error) {
	db := new(AlertDB)
	db.client = client
	db.index = index

	err := makeESIndex(client, index)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *AlertDB) write(alert *Alert) error {

	_, err := db.client.Index().
		Index(db.index).
		Type("alert").
		Id(alert.ID).
		BodyJson(alert).
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

func (db *AlertDB) getByConditionID(conditionID string) ([]Alert, error) {
	termQuery := elastic.NewTermQuery("condition_id", conditionID)
	searchResult, err := db.client.Search().
		Index(db.index).  // search in index "twitter"
		Query(termQuery). // specify the query
		Sort("id", true).
		Do() // execute
	if err != nil {
		return nil, err
	}

	if searchResult.Hits == nil {
		return nil, nil
	}

	var as []Alert
	for _, hit := range searchResult.Hits.Hits {
		var a Alert
		err := json.Unmarshal(*hit.Source, &a)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return as, nil
}

func (db *AlertDB) getAll() (map[string]Alert, error) {
	searchResult, err := db.client.Search().
		Index(db.index).
		Query(elastic.NewMatchAllQuery()).
		Sort("id", true).
		Do()
	if err != nil {
		return nil, err
	}

	m := make(map[string]Alert)

	if searchResult.Hits != nil {
		for _, hit := range searchResult.Hits.Hits {
			var t Alert
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				return nil, err
			}
			m[t.ID] = t
		}
	}

	return m, nil
}

func (db *AlertDB) checkConditions(e Event, conditionDB *ConditionDB) error {
	all, err := conditionDB.getAll()
	if err != nil {
		return nil
	}
	for _, cond := range all {
		//log.Printf("e%s.%s ==? c%s.%s", e.ID, e.Type, cond.ID, cond.Type)
		if cond.Type == e.Type {
			a := newAlert(cond.ID, e.ID)
			db.write(&a)
			log.Printf("HIT! event %s has triggered condition %s: alert %s", e.ID, cond.ID, a.ID)
		}
	}
	return nil
}
