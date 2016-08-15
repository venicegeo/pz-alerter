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
	"strings"

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

func (db *EventDB) PostData(typ string, obj interface{}, id piazza.Ident) (piazza.Ident, error) {
	event, ok := obj.(*Event)
	if !ok {
		return piazza.NoIdent, LoggedError("EventDB.PostData failed: was not given an event")
	}
	{ //ARRAY CHECK
		eventTypeJson := db.service.GetEventType(event.EventTypeId)
		eventTypeObj := eventTypeJson.Data
		eventType, ok := eventTypeObj.(*EventType)
		if !ok {
			return piazza.NoIdent, LoggedError("EvenDB.PostData failed: unable to obtain specified eventtype")
		}
		eventTypeMappingKeys,
			eventTypeMappingValues, err := piazza.GetVariablesFromStructInterface(eventType.Mapping)
		if err != nil {
			return piazza.NoIdent, LoggedError("EventDB.PostData failed: %s", err)
		}
		eventDataKeys, eventDataValues, err := piazza.GetVariablesFromStructInterface(event.Data)
		if err != nil {
			return piazza.NoIdent, LoggedError("EventDB.PostData failed: %s", err)
		}
		var evKey, eKey, evValue, eValue string
		for i := 0; i < len(eventTypeMappingKeys); i++ {
			evKey = eventTypeMappingKeys[i]
			evValue = eventTypeMappingValues[i]
			for j := 0; j < len(eventDataKeys); j++ {
				eKey = eventDataKeys[j]
				eValue = eventDataValues[j]
				if evKey != eKey {
					continue
				}
				if strings.HasPrefix(evValue, "\"") && strings.HasSuffix(evValue, "\"") {
					evValue = evValue[1 : len(evValue)-1]
				}
				if !elasticsearch.IsValidArrayTypeMapping(evValue) {
					if piazza.ValueIsValidArray(eValue) {
						return piazza.NoIdent, LoggedError("EventDB.PostData failed: an array was passed into a non-array field: %s", eKey)
					}
				} else {
					if !piazza.ValueIsValidArray(eValue) {
						return piazza.NoIdent, LoggedError("EventDB.PostData failed: a non-array was passed into an array field: %s", eKey)
					}
				}
			}
		}
	}
	indexResult, err := db.Esi.PostData(typ, id.String(), obj)
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
	var err error

	exists := true
	if mapping != "" {
		exists, err = db.Esi.TypeExists(mapping)
		if err != nil {
			return events, 0, err
		}
	}
	if !exists {
		return nil, 0, fmt.Errorf("Type %s does not exist", mapping)
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
		ok, err := db.Esi.ItemExists(typ, id.String())
		if err != nil {
			return "", err
		}
		if ok {
			mapping = typ
			break
		}
	}

	return mapping, nil
}

// NameExists checks if an EventType name exists.
// This is easier to check in EventDB, as the mappings use the EventType.Name.
func (db *EventDB) NameExists(name string) (bool, error) {
	return db.Esi.TypeExists(name)
}

func (db *EventDB) GetOne(mapping string, id piazza.Ident) (*Event, bool, error) {
	getResult, err := db.Esi.GetByID(mapping, id.String())
	if err != nil {
		return nil,getResult.Found LoggedError("EventDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, LoggedError("EventDB.GetOne failed: no getResult")
	}

	src := getResult.Source
	var event Event
	err = json.Unmarshal(*src, &event)
	if err != nil {
		return nil,getResult.Found, err
	}

	return &event,getResult.Found, nil
}

func (db *EventDB) DeleteByID(mapping string, id piazza.Ident) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(mapping, string(id))
	if err != nil {
		return deleteResult.Found, LoggedError("EventDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, LoggedError("EventDB.DeleteById failed: no deleteResult")
	}

	return deleteResult.Found, nil
}

func (db *EventDB) AddMapping(name string, mapping interface{}) error {
	mapMapping, ok := mapping.(map[string]interface{})
	if !ok {
		return LoggedError("EventDB.AddMapping failed: mapping is not of type map[string]interface{}")
	}

	jsn, err := ConstructEventMappingSchema(name, mapMapping)
	if err != nil {
		return LoggedError("EventDB.AddMapping failed: %s", err)
	}

	err = db.Esi.SetMapping(name, jsn)
	if err != nil {
		return LoggedError("EventDB.AddMapping SetMapping failed: %s", err)
	}

	return nil
}

// ConstructEventMappingSchema takes a map of parameter names to datatypes and
// returns the corresponding ES DSL for it.
func ConstructEventMappingSchema(name string, mapping map[string]interface{}) (piazza.JsonString, error) {
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
	esdsl, err := buildMapping(mapping)
	if err != nil {
		return piazza.JsonString(""), err
	}
	strDsl, err := piazza.StructInterfaceToString(esdsl)
	if err != nil {
		return piazza.JsonString(""), err
	}
	json := fmt.Sprintf(template, name, strDsl)
	return piazza.JsonString(json), nil
}

func buildMapping(input map[string]interface{}) (map[string]interface{}, error) {
	return visitNode(input)
}

func visitNode(inputObj map[string]interface{}) (map[string]interface{}, error) {
	outputObj := map[string]interface{}{}

	for k, v := range inputObj {
		switch t := v.(type) {

		case string:
			tree, err := handleLeaf(k, v)
			if err != nil {
				return nil, err
			}
			outputObj[k] = tree

		case map[string]interface{}:
			tree, err := handleNonleaf(k, v)
			if err != nil {
				return nil, err
			}
			outputObj[k] = tree

		default:
			return nil, fmt.Errorf("unexpected type %T\n", t)
		}
	}
	return outputObj, nil
}

func handleNonleaf(k string, v interface{}) (map[string]interface{}, error) {
	subtree, err := visitNode(v.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	wrapperTree := map[string]interface{}{}
	wrapperTree["dynamic"] = "strict"
	wrapperTree["properties"] = subtree

	return wrapperTree, err
}

func handleLeaf(k string, v interface{}) (map[string]interface{}, error) {
	if !elasticsearch.IsValidMappingType(v) {
		return nil, LoggedError(" %s was not recognized as a valid mapping type", v)
	}
	if elasticsearch.IsValidArrayTypeMapping(v) {
		v = v.(string)[1 : len(v.(string))-1]
	}
	tree := map[string]interface{}{}
	tree["type"] = v
	return tree, nil
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
