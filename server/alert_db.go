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

func NewAlertDB(es *elasticsearch.ElasticsearchClient, index string) (*AlertDB, error) {

	esi := elasticsearch.NewElasticsearchIndex(es, index)

	rdb, err := NewResourceDB(es, esi)
	if err != nil {
		return nil, err
	}
	ardb := AlertDB{ResourceDB: rdb}
	return &ardb, nil
}

func (db *AlertDB) GetByConditionID(mapping string, conditionID string) ([]Alert, error) {
	searchResult, err := db.Esi.FilterByTermQuery(mapping, "condition_id", conditionID)
	if err != nil {
		return nil, err
	}

	if searchResult.Hits == nil {
		return nil, nil
	}

	var as []Alert
	for _, hit := range searchResult.Hits.Hits {
		var a Alert
		err := json.Unmarshal(*hit.Source, &a)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return as, nil
}
