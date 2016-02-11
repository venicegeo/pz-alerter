package client

import (
	"encoding/json"
	"github.com/venicegeo/pz-gocommon"
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

func (db *AlertDB) Write(alert *Alert) error {

	_, err := db.es.PostData(db.index, "alert", alert.ID.String(), alert)
	if err != nil {
		return err
	}

	err = db.es.FlushIndex(db.index)
	if err != nil {
		return err
	}

	return nil
}

func (db *AlertDB) GetByConditionID(conditionID string) ([]Alert, error) {
	searchResult, err := db.es.SearchByTermQuery(db.index, "condition_id", conditionID)
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

func (db *AlertDB) GetAll() (map[Ident]Alert, error) {
	searchResult, err := db.es.SearchByMatchAll(db.index)
	if err != nil {
		return nil, err
	}

	m := make(map[Ident]Alert)

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
			a := NewAlert(NewAlertIdent())
			db.Write(&a)
			//pzService.Log(piazza.SeverityInfo, fmt.Sprintf("HIT! event %s has triggered condition %s: alert %s", e.ID, cond.ID, a.ID))
			log.Printf("INFO: Hit! event %s has triggered condition %s: alert %s", e.ID, cond.ID, a.ID)
		}
	}
	return nil
}


func (db *AlertDB) DeleteByID(id string) (bool, error) {
	res, err := db.es.DeleteById(db.index, "alert", id)
	if err != nil {
		return false, err
	}

	err = db.es.FlushIndex(db.index)
	if err != nil {
		return false, err
	}

	return res.Found, nil
}
