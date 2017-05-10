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

type PercolatorDB struct {
	*ResourceDB
	mapping string
}

func NewPercolatorDB(service *Service, esi elasticsearch.IIndex) (*PercolatorDB, error) {
	rdb, err := NewResourceDB(service, esi)
	if err != nil {
		return nil, err
	}
	prdb := PercolatorDB{ResourceDB: rdb, mapping: PercolatorDBMapping}
	return &prdb, nil
}

func (db *PercolatorDB) PostData(percolator interface{}, id piazza.Ident) error {
	indexResult, err := db.Esi.PostData("queries", id.String(), percolator)
	if err != nil {
		return LoggedError("PercolatorDB.PostData failed: %s", err)
	}
	if !indexResult.Created {
		return LoggedError("PercolatorDB.PostData failed: not created")
	}
	return nil
}

func (db *PercolatorDB) PercolateEventData(eventType string, data map[string]interface{}, id piazza.Ident, actor string) (*[]piazza.Ident, error) {
	fixed := map[string]interface{}{}
	fixed["data"] = data
	percolateResponse, err := db.Esi.AddPercolationDocument(eventType, data)

	if err != nil {
		return nil, LoggedError("EventDB.PercolateEventData failed: %s", err)
	}
	if percolateResponse == nil {
		return nil, LoggedError("EventDB.PercolateEventData failed: no percolateResult")
	}

	// add the triggers to the alert queue
	ids := make([]piazza.Ident, len(percolateResponse.Matches))
	for i, v := range percolateResponse.Matches {
		ids[i] = piazza.Ident(v.Id)
	}

	return &ids, nil
}

func (db *PercolatorDB) AddEventMappingValues(vars map[string]interface{}) error {
	template := `"%s":{"type":"%s"},`
	body := ""
	for k, v := range vars {
		body += fmt.Sprintf(template, k, v)
	}
	body = fmt.Sprintf(`{"properties":{%s}}`, body[0:len(body)-1])
	return db.Esi.SetMapping("doctype", piazza.JsonString(body))
}

func (db *PercolatorDB) GetAll(format *piazza.JsonPagination, actor string) ([]Percolator, int64, error) {
	percolators := []Percolator{}

	exists, err := db.Esi.TypeExists(db.mapping)
	if err != nil {
		return percolators, 0, err
	}
	if !exists {
		return percolators, 0, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(db.mapping, format)
	if err != nil {
		return nil, 0, LoggedError("PercolatorDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("PercolatorDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var percolator Percolator
			if err := json.Unmarshal(*hit.Source, &percolator); err != nil {
				return nil, 0, err
			}
			percolators = append(percolators, percolator)
		}
	}

	return percolators, searchResult.TotalHits(), nil
}

func (db *PercolatorDB) GetOne(id piazza.Ident, actor string) (*Percolator, bool, error) {
	getResult, err := db.Esi.GetByID(db.mapping, id.String())
	if err != nil {
		return nil, false, fmt.Errorf("AlertDB.GetOne failed: %s", err)
	}
	if getResult == nil {
		return nil, true, fmt.Errorf("AlertDB.GetOne failed: %s no getResult", id.String())
	}

	src := getResult.Source
	var percolator Percolator
	if err = json.Unmarshal(*src, &percolator); err != nil {
		return nil, getResult.Found, err
	}

	return &percolator, getResult.Found, nil
}

func (db *PercolatorDB) DeleteByID(id piazza.Ident, actor string) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(db.mapping, string(id))
	if err != nil {
		return deleteResult.Found, fmt.Errorf("PercolatorDB.DeleteById failed: %s", err)
	}
	if deleteResult == nil {
		return false, fmt.Errorf("PercolatorDB.DeleteById failed: no deleteResult")
	}

	if !deleteResult.Found {
		return false, fmt.Errorf("PercolatorDB.DeleteById failed: not found")
	}

	return deleteResult.Found, nil
}
