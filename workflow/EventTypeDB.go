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

func NewEventTypeDB(service *Service, esi elasticsearch.IIndex) (*EventTypeDB, error) {
	rdb, err := NewResourceDB(service, esi, EventTypeIndexSettings)
	if err != nil {
		return nil, err
	}
	etrdb := EventTypeDB{ResourceDB: rdb, mapping: EventTypeDBMapping}
	return &etrdb, nil
}

func (db *EventTypeDB) PostData(obj interface{}, id piazza.Ident, actor string) (piazza.Ident, error) {
	var eventType EventType
	ok1 := false
	eventType, ok1 = obj.(EventType)
	if !ok1 {
		temp, ok2 := obj.(*EventType)
		if !ok2 {
			return piazza.NoIdent, LoggedError("EventTypeDB.PostData failed: was not given an EventType to post")
		}
		eventType = *temp
	}
	vars, err := piazza.GetVarsFromStruct(eventType.Mapping)
	if err != nil {
		return piazza.NoIdent, LoggedError("EventTypeDB.PostData failed: %s", err)
	}
	for _, v := range vars {
		if !elasticsearch.IsValidMappingType(v) {
			return piazza.NoIdent, LoggedError("EventTypeDB.PostData failed: %v was not recognized as a valid mapping type", v)
		}
	}
	indexResult, err := db.Esi.PostData(db.mapping, id.String(), eventType)
	if err != nil {
		db.service.syslogger.Audit(actor, "createEventType", string(id), "EventTypeDB.PostData: failed")
		return piazza.NoIdent, LoggedError("EventTypeDB.PostData failed: %s", err)
	}
	if !indexResult.Created {
		db.service.syslogger.Audit(actor, "createEventType", string(id), "EventTypeDB.PostData: failed")
		return piazza.NoIdent, LoggedError("EventTypeDB.PostData failed: not created")
	}
	db.service.syslogger.Audit(actor, "createEventType", string(id), "EventTypeDB.PostData: success")

	return id, nil
}

func (db *EventTypeDB) GetAll(format *piazza.JsonPagination, actor string) ([]EventType, int64, error) {
	eventTypes := []EventType{}

	db.service.syslogger.Audit(actor, "readType", db.mapping, "EventTypeDB.GetAll: check type exists")
	exists, err := db.Esi.TypeExists(db.mapping)
	if err != nil {
		return eventTypes, 0, err
	}
	if !exists {
		return eventTypes, 0, nil
	}

	db.service.syslogger.Audit(actor, "readEventTypes", db.mapping, "EventTypeDB.GetAll: match all query")
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

func (db *EventTypeDB) GetEventTypesByDslQuery(dslString string, actor string) ([]EventType, int64, error) {
	eventTypes := []EventType{}

	db.service.syslogger.Audit(actor, "readType", db.mapping, "EventTypeDB.GetEventTypesByDslQuery: check type exists")
	exists, err := db.Esi.TypeExists(db.mapping)
	if err != nil {
		return eventTypes, 0, err
	}
	if !exists {
		return eventTypes, 0, nil
	}

	db.service.syslogger.Audit(actor, "readEventTypes", db.mapping, "EventTypeDB.GetEVentTypesByDslQuery: query eventtypes")
	searchResult, err := db.Esi.SearchByJSON(db.mapping, dslString)
	if err != nil {
		return nil, 0, LoggedError("EventTypeDB.GetEventTypesByDslQuery failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("EventTypeDB.GetEventTypesByDslQuery failed: no searchResult")
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

func (db *EventTypeDB) GetOne(id piazza.Ident, actor string) (*EventType, bool, error) {
	db.service.syslogger.Audit(actor, "readItem", string(id), "EventTypeDB.GetOne: query eventtypes")
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

func (db *EventTypeDB) GetIDByName(name string, actor string) (*piazza.Ident, bool, error) {

	db.service.syslogger.Audit(actor, "readEventTypes", db.mapping, "EventTypeDB.GetIDByName: query eventtypes by name [%s]", name)
	getResult, err := db.Esi.FilterByTermQuery(db.mapping, "name", name)
	if err != nil {
		return nil, getResult.Found, LoggedError("EventTypeDB.GetIDByName failed: %s", err.Error())
	}
	if getResult == nil {
		return nil, true, LoggedError("EventTypeDB.GetIDByName failed: no getResult")
	}

	// This should not happen once we have 1 to 1 mappings of EventTypes to names
	if getResult.NumHits() > 1 {
		return nil, true, LoggedError("EventTypeDB.GetIDByName failed: matched more than one EventType!")
	}

	if getResult.NumHits() == 0 {
		return nil, false, nil
	}

	src := getResult.GetHit(0).Source
	var eventType EventType
	err = json.Unmarshal(*src, &eventType)
	if err != nil {
		return nil, getResult.Found, err
	}

	return &eventType.EventTypeID, getResult.Found, nil
}

func (db *EventTypeDB) DeleteByID(id piazza.Ident, actor string) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(db.mapping, string(id))
	if err != nil {
		db.service.syslogger.Audit(actor, "deleteEventType", string(id), "EventTypeDB.DeleteByID: failed")
		return deleteResult.Found, LoggedError("EventTypeDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		db.service.syslogger.Audit(actor, "deleteEventType", string(id), "EventTypeDB.DeleteByID: failed")
		return false, LoggedError("EventTypeDB.DeleteById failed: no deleteResult")
	}
	db.service.syslogger.Audit(actor, "deleteEventType", string(id), "EventTypeDB.DeleteByID: success")

	return deleteResult.Found, nil
}
