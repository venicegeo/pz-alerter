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

package workflow

import (
	"encoding/json"
	"fmt"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

type TriggerDB struct {
	*ResourceDB
	mapping string
}

func NewTriggerDB(service *WorkflowService, esi elasticsearch.IIndex) (*TriggerDB, error) {

	rdb, err := NewResourceDB(service, esi, TriggerIndexSettings)
	if err != nil {
		return nil, err
	}
	ardb := TriggerDB{ResourceDB: rdb, mapping: TriggerDBMapping}
	return &ardb, nil
}

func (db *TriggerDB) PostTrigger(trigger *Trigger, id piazza.Ident) (piazza.Ident, error) {

	ifaceObj := trigger.Condition.Query
	//log.Printf("Query: %v", ifaceObj)
	body, err := json.Marshal(ifaceObj)
	if err != nil {
		return piazza.NoIdent, err
	}

	json := string(body)
	//log.Printf("Current json: %s", json)
	// Remove trailing }
	json = json[:len(json)-1]
	json += ",\"type\":["
	// Add the types that the percolation query can match
	for _, id := range trigger.Condition.EventTypeIds {
		json += fmt.Sprintf("\"%s\",", id)
	}
	json = json[:len(json)-1]
	// Add back trailing } and ] to close array
	json += "]}"

	//log.Printf("Posting percolation query: %s", body)
	indexResult, err := db.service.eventDB.Esi.AddPercolationQuery(string(trigger.TriggerId), piazza.JsonString(body))
	if err != nil {
		return piazza.NoIdent, LoggedError("TriggerDB.PostData addpercquery failed: %s", err)
	}
	if indexResult == nil {
		return piazza.NoIdent, LoggedError("TriggerDB.PostData addpercquery failed: no indexResult")
	}
	if !indexResult.Created {
		return piazza.NoIdent, LoggedError("TriggerDB.PostData addpercquery failed: not created")
	}

	//log.Printf("percolation query added: ID: %s, Type: %s, Index: %s", indexResult.Id, indexResult.Type, indexResult.Index)
	//log.Printf("percolation id: %s", indexResult.Id)
	trigger.PercolationId = piazza.Ident(indexResult.Id)

	indexResult2, err := db.Esi.PostData(db.mapping, id.String(), trigger)
	if err != nil {
		db.service.eventDB.Esi.DeletePercolationQuery(string(trigger.TriggerId))
		return piazza.NoIdent, LoggedError("TriggerDB.PostData failed: %s", err)
	}
	if !indexResult2.Created {
		db.service.eventDB.Esi.DeletePercolationQuery(string(trigger.TriggerId))
		return piazza.NoIdent, LoggedError("TriggerDB.PostData failed: not created")
	}

	return id, nil
}

func (db *TriggerDB) GetAll(format *piazza.JsonPagination) ([]Trigger, int64, error) {
	triggers := []Trigger{}

	exists := db.Esi.TypeExists(db.mapping)
	if !exists {
		return triggers, 0, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(db.mapping, format)
	if err != nil {
		return nil, 0, LoggedError("TriggerDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("TriggerDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {

		for _, hit := range *searchResult.GetHits() {
			var trigger Trigger
			err := json.Unmarshal(*hit.Source, &trigger)
			if err != nil {
				return nil, 0, err
			}
			triggers = append(triggers, trigger)
		}
	}
	return triggers, searchResult.TotalHits(), nil
}

func (db *TriggerDB) GetOne(id piazza.Ident) (*Trigger, error) {

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

func (db *TriggerDB) DeleteTrigger(id piazza.Ident) (bool, error) {

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

	deleteResult2, err := db.service.eventDB.Esi.DeletePercolationQuery(string(trigger.PercolationId))
	if err != nil {
		return false, LoggedError("TriggerDB.DeleteById percquery failed: %s", err)
	}
	if deleteResult2 == nil {
		return false, LoggedError("TriggerDB.DeleteById percquery failed: no deleteResult")
	}

	return deleteResult2.Found, nil
}
