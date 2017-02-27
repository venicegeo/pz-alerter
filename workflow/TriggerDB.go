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

func NewTriggerDB(service *Service, esi elasticsearch.IIndex) (*TriggerDB, error) {
	rdb, err := NewResourceDB(service, esi)
	if err != nil {
		return nil, err
	}
	ardb := TriggerDB{ResourceDB: rdb, mapping: TriggerDBMapping}
	return &ardb, nil
}

func (db *TriggerDB) PostData(trigger *Trigger) error {
	{ //CHECK SERVICE EXISTS
		serviceID := trigger.Job.JobType.Data["serviceId"]
		strServiceID, ok := serviceID.(string)
		if !ok {
			return LoggedError("TriggerDB.PostData failed: serviceId field not of type string")
		}
		serviceControllerURL, err := db.service.sys.GetURL(piazza.PzServiceController)
		if err == nil {
			// TODO:
			// if err is nil, we have a servicecontroller to talk to
			// if err is not nil, we'll assume we are mocking (which means
			// we have no servicecontroller client to mock)
			response, err := http.Get(fmt.Sprintf("%s/service/%s", serviceControllerURL, strServiceID))
			if err != nil {
				return LoggedError("TriggerDB.PostData failed to make request to ServiceController: %s", err)
			}
			// On error, this should close on it's own
			defer func() {
				err = response.Body.Close()
				if err != nil {
					panic(err) // TODO: defer doesn't handle errs well
				}
			}()
			if response.StatusCode != 200 {
				return LoggedError("TriggerDB.PostData failed: serviceID %s does not exist", strServiceID)
			}
		}
	}

	//log.Printf("Query: %v", wrapper)
	body, err := json.Marshal(trigger.Condition)
	if err != nil {
		return err
	}

	//log.Printf("Posting percolation query: %s", body)
	indexResult, err := db.service.eventDB.Esi.AddPercolationQuery(trigger.TriggerID.String(), piazza.JsonString(body))
	if err != nil {
		var errMessage string
		if strings.Contains(err.Error(), "elastic: Error 500 (Internal Server Error): failed to parse query") {
			errMessage = fmt.Sprintf("TriggerDB.PostData addpercquery failed: elastic failed to parse query. Common causes: [Variables do not start with 'data.' or are not found at your specified path, invalid perc query structure].")
		} else {
			errMessage = fmt.Sprintf("TriggerDB.PostData addpercquery failed [unknown cause]: %s ", err)
		}
		return LoggedError(errMessage)
	}
	if indexResult == nil {
		return LoggedError("TriggerDB.PostData addpercquery failed: no indexResult")
	}
	if !indexResult.Created {
		return LoggedError("TriggerDB.PostData addpercquery failed: not created")
	}

	//log.Printf("percolation query added: ID: %s, Type: %s, Index: %s", indexResult.Id, indexResult.Type, indexResult.Index)
	//log.Printf("percolation id: %s", indexResult.Id)
	trigger.PercolationID = piazza.Ident(indexResult.ID)

	strTrigger, err := piazza.StructInterfaceToString(trigger)
	if err != nil {
		_, _ = db.service.eventDB.Esi.DeletePercolationQuery(trigger.TriggerID.String())
		return LoggedError("TriggerDB.PostData failed: %s", err)
	}
	intTrigger, err := piazza.StructStringToInterface(strTrigger)
	if err != nil {
		_, _ = db.service.eventDB.Esi.DeletePercolationQuery(trigger.TriggerID.String())
		return LoggedError("TriggerDB.PostData failed: %s", err)
	}
	mapTrigger, ok := intTrigger.(map[string]interface{})
	if !ok {
		_, _ = db.service.eventDB.Esi.DeletePercolationQuery(trigger.TriggerID.String())
		return LoggedError("TriggerDB.PostData failed: bad trigger")
	}
	fixedTrigger := replaceDot(mapTrigger)

	indexResult2, err := db.Esi.PostData(db.mapping, trigger.TriggerID.String(), fixedTrigger)
	if err != nil {
		_, _ = db.service.eventDB.Esi.DeletePercolationQuery(trigger.TriggerID.String())
		return LoggedError("TriggerDB.PostData failed: %s", err)
	}
	if !indexResult2.Created {
		_, _ = db.service.eventDB.Esi.DeletePercolationQuery(trigger.TriggerID.String())
		return LoggedError("TriggerDB.PostData failed: not created")
	}

	return nil
}

