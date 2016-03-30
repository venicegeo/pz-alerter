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

package server

import (
	"encoding/json"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
)

type EventDB struct {
	*ResourceDB
}

func NewEventDB(server *Server, es *elasticsearch.Client, index string) (*EventDB, error) {

	esi := elasticsearch.NewIndex(es, index)

	rdb, err := NewResourceDB(server, es, esi)
	if err != nil {
		return nil, err
	}
	erdb := EventDB{ResourceDB: rdb}
	return &erdb, nil
}

func (db *EventDB) PostData(mapping string, obj interface{}, id Ident) (Ident, error) {

	indexResult, err := db.Esi.PostData(mapping, id.String(), obj)
	if err != nil {
		return NoIdent, LoggedError("EventDB.PostData failed: %s", err)
	}
	if !indexResult.Created {
		return NoIdent, LoggedError("EventDB.PostData failed: not created")
	}

	err = db.Esi.Flush()
	if err != nil {
		return NoIdent, err
	}

	return id, nil
}

func (db *EventDB) GetAll(mapping string) (*[]Event, error) {
	var events []Event

	exists := db.Esi.TypeExists(mapping)
	if !exists {
		return &events, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(mapping)
	if err != nil {
		return nil, LoggedError("EventDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, LoggedError("EventDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.Hits != nil {

		for _, hit := range searchResult.Hits.Hits {
			var event Event
			err := json.Unmarshal(*hit.Source, &event)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		}
	}
	return &events, nil
}

func (db *EventDB) GetOne(mapping string, id Ident) (*Event, error) {

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

func (db *EventDB) DeleteByID(mapping string, id Ident) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(mapping, string(id))
	if err != nil {
		return false, LoggedError("EventDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, LoggedError("EventDB.DeleteById failed: no deleteResult")
	}

	err = db.Esi.Flush()
	if err != nil {
		return false, err
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

	err = db.Esi.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (db *EventDB) PercolateEventData(eventType string, data map[string]interface{}, id Ident) (*[]Ident, error) {

	percolateResponse, err := db.Esi.AddPercolationDocument(eventType, data)
	if err != nil {
		return nil, LoggedError("EventDB.PercolateEventData failed: %s", err)
	}
	if percolateResponse == nil {
		return nil, LoggedError("EventDB.PercolateEventData failed: no percolateResult")
	}

	err = db.Flush()
	if err != nil {
		return nil, err
	}

	// add the triggers to the alert queue
	ids := make([]Ident, len(percolateResponse.Matches))
	for i, v := range percolateResponse.Matches {
		ids[i] = Ident(v.Id)
		alert := Alert{ID: db.server.NewIdent(), EventID: id, TriggerID: Ident(v.Id)}
		_, err = db.server.alertDB.PostData(&alert, alert.ID)
		if err != nil {
			return nil, err
		}
	}

	err = db.server.alertDB.Flush()
	if err != nil {
		return nil, err
	}

	return &ids, nil
}
