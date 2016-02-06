package client

import (
	"encoding/json"
	"gopkg.in/olivere/elastic.v3"
	piazza "github.com/venicegeo/pz-gocommon"
	"log"
)

//---------------------------------------------------------------------------

type AlertDB struct {
	es *piazza.ElasticSearchService
	index  string
}

func NewAlertDB(es *piazza.ElasticSearchService, index string) (*AlertDB, error) {
	db := new(AlertDB)
	db.es = es
	db.index = index

	err := es.MakeIndex(index)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *AlertDB) Write(alert *Alert) error {

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

func (db *AlertDB) GetByConditionID(conditionID string) ([]Alert, error) {
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

func (db *AlertDB) GetAll() (map[string]Alert, error) {
	searchResult, err := db.es.Client.Search().
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

func (db *AlertDB) CheckConditions(e Event, conditionDB *ConditionDB) error {
	all, err := conditionDB.GetAll()
	if err != nil {
		return nil
	}
	for _, cond := range all {
		//log.Printf("e%s.%s ==? c%s.%s", e.ID, e.Type, cond.ID, cond.Type)
		if cond.Type == e.Type {
			a := NewAlert(cond.ID, e.ID)
			db.Write(&a)
			//pzService.Log(piazza.SeverityInfo, fmt.Sprintf("HIT! event %s has triggered condition %s: alert %s", e.ID, cond.ID, a.ID))
			log.Printf("INFO: Hit! event %s has triggered condition %s: alert %s", e.ID, cond.ID, a.ID)
		}
	}
	return nil
}
