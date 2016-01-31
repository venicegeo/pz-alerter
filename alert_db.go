package main

import (
	"encoding/json"
	"gopkg.in/olivere/elastic.v3"
	piazza "github.com/venicegeo/pz-gocommon"
	"fmt"
)

//---------------------------------------------------------------------------

type AlertDB struct {
	es *piazza.ElasticSearch
	index  string
}

func newAlertDB(es *piazza.ElasticSearch, index string) (*AlertDB, error) {
	db := new(AlertDB)
	db.es = es
	db.index = index

	err := es.MakeIndex(index)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *AlertDB) write(alert *piazza.Alert) error {

	_, err := db.es.Client.Index().
		Index(db.index).
		Type("alert").
		Id(alert.ID).
		BodyJson(alert).
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

func (db *AlertDB) getByConditionID(conditionID string) ([]piazza.Alert, error) {
	termQuery := elastic.NewTermQuery("condition_id", conditionID)
	searchResult, err := db.es.Client.Search().
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

	var as []piazza.Alert
	for _, hit := range searchResult.Hits.Hits {
		var a piazza.Alert
		err := json.Unmarshal(*hit.Source, &a)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return as, nil
}

func (db *AlertDB) getAll() (map[string]piazza.Alert, error) {
	searchResult, err := db.es.Client.Search().
		Index(db.index).
		Query(elastic.NewMatchAllQuery()).
		Sort("id", true).
		Do()
	if err != nil {
		return nil, err
	}

	m := make(map[string]piazza.Alert)

	if searchResult.Hits != nil {
		for _, hit := range searchResult.Hits.Hits {
			var t piazza.Alert
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				return nil, err
			}
			m[t.ID] = t
		}
	}

	return m, nil
}

func (db *AlertDB) checkConditions(e piazza.Event, conditionDB *ConditionDB) error {
	all, err := conditionDB.getAll()
	if err != nil {
		return nil
	}
	for _, cond := range all {
		//log.Printf("e%s.%s ==? c%s.%s", e.ID, e.Type, cond.ID, cond.Type)
		if cond.Type == e.Type {
			a := newAlert(cond.ID, e.ID)
			db.write(&a)
			pzService.Log(piazza.SeverityInfo, fmt.Sprintf("HIT! event %s has triggered condition %s: alert %s", e.ID, cond.ID, a.ID))
		}
	}
	return nil
}
