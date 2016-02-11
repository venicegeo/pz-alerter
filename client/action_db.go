package client

import (
	"encoding/json"
	"github.com/venicegeo/pz-gocommon"
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

	err := es.CreateIndex(index)
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