func (db *TriggerDB) PutTrigger(trigger *Trigger, update *TriggerUpdate, actor string) (*Trigger, error) {
	trigger.Enabled = update.Enabled
	strTrigger, err := piazza.StructInterfaceToString(*trigger)
	if err != nil {
		return trigger, LoggedError("TriggerDB.PutData failed: %s", err)
	}
	intTrigger, err := piazza.StructStringToInterface(strTrigger)
	if err != nil {
		return trigger, LoggedError("TriggerDB.PutData failed: %s", err)
	}
	mapTrigger, ok := intTrigger.(map[string]interface{})
	if !ok {
		return trigger, LoggedError("TriggerDB.PutData failed: bad trigger")
	}
	fixedTrigger := replaceDot(mapTrigger)

	_, err = db.Esi.PutData(db.mapping, trigger.TriggerID.String(), fixedTrigger)
	if err != nil {
		return trigger, LoggedError("TriggerDB.PutData failed: %s", err)
	}
	return trigger, nil
}

func (db *TriggerDB) GetAll(format *piazza.JsonPagination, actor string) ([]Trigger, int64, error) {
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
			if err := json.Unmarshal(*hit.Source, &trigger); err != nil {
				return nil, 0, err
			}
			triggers = append(triggers, trigger)
		}
	}
	return triggers, searchResult.TotalHits(), nil
}

func (db *TriggerDB) GetTriggersByDslQuery(dslString string, actor string) ([]Trigger, int64, error) {
	triggers := []Trigger{}

	exists, err := db.Esi.TypeExists(db.mapping)
	if err != nil {
		return triggers, 0, err
	}
	if !exists {
		return triggers, 0, nil
	}

	searchResult, err := db.Esi.SearchByJSON(db.mapping, dslString)
	if err != nil {
		return nil, 0, LoggedError("TriggerDB.GetTriggersByDslQuery failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("TriggerDB.GetTriggersByDslQuery failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var trigger Trigger
			if err := json.Unmarshal(*hit.Source, &trigger); err != nil {
				return nil, 0, err
			}
			triggers = append(triggers, trigger)
		}
	}
	return triggers, searchResult.TotalHits(), nil
}

