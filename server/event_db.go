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

func (db *EventDB) GetAll(mapping string) (*[]Event, error) {
	searchResult, err := db.Esi.FilterByMatchAll(mapping)
	if err != nil {
		return nil, err
	}

	var events []Event

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
		return nil, err
	}

	if getResult == nil || !getResult.Found {
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

func (db *EventDB) PercolateEventData(eventType string, data map[string]interface{}, id Ident) (*[]Ident, error) {

	resp, err := db.Esi.AddPercolationDocument(eventType, data)
	if err != nil {
		return nil, err
	}

	db.Flush()

	// add the triggers to the alert queue
	ids := make([]Ident, len(resp.Matches))
	for i, v := range resp.Matches {
		ids[i] = Ident(v.Id)
		alert := Alert{ID: db.server.NewIdent(), EventID: id, TriggerID: Ident(v.Id)}
		_, err = db.server.alertDB.PostData("Alert", &alert, alert.ID)
		if err != nil {
			return nil, err
		}
	}

	db.server.alertDB.Flush()

	return &ids, nil
}
