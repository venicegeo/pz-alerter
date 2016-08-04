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
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

type EventDB struct {
	*ResourceDB
}

type ObjIdent struct {
	Index int
	Type  string
}
type ObjPair struct {
	OpenIndex   int
	ClosedIndex int
}

var closed, open = "closed", "open"

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
	jsn, err := ConstructEventMappingSchema(name, MappingInterfaceToString(mapping))
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
func ConstructEventMappingSchema(name string, mapping string) (piazza.JsonString, error) {
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
	_, esdsl, err := ConvertSchemaToESDSL(mapping, true)
	if err != nil {
		return piazza.JsonString(""), err
	}
	json := fmt.Sprintf(template, name, esdsl)
	return piazza.JsonString(json), nil
}

func ConvertSchemaToESDSL(str string, removeEnds bool) (string, string, error) {
	str = elasticsearch.RemoveWhitespace(str)
	if removeEnds {
		str = str[1 : len(str)-1]
	}
	//-------------Find all open and closed brackets----------------------------
	idents := []ObjIdent{}
	for i := 0; i < len(str); i++ {
		char := elasticsearch.CharAt(str, i)
		if char == "{" {
			idents = append(idents, ObjIdent{i, open})
		} else if char == "}" {
			idents = append(idents, ObjIdent{i, closed})
		}
	}
	//-------------Match brackets into pairs------------------------------------
	pairs := []ObjPair{}
	pairMap := map[int]int{}
	for len(idents) > 0 {
		for i := 0; i < len(idents)-1; i++ {
			a := idents[i]
			b := idents[i+1]
			if a.Type == open && b.Type == closed {
				pairMap[a.Index] = b.Index
				idents = append(idents[:i], idents[i+1:]...)
				idents = append(idents[:i], idents[i+1:]...)
				break
			}
		}
	}
	//-------------Sort pairs based off open bracket index----------------------
	keys := []int{}
	for k, _ := range pairMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		v := pairMap[k]
		pairs = append(pairs, ObjPair{k, v})
	}
	//-------------Add properties after each open index-------------------------
	for i := 0; i < len(pairs); i++ {
		str = elasticsearch.InsertString(str, `"properties":{`, pairs[i].OpenIndex+1)
		for j := i + 1; j < len(pairs); j++ {
			pairs[j].OpenIndex += 14
		}
		for j := 0; j < len(pairs); j++ {
			if pairs[j].ClosedIndex >= pairs[i].OpenIndex {
				pairs[j].ClosedIndex += 14
			}
		}
	}
	//-------------Add a closed bracket at each closed index--------------------
	for i := 0; i < len(pairs); i++ {
		str = elasticsearch.InsertString(str, `}`, pairs[i].ClosedIndex)
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].OpenIndex >= pairs[i].ClosedIndex {
				pairs[j].OpenIndex++
			}
		}
		for j := 0; j < len(pairs); j++ {
			if pairs[j].ClosedIndex >= pairs[i].ClosedIndex {
				pairs[j].ClosedIndex++
			}
		}
	}
	//-------------Seperate pieces of mapping onto seperate lines---------------
	temp := ""
	for i := 0; i < len(str); i++ {
		if elasticsearch.CharAt(str, i) == "{" || elasticsearch.CharAt(str, i) == "}" || elasticsearch.CharAt(str, i) == "," {
			temp += "\n" + elasticsearch.CharAt(str, i) + "\n"
		} else {
			temp += elasticsearch.CharAt(str, i)
		}
	}
	lines := strings.Split(temp, "\n")
	fixed := []string{}
	//-------------Format individual values-------------------------------------
	for _, line := range lines {
		if strings.Contains(line, `":"`) {
			formatedLine, err := formatKeyValue(line)
			if err != nil {
				return "", "", err
			}
			fixed = append(fixed, formatedLine)
		} else {
			fixed = append(fixed, line)
		}
	}
	temp = ""
	for _, line := range fixed {
		temp += line
	}
	return temp, "{" + temp + "}", nil
}

func formatKeyValue(str string) (string, error) {
	parts := strings.Split(str, ":")
	value := strings.Replace(parts[1], "\"", "", -1)
	valid := elasticsearch.IsValidMappingType(value)
	if !valid {
		return "", errors.New(fmt.Sprintf(" '%s' was not recognized as a valid mapping type", value))
	}
	res := fmt.Sprintf(`%s:{"type":%s}`, parts[0], parts[1])
	return res, nil
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
