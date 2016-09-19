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

type TestElasticsearchDB struct {
	*ResourceDB
	typ string
}

type TestElasticsearchBody struct {
	ID    piazza.Ident `json:"id"`
	Value int          `json:"value"`
}

func NewTestElasticsearchDB(service *Service, esi elasticsearch.IIndex) (*TestElasticsearchDB, error) {

	const mapping = `{
		"mapping" : {
			"Obj2":{
				"properties":{
					"id2": {
						"type":"integer"
					},
					"data2": {
						"type":"string"
					},
					"foo2": {
						"type":"boolean"
					}
				}
			}
		}
	}`

	rdb, err := NewResourceDB(service, esi, mapping)
	if err != nil {
		return nil, err
	}

	typ := "Obj2"
	etrdb := TestElasticsearchDB{ResourceDB: rdb, typ: typ}

	/*time.Sleep(5 * time.Second)

	ok, err := esi.IndexExists()
	if err != nil {
		return nil, fmt.Errorf("Index %s failes to exist after creation", esi.IndexName())
	}
	if !ok {
		return nil, fmt.Errorf("Index %s does not exist after creation", esi.IndexName())
	}

	ok, err = esi.TypeExists(TestElasticsearchDBMapping)
	if err != nil {
		return nil, fmt.Errorf("Type %s fails to exist in index %s after creation", TestElasticsearchDBMapping, esi.IndexName())
	}
	if !ok {
		return nil, fmt.Errorf("Index %s does not exist in index %s after creation", TestElasticsearchDBMapping, esi.IndexName())
	}*/

	err = esi.SetMapping(typ, piazza.JsonString(mapping))
	if err != nil {
		return nil, err
	}

	return &etrdb, nil
}

func (db *TestElasticsearchDB) PostData(obj interface{}, id piazza.Ident) (piazza.Ident, error) {
	var p *TestElasticsearchBody
	ok1 := false
	temp1, ok1 := obj.(TestElasticsearchBody)
	if !ok1 {
		temp2, ok2 := obj.(*TestElasticsearchBody)
		if !ok2 {
			return piazza.NoIdent, LoggedError("TestElasticsearchDB.PostData failed: was not given an TestElasticsearchBody to post")
		}
		p = temp2
	} else {
		p = &temp1
	}

	indexResult, err := db.Esi.PostData(db.typ, id.String(), p)
	if err != nil {
		return piazza.NoIdent, LoggedError("TestElasticsearchDB.PostData failed: %s\n%#v\n%#v", err, db.typ, p)
	}
	if !indexResult.Created {
		return piazza.NoIdent, LoggedError("TestElasticsearchDB.PostData failed: not created")
	}

	return id, nil
}

func (db *TestElasticsearchDB) GetAll(format *piazza.JsonPagination) ([]TestElasticsearchBody, int64, error) {
	bodies := []TestElasticsearchBody{}

	exists, err := db.Esi.TypeExists(db.typ)
	if err != nil {
		return bodies, 0, err
	}
	if !exists {
		return bodies, 0, nil
	}

	searchResult, err := db.Esi.FilterByMatchAll(db.typ, format)
	if err != nil {
		return nil, 0, LoggedError("TestElasticsearchDB.GetAll failed: %s", err)
	}
	if searchResult == nil {
		return nil, 0, LoggedError("TestElasticsearchDB.GetAll failed: no searchResult")
	}

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			var body TestElasticsearchBody
			err := json.Unmarshal(*hit.Source, &body)
			if err != nil {
				return nil, 0, err
			}
			bodies = append(bodies, body)
		}
	}

	return bodies, searchResult.TotalHits(), nil
}

func (db *TestElasticsearchDB) GetVersion() (string, error) {
	v := db.Esi.GetVersion()
	return v, nil
}
