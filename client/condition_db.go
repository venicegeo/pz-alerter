package client

import (
	"encoding/json"
	"github.com/venicegeo/pz-gocommon"
)

type ConditionDB struct {
	es    *piazza.ElasticSearchService
	index string
}

func NewConditionDB(es *piazza.ElasticSearchService, index string) (*ConditionDB, error) {
	db := new(ConditionDB)
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

func (db *ConditionDB) Write(condition *Condition) error {

	id := NewConditionIdent()
	condition.ID = id

	_, err := db.es.PostData(db.index, "condition", condition.ID.String(), condition)
	if err != nil {
		return err
	}

	err = db.es.FlushIndex(db.index)
	if err != nil {
		return err
	}

	return nil
}

func (db *ConditionDB) Update(condition *Condition) bool {
	/**	_, ok := db.data[condition.ID]
	if ok {
		db.data[condition.ID] = *condition
		return true
	}
	**/
	return false
}

func (db *ConditionDB) ReadByID(id Ident) (*Condition, error) {
	getResult, err := db.es.GetById(db.index, id.String())
	if err != nil {
		return nil, err
	}
	if !getResult.Found {
		return nil, nil
	}

	var tmp Condition
	src := getResult.Source
	err = json.Unmarshal(*src, &tmp)
	if err != nil {
		return nil, err
	}
	return &tmp, nil
}

func (db *ConditionDB) DeleteByID(id string) (bool, error) {
	res, err := db.es.DeleteById(db.index, "condition", id)
	if err != nil {
		return false, err
	}

	err = db.es.FlushIndex(db.index)
	if err != nil {
		return false, err
	}

	return res.Found, nil
}

func (db *ConditionDB) GetAll() (map[Ident]Condition, error) {

	searchResult, err := db.es.SearchByMatchAll(db.index)
	if err != nil {
		return nil, err
	}

	m := make(map[Ident]Condition)

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
