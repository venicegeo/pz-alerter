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
}

func NewPercolatorDB(service *Service, esi elasticsearch.IIndex) (*PercolatorDB, error) {
	rdb, err := NewResourceDB(service, esi)
	if err != nil {
		return nil, err
	}
	prdb := PercolatorDB{ResourceDB: rdb}
	return &prdb, nil
}

func NewPercolator(doc interface{}) *Percolator {
	return &Percolator{PercolatorQuery{PercolatorPercolate{Field: PercolatorQueryField, DocumentType: PercolatorFieldDBMapping, Document: doc}}}
}

func (db *PercolatorDB) PostPercolatorQuery(percolator interface{}, id piazza.Ident) error {
	indexResult, err := db.Esi.PostData(PercolatorQueryDBMapping, id.String(), percolator)
	if err != nil {
		return LoggedError("PercolatorDB.PostData failed: %s", err)
	} else if !indexResult.Created {
		return LoggedError("PercolatorDB.PostData failed: not created")
	}
	return nil
}
func (db *PercolatorDB) DeletePercolatorQuery(id piazza.Ident) (bool, error) {
	deleteResult, err := db.Esi.DeleteByID(PercolatorQueryDBMapping, string(id))
	if err != nil {
		return deleteResult.Found, fmt.Errorf("PercolatorDB.DeleteById failed: %s", err)
	} else if deleteResult == nil {
		return false, fmt.Errorf("PercolatorDB.DeleteById failed: no deleteResult")
	} else if !deleteResult.Found {
		return false, fmt.Errorf("PercolatorDB.DeleteById failed: not found")
	}
	return deleteResult.Found, nil
}

func (db *PercolatorDB) PercolateEventData(data interface{}) (*[]piazza.Ident, error) {
	percolator := NewPercolator(data)
	dat, err := json.Marshal(*percolator)
	if err != nil {
		return nil, LoggedError("EventDB.PercolateEventData failed: %s", err)
	}
	searchResponse, err := db.Esi.SearchByJSON("", string(dat))
	if err != nil {
		return nil, LoggedError("EventDB.PercolateEventData failed: %s", err)
	} else if searchResponse == nil {
		return nil, LoggedError("EventDB.PercolateEventData failed: no percolateResult")
	}

	// add the triggers to the alert queue
	ids := make([]piazza.Ident, searchResponse.NumHits())
	for i, v := range *searchResponse.GetHits() {
		ids[i] = piazza.Ident(v.ID)
	}

	return &ids, nil
}

func (db *PercolatorDB) AddEventMappingValues(vars map[string]interface{}) error {
	template, body := `"%s":{"type":"%s"},`, ""
	for k, v := range vars {
		body += fmt.Sprintf(template, k, v)
	}
	body = fmt.Sprintf(`{"properties":{%s}}`, body[0:len(body)-1])
	return db.Esi.SetMapping(PercolatorFieldDBMapping, piazza.JsonString(body))
}

func (db *PercolatorDB) GetAll(format *piazza.JsonPagination, actor string) ([]Percolator, int64, error) {
	percolators := PercolatorList{}

	exists, err := db.Esi.TypeExists(PercolatorQueryDBMapping)
	if err != nil {
		return percolators, 0, err
	} else if !exists {
		return percolators, 0, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(PercolatorQueryDBMapping, format)
	if err != nil {
		return nil, 0, LoggedError("PercolatorDB.GetAll failed: %s", err)
	} else if searchResult == nil {
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

func (db *PercolatorDB) GetOne(id piazza.Ident) (*Percolator, bool, error) {
	getResult, err := db.Esi.GetByID(PercolatorQueryDBMapping, id.String())
	if err != nil {
		return nil, false, fmt.Errorf("AlertDB.GetOne failed: %s", err)
	} else if getResult == nil {
		return nil, false, fmt.Errorf("AlertDB.GetOne failed: %s no getResult", id.String())
	}

	src := getResult.Source
	var percolator Percolator
	if err = json.Unmarshal(*src, &percolator); err != nil {
		return nil, getResult.Found, err
	}

	return &percolator, getResult.Found, nil
}
