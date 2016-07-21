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

// CronDB TODO
type CronDB struct {
	*ResourceDB
	mapping string
}

// NewCronDB TODO
func NewCronDB(service *WorkflowService, esi elasticsearch.IIndex) (*CronDB, error) {
	rdb, err := NewResourceDB(service, esi, CronIndexSettings)
	if err != nil {
		return nil, err
	}
	crdb := CronDB{ResourceDB: rdb, mapping: cronDBMapping}
	return &crdb, nil
}

// PostData TODO
func (db *CronDB) PostData(obj interface{}, id piazza.Ident) error {

	indexResult, err := db.Esi.PostData(db.mapping, id.String(), obj)
	if err != nil {
		return LoggedError("CronDB.PostData failed: %s", err)
	} else if !indexResult.Created {
		return LoggedError("CronDB.PostData failed: not created")
	}

	return nil
}

// GetAll TODO
func (db *CronDB) GetAll() (*[]Event, error) {
	var events []Event

	exists := db.Esi.TypeExists(db.mapping)
	if !exists {
		return nil, LoggedError("Type %s does not exist", db.mapping)
	}

	searchResult, err := db.Esi.GetAllElements(db.mapping)
	if err != nil {
		return nil, LoggedError("CronDB.GetAll failed: %s", err)
	} else if searchResult == nil {
		return nil, LoggedError("CronDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var event Event
			err := json.Unmarshal(*hit.Source, &event)
			if err != nil {
				return nil, LoggedError("CronDB.GetAll failed: %s", err)
			}
			events = append(events, event)
		}
	}

	return &events, nil
}
