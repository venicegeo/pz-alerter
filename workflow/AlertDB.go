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

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

type AlertDB struct {
	*ResourceDB
	mapping string
}

func NewAlertDB(service *Service, esi elasticsearch.IIndex) (*AlertDB, error) {
	rdb, err := NewResourceDB(service, esi, AlertIndexSettings)
	if err != nil {
		return nil, err
	}
	ardb := AlertDB{ResourceDB: rdb, mapping: AlertDBMapping}
	return &ardb, nil
}

func (db *AlertDB) PostData(obj interface{}, id piazza.Ident) (piazza.Ident, error) {

	indexResult, err := db.Esi.PostData(db.mapping, id.String(), obj)
	if err != nil {
		return piazza.NoIdent, LoggedError("AlertDB.PostData failed: %s", err)
	}
	if !indexResult.Created {
		return piazza.NoIdent, LoggedError("AlertDB.PostData failed: not created")
	}

	return id, nil
}

func (db *AlertDB) GetAll(format *piazza.JsonPagination) ([]Alert, int64, error) {
	alerts := []Alert{}

	exists, err := db.Esi.TypeExists(db.mapping)
	if err != nil {
		return alerts, 0, err
	}
	if !exists {
		return alerts, 0, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(db.mapping, format)
	if err != nil {
		return nil, 0, LoggedError("AlertDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("AlertDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var alert Alert
			err := json.Unmarshal(*hit.Source, &alert)
			if err != nil {
				return nil, 0, err
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, searchResult.TotalHits(), nil
}

func (db *AlertDB) GetAllByTrigger(format *piazza.JsonPagination, triggerID piazza.Ident) ([]Alert, int64, error) {
	alerts := []Alert{}
	var count = int64(-1)

	exists, err := db.Esi.TypeExists(db.mapping)
	if err != nil {
		return alerts, 0, err
	}
	if !exists {
		return alerts, 0, nil
	}

	// This will be an Elasticsearch term query of roughly the following structure:
	// { "term": { "_id": triggerId } }
	// This matches the '_id' field of the Elasticsearch document exactly
	searchResult, err := db.Esi.FilterByTermQuery(db.mapping, "triggerId", triggerID)
	if err != nil {
		return nil, 0, LoggedError("AlertDB.GetAllByTrigger failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("AlertDB.GetAllByTrigger failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		count = searchResult.TotalHits()
		// If we don't find any alerts by the given triggerId, don't error out, just return an empty list
		if count == 0 {
			return alerts, 0, nil
		}
		for _, hit := range *searchResult.GetHits() {
			var alert Alert
			err := json.Unmarshal(*hit.Source, &alert)
			if err != nil {
				return nil, 0, err
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, searchResult.TotalHits(), nil
}

func (db *AlertDB) GetOne(id piazza.Ident) (*Alert, bool, error) {
	getResult, err := db.Esi.GetByID(db.mapping, id.String())
	if err != nil {
		return nil, false, fmt.Errorf("AlertDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, true, fmt.Errorf("AlertDB.GetOne failed: %s no getResult", id.String())
	}

	src := getResult.Source
	var alert Alert
	err = json.Unmarshal(*src, &alert)
	if err != nil {
		return nil, getResult.Found, err
	}

	return &alert, getResult.Found, nil
}

func (db *AlertDB) DeleteByID(id piazza.Ident) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(db.mapping, string(id))
	if err != nil {
		return deleteResult.Found, fmt.Errorf("AlertDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, fmt.Errorf("AlertDB.DeleteById failed: no deleteResult")
	}

	if !deleteResult.Found {
		return false, fmt.Errorf("AlertDB.DeleteById failed: not found")
	}

	return deleteResult.Found, nil
}
