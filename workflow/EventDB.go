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

func (db *EventDB) GetAll(mapping string, format *piazza.JsonPagination) ([]Event, int64, error) {
	events := []Event{}

	exists := true
	if mapping != "" {
		exists = db.Esi.TypeExists(mapping)
	}
	if !exists {
		return nil, 0, LoggedError("Type %s does not exist", mapping)
	}

	searchResult, err := db.Esi.FilterByMatchAll(mapping, format)
	if err != nil {
		return nil, 0, LoggedError("EventDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("EventDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var event Event
			err := json.Unmarshal(*hit.Source, &event)
			if err != nil {
				return nil, 0, err
			}
			events = append(events, event)
		}
	}

	return events, searchResult.TotalHits(), nil
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

// NameExists checks if an EventType name exists.
// This is easier to check in EventDB, as the mappings use the EventType.Name.
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

func (db *EventDB) AddMapping(name string, mapping interface{}) error {
	fmt.Printf("Creating mapping in EventDB: %s\n%s\n", name, mapping)
	jsn, err := ConstructEventMappingSchema(name, MappingInterfaceToString(mapping))
	if err != nil {
		return LoggedError("EventDB.AddMapping failed: %s", err)
	}
	fmt.Printf("%s\n", jsn)

	err = db.Esi.SetMapping(name, jsn)
	if err != nil {
		return LoggedError("EventDB.AddMapping SetMapping failed: %s", err)
	}

	return nil
}

// ConstructEventMappingSchema takes a map of parameter names to datatypes and
// returns the corresponding ES DSL for it.
func ConstructEventMappingSchema(name string, mapping string) (piazza.JsonString, error) {
	fmt.Printf("ConstructEventMappingSchema recieved mapping:\n%s\n", mapping)
	const template string = `{
			"%s":{
				"dynamic": false,
				"properties":{
					"data": {
						"dynamic": "strict",
						"properties": %s
					}
				}
			}
		}`

	/*
		stuff := make([]string, len(items))
		i := 0
		for k, v := range items {
			stuff[i] = fmt.Sprintf(`"%s": {"type":"%s"}`, k, v)
			i++
		}

		json := fmt.Sprintf(template, name, strings.Join(stuff, ","))
	*/
	json := fmt.Sprintf(template, name, mapping)
	return piazza.JsonString(json), nil
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
