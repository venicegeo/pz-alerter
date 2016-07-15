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

type EventDB struct {
	*ResourceDB
}

func NewEventDB(service *WorkflowService, esi elasticsearch.IIndex) (*EventDB, error) {
	rdb, err := NewResourceDB(service, esi, EventIndexSettings)
	if err != nil {
		return nil, err
	}
	erdb := EventDB{ResourceDB: rdb}
	return &erdb, nil
}

func (db *EventDB) PostData(mapping string, obj interface{}, id piazza.Ident) (piazza.Ident, error) {
	indexResult, err := db.Esi.PostData(mapping, id.String(), obj)
	if err != nil {
		return piazza.NoIdent, LoggedError("EventDB.PostData failed: %s", err)
	}
	if !indexResult.Created {
		return piazza.NoIdent, LoggedError("EventDB.PostData failed: not created")
	}

	return id, nil
}

func (db *EventDB) GetAll(mapping string, format *piazza.JsonPagination) (*[]Event, int64, error) {
	var events []Event
	var count = int64(-1)

	exists := true
	if mapping != "" {
		exists = db.Esi.TypeExists(mapping)
	}
	if !exists {
		return nil, count, LoggedError("Type %s does not exist", mapping)
	}

	searchResult, err := db.Esi.FilterByMatchAll(mapping, format)
	if err != nil {
		return nil, count, LoggedError("EventDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, count, LoggedError("EventDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		count = searchResult.NumberMatched()
		for _, hit := range *searchResult.GetHits() {
			var event Event
			err := json.Unmarshal(*hit.Source, &event)
			if err != nil {
				return nil, count, err
			}
			events = append(events, event)
		}
	}

	return &events, count, nil
}

func (db *EventDB) lookupEventTypeNameByEventID(id piazza.Ident) (string, error) {
	var mapping string = ""

	types, err := db.Esi.GetTypes()
	if err != nil {
		return "", err
	}
	for _, typ := range types {
		if db.Esi.ItemExists(typ, id.String()) {
			mapping = typ
			break
		}
	}

	return mapping, nil
}

// Function to check if an EventType name exists
// This is easier to check in EventDB, as the mappings use the EventType.Name
func (db *EventDB) NameExists(name string) bool {
	return db.Esi.TypeExists(name)
}

func (db *EventDB) GetOne(mapping string, id piazza.Ident) (*Event, error) {
	getResult, err := db.Esi.GetByID(mapping, id.String())
	if err != nil {
		return nil, LoggedError("EventDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, LoggedError("EventDB.GetOne failed: no getResult")
	}

	if !getResult.Found {
		return nil, nil
	}

	src := getResult.Source
	var event Event
	err = json.Unmarshal(*src, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (db *EventDB) DeleteByID(mapping string, id piazza.Ident) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(mapping, string(id))
	if err != nil {
		return false, LoggedError("EventDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, LoggedError("EventDB.DeleteById failed: no deleteResult")
	}

	return deleteResult.Found, nil
}

func (db *EventDB) AddMapping(name string, mapping map[string]elasticsearch.MappingElementTypeName) error {
	jsn, err := elasticsearch.ConstructMappingSchema(name, mapping)
	if err != nil {
		return LoggedError("EventDB.AddMapping failed: %s", err)
	}

	err = db.Esi.SetMapping(name, jsn)
	if err != nil {
		return LoggedError("EventDB.AddMapping SetMapping failed: %s", err)
	}

	return nil
}

func (db *EventDB) PercolateEventData(eventType string, data map[string]interface{}, id piazza.Ident) (*[]piazza.Ident, error) {
	percolateResponse, err := db.Esi.AddPercolationDocument(eventType, data)

	if err != nil {
		return nil, LoggedError("EventDB.PercolateEventData failed: %s", err)
	}
	if percolateResponse == nil {
		return nil, LoggedError("EventDB.PercolateEventData failed: no percolateResult")
	}

	// add the triggers to the alert queue
	ids := make([]piazza.Ident, len(percolateResponse.Matches))
	for i, v := range percolateResponse.Matches {
		ids[i] = piazza.Ident(v.Id)
	}

	return &ids, nil
}