func (db *TriggerDB) GetOne(id piazza.Ident, actor string) (*Trigger, bool, error) {
	getResult, err := db.Esi.GetByID(db.mapping, id.String())
	if err != nil {
		return nil, false, LoggedError("TriggerDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, false, LoggedError("TriggerDB.GetOne failed: no getResult")
	}

	src := getResult.Source
	var obj Trigger
	if err = json.Unmarshal(*src, &obj); err != nil {
		return nil, getResult.Found, LoggedError("TriggerDB.GetOne failed: %s", err)
	}

	fixedCondition := replaceTilde(obj.Condition)
	obj.Condition = fixedCondition.(map[string]interface{})

	return &obj, getResult.Found, nil
}

func (db *TriggerDB) GetTriggersByEventTypeID(format *piazza.JsonPagination, id piazza.Ident, actor string) ([]Trigger, int64, error) {
	triggers := []Trigger{}

	exists, err := db.Esi.TypeExists(db.mapping)
	if err != nil {
		return triggers, 0, err
	}
	if !exists {
		return triggers, 0, nil
	}

	searchResult, err := db.Esi.FilterByTermQuery(db.mapping, "eventTypeId", id, format)
	if err != nil {
		return nil, 0, LoggedError("TriggerDB.GetTriggersByEventTypeId failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("TriggerDB.GetTriggersByEventTypeId failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var trigger Trigger
			if err := json.Unmarshal(*hit.Source, &trigger); err != nil {
				return nil, 0, err
			}
			triggers = append(triggers, trigger)
		}
	}
	return triggers, searchResult.TotalHits(), nil
}

func (db *TriggerDB) DeleteTrigger(id piazza.Ident, actor string) (bool, error) {
	trigger, found, err := db.GetOne(id, actor)
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

	deleteResult2, err := db.service.eventDB.Esi.DeletePercolationQuery(string(trigger.PercolationID))
	if err != nil {
		return deleteResult2.Found, LoggedError("TriggerDB.DeleteById percquery failed: %s", err)
	}
	if deleteResult2 == nil {
		return false, LoggedError("TriggerDB.DeleteById percquery failed: no deleteResult")
	}

	return deleteResult2.Found, nil
}

func replaceDot(input interface{}) interface{} {
	var output interface{}
	switch input.(type) {
	case map[string]interface{}:
		output = replaceDotMap(input.(map[string]interface{}))
	case []interface{}:
		output = replaceDotArr(input.([]interface{}))
	}
	return output
}
func replaceDotMap(inputObj map[string]interface{}) map[string]interface{} {
	outputObj := map[string]interface{}{}
	for k, v := range inputObj {
		switch v.(type) {
		case []interface{}:
			outputObj[strings.Replace(k, ".", "~", -1)] = replaceDotArr(v.([]interface{}))
		case map[string]interface{}:
			outputObj[strings.Replace(k, ".", "~", -1)] = replaceDotMap(v.(map[string]interface{}))
		default:
			outputObj[strings.Replace(k, ".", "~", -1)] = v
		}
	}
	return outputObj
}
func replaceDotArr(inputObj []interface{}) []interface{} {
	outputObj := []interface{}{}
	for _, v := range inputObj {
		switch v.(type) {
		case []interface{}:
			outputObj = append(outputObj, replaceDotArr(v.([]interface{})))
		case map[string]interface{}:
			outputObj = append(outputObj, replaceDotMap(v.(map[string]interface{})))
		default:
			outputObj = append(outputObj, v)
		}
	}
	return outputObj
}

func replaceTilde(input interface{}) interface{} {
	var output interface{}
	switch input.(type) {
	case map[string]interface{}:
		output = replaceTildeMap(input.(map[string]interface{}))
	case []interface{}:
		output = replaceTildeArr(input.([]interface{}))
	}
	return output
}
func replaceTildeMap(inputObj map[string]interface{}) map[string]interface{} {
	outputObj := map[string]interface{}{}
	for k, v := range inputObj {
		switch v.(type) {
		case []interface{}:
			outputObj[strings.Replace(k, "~", ".", -1)] = replaceTildeArr(v.([]interface{}))
		case map[string]interface{}:
			outputObj[strings.Replace(k, "~", ".", -1)] = replaceTildeMap(v.(map[string]interface{}))
		default:
			outputObj[strings.Replace(k, "~", ".", -1)] = v
		}
	}
	return outputObj
}
func replaceTildeArr(inputObj []interface{}) []interface{} {
	outputObj := []interface{}{}
	for _, v := range inputObj {
		switch v.(type) {
		case []interface{}:
			outputObj = append(outputObj, replaceTildeArr(v.([]interface{})))
		case map[string]interface{}:
			outputObj = append(outputObj, replaceTildeMap(v.(map[string]interface{})))
		default:
			outputObj = append(outputObj, v)
		}
	}
	return outputObj
}

func (db *TriggerDB) addUniqueParamsToQuery(input interface{}, eventType *EventType) interface{} {
	var output interface{}
	switch input.(type) {
	case map[string]interface{}:
		output = db.addUniqueParamsToQueryMap(input.(map[string]interface{}), eventType)
	case []interface{}:
		output = db.addUniqueParamsToQueryArr(input.([]interface{}), eventType)
	}
	return output
}

func (db *TriggerDB) addUniqueParamsToQueryMap(inputObj map[string]interface{}, eventType *EventType) map[string]interface{} {
	outputObj := map[string]interface{}{}
	for k, v := range inputObj {
		switch v.(type) {
		case []interface{}:
			outputObj[db.getNewKeyName(eventType, k)] = db.addUniqueParamsToQueryArr(v.([]interface{}), eventType)
		case map[string]interface{}:
			outputObj[db.getNewKeyName(eventType, k)] = db.addUniqueParamsToQueryMap(v.(map[string]interface{}), eventType)
		default:
			outputObj[db.getNewKeyName(eventType, k)] = v
		}
	}
	return outputObj
}

func (db *TriggerDB) addUniqueParamsToQueryArr(inputObj []interface{}, eventType *EventType) []interface{} {
	outputObj := []interface{}{}
	for _, v := range inputObj {
		switch v.(type) {
		case []interface{}:
			outputObj = append(outputObj, db.addUniqueParamsToQueryArr(v.([]interface{}), eventType))
		case map[string]interface{}:
			outputObj = append(outputObj, db.addUniqueParamsToQueryMap(v.(map[string]interface{}), eventType))
		default:
			outputObj = append(outputObj, v)
		}
	}
	return outputObj
}

func (db *TriggerDB) getNewKeyName(eventType *EventType, key string) string {
	mapping := db.service.removeUniqueParams(eventType.Name, eventType.Mapping)
	vars, _ := piazza.GetVarsFromStruct(mapping)
	for varName := range vars {
		if "data."+varName == key {
			return strings.Replace(key, "data.", "data."+eventType.Name+".", 1)
		}
	}
	return key
}
