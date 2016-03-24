// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"encoding/json"
	"errors"

	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
)

type TriggerDB struct {
	*ResourceDB
}

func NewTriggerDB(es *elasticsearch.Client, index string) (*TriggerDB, error) {

	esi := elasticsearch.NewIndex(es, index)

	rdb, err := NewResourceDB(es, esi)
	if err != nil {
		return nil, err
	}
	ardb := TriggerDB{ResourceDB: rdb}
	return &ardb, nil
}

func (db *TriggerDB) PostTrigger(mapping string, trigger *Trigger, id Ident, eventDB *EventDB) (Ident, error) {

	ifaceObj := trigger.Condition.Query
	body, err := json.Marshal(ifaceObj)
	if err != nil {
		return NoIdent, err
	}

	indexResult, err := eventDB.Esi.AddPercolationQuery(string(trigger.ID), piazza.JsonString(body))
	if err != nil {
		return NoIdent, err
	}

	trigger.PercolationID = Ident(indexResult.Id)

	_, err = db.Esi.PostData(mapping, id.String(), trigger)
	if err != nil {
		return NoIdent, err
	}

	err = db.Esi.Flush()
	if err != nil {
		return NoIdent, err
	}

	return id, nil
}

func (db *TriggerDB) DeleteTrigger(mapping string, id Ident, eventDB *EventDB) (bool, error) {

	trigger, err := db.GetOne(mapping, id)
	if err != nil {
		return false, err
	}
	if trigger == nil {
		return false, nil
	}

	res, err := db.Esi.DeleteByID(mapping, string(id))
	if err != nil {
		return false, err
	}

	err = db.Esi.Flush()
	if err != nil {
		return false, err
	}

	deleteResult, err := eventDB.Esi.DeletePercolationQuery(string(trigger.PercolationID))
	if !deleteResult.Found {
		return false, errors.New("unable to delete percolation")
	}

	err = db.Esi.Flush()
	if err != nil {
		return false, err
	}

	return res.Found, nil
}

func (db *TriggerDB) GetAll(mapping string) (*[]Trigger, error) {
	searchResult, err := db.Esi.FilterByMatchAll(mapping)
	if err != nil {
		return nil, err
	}

	var triggers []Trigger

	if searchResult.Hits != nil {

		for _, hit := range searchResult.Hits.Hits {
			var trigger Trigger
			err := json.Unmarshal(*hit.Source, &trigger)
			if err != nil {
				return nil, err
			}
			triggers = append(triggers, trigger)
		}
	}
	return &triggers, nil
}

func (db *TriggerDB) GetOne(mapping string, id Ident) (*Trigger, error) {

	getResult, err := db.Esi.GetByID(mapping, id.String())
	if err != nil {
		return nil, err
	}

	if !getResult.Found {
		return nil, nil
	}

	src := getResult.Source
	var obj Trigger
	err = json.Unmarshal(*src, &obj)
	if err != nil {
		return nil, err
	}

	return &obj, nil
}
