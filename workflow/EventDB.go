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
	"bytes"
	"fmt"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

type EventDB struct {
	*ResourceDB
}

func NewEventDB(service *Service, esi elasticsearch.IIndex) (*EventDB, error) {
	rdb, err := NewResourceDB(service, esi)
	if err != nil {
		return nil, err
	}
	erdb := EventDB{ResourceDB: rdb}
	return &erdb, nil
}

func (db *EventDB) PostData(event *Event, typ string) error {
	if err := db.verifyEventReadyToPost(event); err != nil {
		return err
	}

	indexResult, err := db.Esi.PostData(typ, event.EventID.String(), event)
	if err != nil {
		return LoggedError("EventDB.PostData failed: %s", err)
	}
	if !indexResult.Created {
		return LoggedError("EventDB.PostData failed: not created")
	}

	return nil
}

func (db *EventDB) verifyEventReadyToPost(event *Event) error {
	eventTypeJson := db.service.GetEventType(event.EventTypeID, event.CreatedBy)
	eventTypeObj := eventTypeJson.Data
	eventType, ok := eventTypeObj.(*EventType)
	if !ok {
		return LoggedError("EventDB.PostData failed: unable to obtain specified eventtype")
	}
	eventTypeMapping := eventType.Mapping
	eventTypeMappingVars, err := piazza.GetVarsFromStruct(eventTypeMapping)
	if err != nil {
		return LoggedError("EventDB.PostData failed: %s", err)
	}
	exclude := map[string]bool{}
	for k, v := range eventTypeMappingVars {
		if fmt.Sprint(v) == string(elasticsearch.MappingElementTypeGeoPoint) || fmt.Sprint(v) == string(elasticsearch.MappingElementTypeGeoShape) {
			exclude[k] = false
		}
	}
	eventdata := db.service.removeUniqueParams(eventType.Name, event.Data)
	eventDataVars, err := piazza.GetVarsFromStructSkip(eventdata, exclude)
	if err != nil {
		return LoggedError("EventDB.PostData failed: %s", err)
	}
	if len(eventTypeMappingVars) > len(eventDataVars) {
		notFound := []string{}
		for k, _ := range eventTypeMappingVars {
			if _, ok := eventDataVars[k]; !ok {
				notFound = append(notFound, k)
			}
		}
		return LoggedError("EventDB.PostData failed: the variables %s were specified in the EventType but were not found in the Event", notFound)
	} else if len(eventTypeMappingVars) < len(eventDataVars) {
		extra := []string{}
		for k, _ := range eventDataVars {
			if _, ok := eventTypeMappingVars[k]; !ok {
				extra = append(extra, k)
			}
		}
		return LoggedError("EventDB.PostData failed: the variables %s were not specified in the EventType but were found in the Event", extra)
	}
	for k, v := range eventTypeMappingVars {
		for k2, v2 := range eventDataVars {
			if k2 == k {
				if !elasticsearch.IsValidArrayTypeMapping(v) {
					if piazza.ValueIsValidArray(v2) {
						return LoggedError("EventDB.PostData failed: an array was passed into the non-array field %s", k)
					}
				} else {
					if !piazza.ValueIsValidArray(v2) {
						return LoggedError("EventDB.PostData failed: a non-array was pasted into the array field %s", k)
					}
				}
				break
			}
		}
	}
	return nil
}

func (db *EventDB) GetAll(mapping string, format *piazza.JsonPagination, actor string) ([]Event, int64, error) {
	events := []Event{}
	var err error

	exists := true
	if mapping != "" {
		if exists, err = db.Esi.TypeExists(mapping); err != nil {
			return events, 0, err
		}
	}
	if !exists {
		return nil, 0, fmt.Errorf("Type %s does not exist (1)", mapping)
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
			event := &Event{}
			dat, err := hit.Source.MarshalJSON()
			if err != nil {
				return nil, 0, err
			}
			if err = piazza.UnmarshalNumber(bytes.NewReader(dat), event); err != nil {
				return nil, 0, err
			}
			events = append(events, *event)
		}
	}

	return events, searchResult.TotalHits(), nil
}

func (db *EventDB) GetEventsByDslQuery(mapping string, jsnString string, actor string) ([]Event, int64, error) {
	events := []Event{}
	var err error

	exists := true
	if mapping != "" {
		if exists, err = db.Esi.TypeExists(mapping); err != nil {
			return events, 0, err
		}
	}
	if !exists {
		return nil, 0, fmt.Errorf("Type %s does not exist (2)", mapping)
	}

	searchResult, err := db.Esi.SearchByJSON(mapping, jsnString)
	if err != nil {
		return nil, 0, LoggedError("EventDB.GetEventsByDslQuery failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("EventDB.GetEventsByDslQuery failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			event := &Event{}
			dat, err := hit.Source.MarshalJSON()
			if err != nil {
				return nil, 0, err
			}
			if err = piazza.UnmarshalNumber(bytes.NewReader(dat), event); err != nil {
				return nil, 0, err
			}
			events = append(events, *event)
		}
	}

	return events, searchResult.TotalHits(), nil
}

