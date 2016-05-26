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

type AlertDB struct {
	*ResourceDB
	mapping string
}

func NewAlertDB(server *Server, esi elasticsearch.IIndex) (*AlertDB, error) {

	rdb, err := NewResourceDB(server, esi)
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

func (db *AlertDB) GetAll(format elasticsearch.QueryFormat) (*[]Alert, error) {

	alerts := []Alert{}

	exists := db.Esi.TypeExists(db.mapping)
	if !exists {
		return &alerts, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(db.mapping, format)
	if err != nil {
		return nil, LoggedError("AlertDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, LoggedError("AlertDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var alert Alert
			err := json.Unmarshal(*hit.Source, &alert)
			if err != nil {
				return nil, err
			}
			alerts = append(alerts, alert)
		}
	}

	return &alerts, nil
}

func (db *AlertDB) GetAllWithCount(format elasticsearch.QueryFormat) (*[]Alert, int64, error) {

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
