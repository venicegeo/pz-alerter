package client

import (
	"encoding/json"
	"github.com/venicegeo/pz-gocommon"
	"log"
)

//---------------------------------------------------------------------------

type ActionDB struct {
	es *piazza.ElasticSearchService
	index  string
}

func NewActionDB(es *piazza.ElasticSearchService, index string) (*ActionDB, error) {
	db := new(ActionDB)
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

func (db *ActionDB) Write(action *Action) error {

	_, err := db.es.PostData(db.index, "action", action.ID.String(), action)
	if err != nil {
		return err
	}

	err = db.es.FlushIndex(db.index)
	if err != nil {
		return err
	}

	return nil
}


func (db *ActionDB) GetAll() (map[Ident]Action, error) {
	searchResult, err := db.es.SearchByMatchAll(db.index)
	if err != nil {
		return nil, err
	}

	m := make(map[Ident]Action)

	if searchResult.Hits != nil {
		for _, hit := range searchResult.Hits.Hits {
			var t Action
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				return nil, err
			}
			m[t.ID] = t
		}
	}

	return m, nil
}

func (db *ActionDB) GetByID(id Ident) (*Action, error) {

	getResult, err := db.es.GetById(db.index, id.String())
	if err != nil {
		return nil, err
	}
	var tmp Action
	src := getResult.Source
	err = json.Unmarshal(*src, &tmp)
	if err != nil {
		return nil, err
	}
	return &tmp, nil
}

func (db *ActionDB) DeleteByID(id string) (bool, error) {
	res, err := db.es.DeleteById(db.index, "action", id)
	if err != nil {
		return false, err
	}

	err = db.es.FlushIndex(db.index)
	if err != nil {
		return false, err
	}

	return res.Found, nil
}

func (db *ActionDB) CheckActions(event Event, conditionDB *ConditionDB, alertDB *AlertDB) error {
	actions, err := db.GetAll()
	if err != nil {
		return nil
	}

	for _, action := range actions {
		//log.Printf("e%s.%s ==? c%s.%s", e.ID, e.Type, cond.ID, cond.Type)
		match := true
		for _, condId := range(action.Conditions) {
			cond, err := conditionDB.ReadByID(condId)
			if err != nil {
				return nil
			}

			if cond.Type != event.Type {
				match = false
				break
			}
		}
		if match {
			alert := NewAlert(NewAlertIdent())
			alert.ActionId = action.ID
			alert.EventId = event.ID
			alertDB.Write(&alert)
			log.Printf("INFO: Hit! event %s has triggered action %s, causing alert %s", event.ID, action.ID, alert.ID)
		}
	}
	return nil
}