func (db *EventDB) GetEventsByEventTypeID(format *piazza.JsonPagination, mapping string, eventTypeID piazza.Ident, actor string) ([]Event, int64, error) {
	events := []Event{}
	var err error

	exists := true
	if mapping != "" {
		if exists, err = db.Esi.TypeExists(mapping); err != nil {
			return events, 0, err
		}
	}
	if !exists {
		return nil, 0, fmt.Errorf("Type %s does not exist (3)", mapping)
	}

	searchResult, err := db.Esi.FilterByTermQuery(mapping, "eventTypeId", eventTypeID, format)
	if err != nil {
		return nil, 0, LoggedError("EventDB.GetEventsByEventTypeId failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("EventDB.GetEventsByEventTypeId failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			event := &Event{}
			dat, err := hit.Source.MarshalJSON()
			if err != nil {
				return nil, 0, err
			}
			if err = piazza.UnmarshalNumber(bytes.NewReader(dat), event); err != nil {
				return nil, 0, err
			}
			events = append(events, *event)
		}
	}

	return events, searchResult.TotalHits(), nil
}

func (db *EventDB) lookupEventTypeNameByEventID(id piazza.Ident, actor string) (string, error) {
	var mapping string

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
func (db *EventDB) NameExists(name string, actor string) (bool, error) {
	return db.Esi.TypeExists(name)
}

func (db *EventDB) GetOne(mapping string, id piazza.Ident, actor string) (*Event, bool, error) {
	getResult, err := db.Esi.GetByID(mapping, id.String())
	if err != nil {
		return nil, false, LoggedError("EventDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, false, LoggedError("EventDB.GetOne failed: no getResult")
	}

	event := &Event{}
	dat, err := getResult.Source.MarshalJSON()
	if err != nil {
		return nil, getResult.Found, err
	}
	if err = piazza.UnmarshalNumber(bytes.NewReader(dat), event); err != nil {
		return nil, getResult.Found, err
	}

	return event, getResult.Found, nil
}

func (db *EventDB) DeleteByID(mapping string, id piazza.Ident, actor string) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(mapping, string(id))
	if err != nil {
		return deleteResult.Found, LoggedError("EventDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, LoggedError("EventDB.DeleteById failed: no deleteResult")
	}

	return deleteResult.Found, nil
}

func (db *EventDB) AddMapping(name string, mapping map[string]interface{}, actor string) error {
	jsn, err := ConstructEventMappingSchema(name, mapping)
	if err != nil {
		return LoggedError("EventDB.AddMapping failed: %s", err)
	}
	if err = db.Esi.SetMapping(name, jsn); err != nil {
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
	return visitNodeE(input)
}
func visitNodeE(inputObj map[string]interface{}) (map[string]interface{}, error) {
	outputObj := map[string]interface{}{}
	for k, v := range inputObj {
		switch t := v.(type) {
		case string:
			tree, err := visitLeafE(k, v)
			if err != nil {
				return nil, err
			}
			outputObj[k] = tree
		case map[string]interface{}:
			tree, err := visitTreeE(k, v.(map[string]interface{}))
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
func visitTreeE(k string, v map[string]interface{}) (map[string]interface{}, error) {
	subtree, err := visitNodeE(v)
	if err != nil {
		return nil, err
	}
	wrapperTree := map[string]interface{}{}
	wrapperTree["dynamic"] = "strict"
	wrapperTree["properties"] = subtree
	return wrapperTree, nil
}
func visitLeafE(k string, v interface{}) (map[string]interface{}, error) {
	if !elasticsearch.IsValidMappingType(v) {
		return nil, LoggedError("EventDB.ConstructEventMappingSchema failed: \"%#v\" was not recognized as a valid mapping type", v)
	}
	if elasticsearch.IsValidArrayTypeMapping(v) {
		v = v.(string)[1 : len(v.(string))-1]
	}
	tree := map[string]interface{}{}
	tree["type"] = v
	return tree, nil
}

func (db *EventDB) PercolateEventData(eventType string, data map[string]interface{}, id piazza.Ident, actor string) (*[]piazza.Ident, error) {
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
