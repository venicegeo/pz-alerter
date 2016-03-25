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

//---------------------------------------------------------------------------

type EventTypeDB struct {
	*ResourceDB
}

func NewEventTypeDB(server *Server, es *elasticsearch.Client, index string) (*EventTypeDB, error) {

	esi := elasticsearch.NewIndex(es, index)

	rdb, err := NewResourceDB(server, es, esi)
	if err != nil {
		return nil, err
	}
	etrdb := EventTypeDB{ResourceDB: rdb}
	return &etrdb, nil
}

func (db *EventTypeDB) GetAll(mapping string) (*[]EventType, error) {
	searchResult, err := db.Esi.FilterByMatchAll(mapping)
	if err != nil {
		return nil, err
	}

	var eventTypes []EventType

	if searchResult != nil && searchResult.Hits != nil {
		for _, hit := range searchResult.Hits.Hits {
			var eventType EventType
			err := json.Unmarshal(*hit.Source, &eventType)
			if err != nil {
				return nil, err
			}
			eventTypes = append(eventTypes, eventType)
		}
	}

	return &eventTypes, nil
}

func (db *EventTypeDB) GetOne(mapping string, id Ident) (*EventType, error) {

	getResult, err := db.Esi.GetByID(mapping, id.String())
	if err != nil {
		return nil, err
	}

	if getResult == nil || !getResult.Found {
		return nil, nil
	}

	src := getResult.Source
	var eventType EventType
	err = json.Unmarshal(*src, &eventType)
	if err != nil {
		return nil, err
	}

	return &eventType, nil
}
