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
}

func NewAlertDB(es *elasticsearch.Client, index string) (*AlertDB, error) {

	esi := elasticsearch.NewIndex(es, index)

	rdb, err := NewResourceDB(es, esi)
	if err != nil {
		return nil, err
	}
	ardb := AlertDB{ResourceDB: rdb}
	return &ardb, nil
}

func (db *AlertDB) GetAll(mapping string) (*[]Alert, error) {
	searchResult, err := db.Esi.FilterByMatchAll(mapping)

	if err != nil {
		return nil, err
	}

	var alerts []Alert

	if searchResult.Hits != nil {
		for _, hit := range searchResult.Hits.Hits {
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

func (db *AlertDB) GetOne(mapping string, id Ident) (*Alert, error) {

	getResult, err := db.Esi.GetByID(mapping, id.String())
	if err != nil {
		return nil, err
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
