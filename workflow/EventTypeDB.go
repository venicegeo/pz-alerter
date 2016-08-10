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

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

type EventTypeDB struct {
	*ResourceDB
	mapping string
}

func NewEventTypeDB(service *WorkflowService, esi elasticsearch.IIndex) (*EventTypeDB, error) {

	rdb, err := NewResourceDB(service, esi, EventTypeIndexSettings)
	if err != nil {
		return nil, err
	}
	etrdb := EventTypeDB{ResourceDB: rdb, mapping: EventTypeDBMapping}
	return &etrdb, nil
}

func (db *EventTypeDB) PostData(obj interface{}, id piazza.Ident) (piazza.Ident, error) {

	indexResult, err := db.Esi.PostData(db.mapping, id.String(), obj)
	if err != nil {
		return piazza.NoIdent, LoggedError("EventTypeDB.PostData failed: %s", err)
	}
	if !indexResult.Created {
		return piazza.NoIdent, LoggedError("EventTypeDB.PostData failed: not created")
	}

	return id, nil
}

func (db *EventTypeDB) GetAll(format *piazza.JsonPagination) ([]EventType, int64, error) {
	eventTypes := []EventType{}

	exists := db.Esi.TypeExists(db.mapping)
	if !exists {
		return eventTypes, 0, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(db.mapping, format)
	if err != nil {
		return nil, 0, LoggedError("EventTypeDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("EventTypeDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var eventType EventType
			err := json.Unmarshal(*hit.Source, &eventType)
			if err != nil {
				return nil, 0, err
			}
			eventTypes = append(eventTypes, eventType)
		}
	}

	return eventTypes, searchResult.TotalHits(), nil
}

func (db *EventTypeDB) GetOne(id piazza.Ident) (*EventType, bool, error) {

	getResult, err := db.Esi.GetByID(db.mapping, id.String())
	if err != nil {
		return nil, getResult.Found, LoggedError("EventTypeDB.GetOne failed: %s", err.Error())
	}
	if getResult == nil {
		return nil, true, LoggedError("EventTypeDB.GetOne failed: no getResult")
	}

	src := getResult.Source
	var eventType EventType
	err = json.Unmarshal(*src, &eventType)
	if err != nil {
		return nil, getResult.Found, err
	}

	return &eventType, getResult.Found, nil
}

func (db *EventTypeDB) GetIDByName(name string) (*piazza.Ident, bool, error) {

	getResult, err := db.Esi.FilterByTermQuery(db.mapping, "name", name)
	if err != nil {
		return nil, getResult.Found, LoggedError("EventTypeDB.GetIDByName failed: %s", err.Error())
	}
	if getResult == nil {
		return nil, true, LoggedError("EventTypeDB.GetIDByName failed: no getResult")
	}

	// This should not happen once we have 1 to 1 mappings of EventTypes to names
	if getResult.TotalHits() > 1 {
		return nil, true, LoggedError("EventTypeDB.GetIDByName failed: matched more than one EventType!")
	}

	if getResult.TotalHits() == 0 {
		return nil, false, LoggedError("EventTypeDB.GetIDByName failed: %s matched no existing EventTypeIds", name)
	}

	src := getResult.GetHit(0).Source
	var eventType EventType
	err = json.Unmarshal(*src, &eventType)
	if err != nil {
		return nil, getResult.Found, err
	}

	return &eventType.EventTypeId, getResult.Found, nil
}

func (db *EventTypeDB) DeleteByID(id piazza.Ident) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(db.mapping, string(id))
	if err != nil {
		return deleteResult.Found, LoggedError("EventTypeDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, LoggedError("EventTypeDB.DeleteById failed: no deleteResult")
	}

	return deleteResult.Found, nil
}
