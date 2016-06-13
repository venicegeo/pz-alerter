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
	"log"

	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
)

type TriggerDB struct {
	*ResourceDB
	mapping string
}

func NewTriggerDB(server *Server, esi elasticsearch.IIndex) (*TriggerDB, error) {

	rdb, err := NewResourceDB(server, esi)
	if err != nil {
		return nil, err
	}
	ardb := TriggerDB{ResourceDB: rdb, mapping: "Trigger"}
	return &ardb, nil
}

func (db *TriggerDB) PostTrigger(trigger *Trigger, id Ident) (Ident, error) {

	ifaceObj := trigger.Condition.Query
	body, err := json.Marshal(ifaceObj)
	if err != nil {
		return NoIdent, err
	}

	log.Printf("Posting percolation query: %s", string(body))
	indexResult, err := db.server.eventDB.Esi.AddPercolationQuery(string(trigger.ID), piazza.JsonString(body))
	if err != nil {
		return NoIdent, LoggedError("TriggerDB.PostData addpercquery failed: %s", err)
	}
	if indexResult == nil {
		return NoIdent, LoggedError("TriggerDB.PostData addpercquery failed: no indexResult")
	}
	if !indexResult.Created {
		return NoIdent, LoggedError("TriggerDB.PostData addpercquery failed: not created")
	}

	log.Printf("percolation indexResult: %v", indexResult)
	log.Printf("percolation id: %s", indexResult.Id)
	trigger.PercolationID = Ident(indexResult.Id)

	indexResult2, err := db.Esi.PostData(db.mapping, id.String(), trigger)
	if err != nil {
		return NoIdent, LoggedError("TriggerDB.PostData failed: %s", err)
	}
	if !indexResult2.Created {
		return NoIdent, LoggedError("TriggerDB.PostData failed: not created")
	}

	return id, nil
}

func (db *TriggerDB) GetAll(format elasticsearch.QueryFormat) (*[]Trigger, error) {
	triggers := []Trigger{}

	exists := db.Esi.TypeExists(db.mapping)
	if !exists {
		return &triggers, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(db.mapping, format)
	if err != nil {
		return nil, LoggedError("TriggerDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, LoggedError("TriggerDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {

		for _, hit := range *searchResult.GetHits() {
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

func (db *TriggerDB) GetAllWithCount(format elasticsearch.QueryFormat) (*[]Trigger, int64, error) {
	var triggers []Trigger
	var count = int64(-1)

	exists := db.Esi.TypeExists(db.mapping)
	if !exists {
		return &triggers, count, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(db.mapping, format)
	if err != nil {
		return nil, count, LoggedError("TriggerDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, count, LoggedError("TriggerDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		count = searchResult.NumberMatched()

		for _, hit := range *searchResult.GetHits() {
			var trigger Trigger
			err := json.Unmarshal(*hit.Source, &trigger)
			if err != nil {
				return nil, count, err
			}
			triggers = append(triggers, trigger)
		}
	}
	return &triggers, count, nil
}

func (db *TriggerDB) GetOne(id Ident) (*Trigger, error) {

	getResult, err := db.Esi.GetByID(db.mapping, id.String())
	if err != nil {
		return nil, LoggedError("TriggerDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, LoggedError("TriggerDB.GetOne failed: no getResult")
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

func (db *TriggerDB) DeleteTrigger(id Ident) (bool, error) {

	trigger, err := db.GetOne(id)
	if err != nil {
		return false, err
	}
	if trigger == nil {
		return false, nil
	}

	deleteResult, err := db.Esi.DeleteByID(db.mapping, string(id))
	if err != nil {
		return false, LoggedError("TriggerDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, LoggedError("TriggerDB.DeleteById failed: no deleteResult")
	}
	if !deleteResult.Found {
		return false, nil
	}

	deleteResult2, err := db.server.eventDB.Esi.DeletePercolationQuery(string(trigger.PercolationID))
	if err != nil {
		return false, LoggedError("TriggerDB.DeleteById percquery failed: %s", err)
	}
	if deleteResult2 == nil {
		return false, LoggedError("TriggerDB.DeleteById percquery failed: no deleteResult")
	}

	return deleteResult2.Found, nil
}
