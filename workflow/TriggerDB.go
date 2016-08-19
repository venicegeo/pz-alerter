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
	"net/http"
	"strings"

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

	{ //CHECK SERVICE EXISTS
		jobData := trigger.Job.JobType.Data
		serviceId := jobData["serviceId"]
		strServiceId, ok := serviceId.(string)
		if !ok {
			return piazza.NoIdent, LoggedError("TriggerDB.PostData failed: serviceId field not of type string")
		}
		serviceControllerURL, err := db.service.sys.GetURL("pz-servicecontroller")
		if err == nil {
			// TODO:
			// if err is nil, we have a servicecontroller to talk to
			// if err is not nil, we'll assume we are mocking (which means
			// we have no servicecontroller client to mock)
			response, err := http.Get(serviceControllerURL)
			if err != nil {
				return piazza.NoIdent, LoggedError("TriggerDB.PostData failed to make request to ServiceController: %s", err)
			}
			if response.StatusCode != 200 {
				return piazza.NoIdent, LoggedError("TriggerDB.PostData failed: serviceId %s does not exist", strServiceId)
			}
		}
	}
	{ //CHECK EVENTTYPE IDS
		if len(trigger.Condition.EventTypeIds) == 0 {
			return piazza.NoIdent, LoggedError("TriggerDB.PostData failed: no eventTypeIds were specified")
		}
		for _, id := range trigger.Condition.EventTypeIds {
			_, found, err := db.service.eventTypeDB.GetOne(id)
			if !found || err != nil {
				return piazza.NoIdent, LoggedError("TriggerDB.PostData failed: eventType %s could not be found", id.String())
			}
		}
	}

	ifaceObj := trigger.Condition.Query
	//log.Printf("Query: %v", ifaceObj)
	body, err := json.Marshal(ifaceObj)
	if err != nil {
		return piazza.NoIdent, err
	}

	jsn := string(body)
	//log.Printf("Current json: %s", jsn)
	// Remove trailing }
	jsn = jsn[:len(jsn)-1]
	jsn += ",\"type\":["
	// Add the types that the percolation query can match
	for _, id := range trigger.Condition.EventTypeIds {
		jsn += fmt.Sprintf("\"%s\",", id)
	}
	jsn = jsn[:len(jsn)-1]
	// Add back trailing } and ] to close array
	jsn += "]}"

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

	strTrigger, err := piazza.StructInterfaceToString(trigger)
	if err != nil {
		db.service.eventDB.Esi.DeletePercolationQuery(string(trigger.TriggerId))
		return piazza.NoIdent, LoggedError("TriggerDB.PostData failed: %s", err)
	}
	intTrigger, err := piazza.StructStringToInterface(strTrigger)
	if err != nil {
		db.service.eventDB.Esi.DeletePercolationQuery(string(trigger.TriggerId))
		return piazza.NoIdent, LoggedError("TriggerDB.PostData failed: %s", err)
	}
	mapTrigger, ok := intTrigger.(map[string]interface{})
	if !ok {
		db.service.eventDB.Esi.DeletePercolationQuery(string(trigger.TriggerId))
		return piazza.NoIdent, LoggedError("TriggerDB.PostData failed: bad trigger")
	}
	fixedTrigger := fixTrigger(mapTrigger)

	indexResult2, err := db.Esi.PostData(db.mapping, id.String(), fixedTrigger)
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

func (db *TriggerDB) PutTrigger(id piazza.Ident, trigger *Trigger, update *TriggerUpdate) (*Trigger, error) {
	trigger.Enabled = update.Enabled
	_, err := db.Esi.PostData(db.mapping, id.String(), trigger)
	if err != nil {
		return trigger, LoggedError("TriggerDB.PutData failed: %s", err)
	}
	return trigger, nil
}

func (db *TriggerDB) GetAll(format *piazza.JsonPagination) ([]Trigger, int64, error) {
	triggers := []Trigger{}

	exists, err := db.Esi.TypeExists(db.mapping)
	if err != nil {
		return triggers, 0, err
	}
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

func (db *TriggerDB) GetOne(id piazza.Ident) (*Trigger, bool, error) {

	getResult, err := db.Esi.GetByID(db.mapping, id.String())
	if err != nil {
		return nil, getResult.Found, LoggedError("TriggerDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, true, LoggedError("TriggerDB.GetOne failed: no getResult")
	}

	src := getResult.Source
	var obj Trigger
	err = json.Unmarshal(*src, &obj)
	if err != nil {
		return nil, getResult.Found, err
	}

	return &obj, getResult.Found, nil
}

func (db *TriggerDB) DeleteTrigger(id piazza.Ident) (bool, error) {

	trigger, found, err := db.GetOne(id)
	if err != nil {
		return found, err
	}
	if trigger == nil {
		return false, nil
	}

	deleteResult, err := db.Esi.DeleteByID(db.mapping, string(id))
	if err != nil {
		return deleteResult.Found, LoggedError("TriggerDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, LoggedError("TriggerDB.DeleteById failed: no deleteResult")
	}
	if !deleteResult.Found {
		return false, nil
	}

	deleteResult2, err := db.service.eventDB.Esi.DeletePercolationQuery(string(trigger.PercolationId))
	if err != nil {
		return deleteResult2.Found, LoggedError("TriggerDB.DeleteById percquery failed: %s", err)
	}
	if deleteResult2 == nil {
		return false, LoggedError("TriggerDB.DeleteById percquery failed: no deleteResult")
	}

	return deleteResult2.Found, nil
}

func fixTrigger(input interface{}) interface{} {
	var output interface{}
	switch input.(type) {
	case map[string]interface{}:
		output = fixTriggerNodeMap(input.(map[string]interface{}))
	case []interface{}:
		output = fixTriggerNodeArr(input.([]interface{}))
	}
	return output
}
func fixTriggerNodeMap(inputObj map[string]interface{}) map[string]interface{} {
	outputObj := map[string]interface{}{}
	for k, v := range inputObj {
		switch v.(type) {
		case []interface{}:
			outputObj[strings.Replace(k, ".", "~", -1)] = fixTriggerNodeArr(v.([]interface{}))
		case map[string]interface{}:
			outputObj[strings.Replace(k, ".", "~", -1)] = fixTriggerNodeMap(v.(map[string]interface{}))
		default:
			outputObj[strings.Replace(k, ".", "~", -1)] = v
		}
	}
	return outputObj
}

func fixTriggerNodeArr(inputObj []interface{}) []interface{} {
	outputObj := []interface{}{}
	for _, v := range inputObj {
		switch v.(type) {
		case []interface{}:
			outputObj = append(outputObj, fixTriggerNodeArr(v.([]interface{})))
		case map[string]interface{}:
			outputObj = append(outputObj, fixTriggerNodeMap(v.(map[string]interface{})))
		default:
			outputObj = append(outputObj, v)
		}
	}
	return outputObj
}
