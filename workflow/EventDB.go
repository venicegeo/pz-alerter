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
	var event Event
	ok1 := false
	event, ok1 = obj.(Event)
	if !ok1 {
		temp, ok2 := obj.(*Event)
		if !ok2 {
			return piazza.NoIdent, LoggedError("EventDB.PostData failed: was not given an Event")
		}
		event = *temp
	}
	{ //Check array types, elasticsearch doesn't know the difference
		eventTypeJson := db.service.GetEventType(event.EventTypeId)
		eventTypeObj := eventTypeJson.Data
		eventType, ok := eventTypeObj.(*EventType)
		if !ok {
			return piazza.NoIdent, LoggedError("EventDB.PostData failed: unable to obtain specified eventtype")
		}
		eventTypeMappingVars, err := piazza.GetVarsFromStruct(eventType.Mapping)
		if err != nil {
			return piazza.NoIdent, LoggedError("EventDB.PostData failed: %s", err)
		}
		eventDataVars, err := piazza.GetVarsFromStruct(event.Data)
		if err != nil {
			return piazza.NoIdent, LoggedError("EventDB.PostData failed: %s", err)
		}
		notFound := []string{}
		for k, v := range eventTypeMappingVars {
			found := false
			for k2, v2 := range eventDataVars {
				if k2 == k {
					found = true
					if !elasticsearch.IsValidArrayTypeMapping(v) {
						if piazza.ValueIsValidArray(v2) {
							return piazza.NoIdent, LoggedError("EventDB.PostData failed: an array was passed into the non-array field %s", k)
						}
					} else {
						if !piazza.ValueIsValidArray(v2) {
							return piazza.NoIdent, LoggedError("EventDB.PostData failed: a non-array was pasted into the array field %s", k)
						}
					}
					break
				}
			}
			if !found {
				notFound = append(notFound, k)
			}
		}
		if len(notFound) > 0 {
			return piazza.NoIdent, LoggedError("EventDB.PostData failed: the variables %s were specified in the EventType but were not found in the Event", notFound)
		}
	}

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
	if mapping == "" {
		return "", LoggedError("EventDB.lookupEventTypeNameByEventID failed: [Item %s in index events does not exist]", id.String())
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
		return nil, false, LoggedError("EventDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, true, LoggedError("EventDB.GetOne failed: no getResult")
	}

	src := getResult.Source
	var event Event
	err = json.Unmarshal(*src, &event)
	if err != nil {
		return nil, getResult.Found, err
	}

	return &event, getResult.Found, nil
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

func (db *EventDB) AddMapping(name string, mapping map[string]interface{}) error {
	jsn, err := ConstructEventMappingSchema(name, mapping)
	if err != nil {
		return LoggedError("EventDB.AddMapping failed: %s", err)
	}
	err = db.Esi.SetMapping(name, jsn)
	if err != nil {
		return LoggedError("EventDB.AddMapping SetMapping failed: %s", err)
	}

	return nil
}

func ConstructEventMappingSchema(name string, mapping map[string]interface{}) (piazza.JsonString, error) {
	const template string = `{
		"%s":{
			"properties":{
				"data":{
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
	jsn := fmt.Sprintf(template, name, strDsl)
	return piazza.JsonString(jsn), nil
}

func buildMapping(input map[string]interface{}) (map[string]interface{}, error) {
	return visitNode(input)
}
func visitNode(inputObj map[string]interface{}) (map[string]interface{}, error) {
	outputObj := map[string]interface{}{}
	for k, v := range inputObj {
		switch t := v.(type) {
		case string:
			tree, err := visitLeaf(k, v)
			if err != nil {
				return nil, err
			}
			outputObj[k] = tree
		case map[string]interface{}:
			tree, err := visitTree(k, v.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			outputObj[k] = tree
		default:
			return nil, LoggedError("EventDB.ConstructEventMappingSchema failed: unexpected type %T", t)
		}
	}
	return outputObj, nil
}
func visitTree(k string, v map[string]interface{}) (map[string]interface{}, error) {
	subtree, err := visitNode(v)
	if err != nil {
		return nil, err
	}
	wrapperTree := map[string]interface{}{}
	wrapperTree["dynamic"] = "strict"
	wrapperTree["properties"] = subtree
	return wrapperTree, nil
}
func visitLeaf(k string, v interface{}) (map[string]interface{}, error) {
	if !elasticsearch.IsValidMappingType(v) {
		return nil, LoggedError("EventDB.ConstructEventMappingSchema failed: %s was not recognized as a valid mapping type")
	}
	if elasticsearch.IsValidArrayTypeMapping(v) {
		v = v.(string)[1 : len(v.(string))-1]
	}
	tree := map[string]interface{}{}
	tree["type"] = v
	return tree, nil
}

func (db *EventDB) PercolateEventData(eventType string, data map[string]interface{}, id piazza.Ident) (*[]piazza.Ident, error) {
	fixed := map[string]interface{}{}
	fixed["data"] = data
	percolateResponse, err := db.Esi.AddPercolationDocument(eventType, fixed)

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
