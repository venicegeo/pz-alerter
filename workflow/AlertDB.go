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
	"log"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
)

type AlertDB struct {
	*ResourceDB
	mapping string
}

// The default settings for our Elasticsearch alerts index
// Explanation:
//   "index.mapper.dynamic": false
//     This removes the ability for us to create new fields within documents
//     after the type has been defined, rather than adding them in dynamically
//     if they aren't recognized
//   "index": "not_analyzed"
//     This means that these properties are not analyzed by Elasticsearch.
//     Previously, these ids were analyzed by ES and thus broken up into chunks;
//     in the case of a UUID this would happen via break-up by the "-" character.
//     For example, the UUID "ab3142cd-1a8e-44f8-6a01-5ce8a9328fb2" would be broken
//     into "ab3142cd", "1a8e", "44f8", "6a01" and "5ce8a9328fb2", and queries would
//     match on all of these separate strings, which was undesired behavior.
const (
	alertIndexSettings = `
{
	"settings": {
		"index.mapper.dynamic": false
	}
	"mappings": {
		"Alert": {
			"properties": {
				"alertId": {
					"type": "string",
					"index": "not_analyzed"
				}
				"triggerId": {
					"type": "string",
					"index": "not_analyzed"
				}
				"jobId": {
					"type": "string",
					"index": "not_analyzed"
				}
				"eventId": {
					"type": "string",
					"index": "not_analyzed"
				}
			}
		}
	}
}
`
)

func NewAlertDB(service *WorkflowService, esi elasticsearch.IIndex) (*AlertDB, error) {

	rdb, err := NewResourceDB(service, esi, "")
	if err != nil {
		return nil, err
	}
	ardb := AlertDB{ResourceDB: rdb, mapping: "Alert"}
	return &ardb, nil
}

func (db *AlertDB) PostData(obj interface{}, id Ident) (Ident, error) {

	indexResult, err := db.Esi.PostData(db.mapping, id.String(), obj)
	if err != nil {
		return NoIdent, LoggedError("AlertDB.PostData failed: %s", err)
	}
	if !indexResult.Created {
		return NoIdent, LoggedError("AlertDB.PostData failed: not created")
	}

	return id, nil
}

func (db *AlertDB) GetAll(format elasticsearch.QueryFormat) (*[]Alert, int64, error) {
	var alerts []Alert
	var count = int64(-1)

	exists := db.Esi.TypeExists(db.mapping)
	if !exists {
		return &alerts, count, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(db.mapping, format)
	if err != nil {
		return nil, count, LoggedError("AlertDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, count, LoggedError("AlertDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		count = searchResult.NumberMatched()
		for _, hit := range *searchResult.GetHits() {
			var alert Alert
			err := json.Unmarshal(*hit.Source, &alert)
			if err != nil {
				return nil, count, err
			}
			alerts = append(alerts, alert)
		}
	}

	return &alerts, count, nil
}

func (db *AlertDB) GetAllByTrigger(format elasticsearch.QueryFormat, triggerId string) (*[]Alert, int64, error) {

	alerts := []Alert{}
	var count = int64(-1)

	exists := db.Esi.TypeExists(db.mapping)
	if !exists {
		return &alerts, count, nil
	}

	log.Printf("Type exists: %s", db.mapping)

	// This will be an Elasticsearch term query of roughly the following structure:
	// { "term": { "_id": triggerId } }
	// This matches the '_id' field of the Elasticsearch document exactly
	searchResult, err := db.Esi.FilterByTermQuery(db.mapping, "triggerId", triggerId)
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, count, LoggedError("AlertDB.GetAllByTrigger failed: %s", err)
	}
	if searchResult == nil {
		log.Printf("Search returned nil")
		return nil, count, LoggedError("AlertDB.GetAllByTrigger failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		count = searchResult.NumberMatched()
		// If we don't find any alerts by the given triggerId, don't error out, just return an empty list
		if count == 0 {
			return &alerts, count, nil
		}
		log.Printf("Adding %d search results", count)
		for _, hit := range *searchResult.GetHits() {
			var alert Alert
			log.Printf("Adding search result: %v", *hit.Source)
			err := json.Unmarshal(*hit.Source, &alert)
			if err != nil {
				return nil, count, err
			}
			alerts = append(alerts, alert)
		}
	}

	log.Printf("Returning alerts by trigger...")
	return &alerts, count, nil
}

func (db *AlertDB) GetOne(id Ident) (*Alert, error) {

	getResult, err := db.Esi.GetByID(db.mapping, id.String())
	if err != nil {
		return nil, LoggedError("AlertDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, LoggedError("AlertDB.GetOne failed: no getResult")
	}

	if !getResult.Found {
		return nil, nil
	}

	src := getResult.Source
	var alert Alert
	err = json.Unmarshal(*src, &alert)
	if err != nil {
		return nil, err
	}

	return &alert, nil
}

func (db *AlertDB) DeleteByID(id Ident) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(db.mapping, string(id))
	if err != nil {
		return false, LoggedError("AlertDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, LoggedError("AlertDB.DeleteById failed: no deleteResult")
	}

	return deleteResult.Found, nil
}
