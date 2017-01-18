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
	mapping string
}

type TestElasticsearchBody struct {
	ID    piazza.Ident `json:"id"`
	Value int          `json:"value"`
}

const TestElasticsearchSettings = `{
	"TestElasticsearch":{
		"properties":{
			"id": {
				"type":"string"
			},
			"data": {
				"type":"string"
			},
			"tags": {
				"type":"string"
			}
		}
	}
}`

const TestElasticsearchMapping = "TestElasticsearch"

func NewTestElasticsearchDB(service *Service, esi elasticsearch.IIndex) (*TestElasticsearchDB, error) {

	rdb, err := NewResourceDB(service, esi, "")
	if err != nil {
		return nil, err
	}

	err = esi.SetMapping(TestElasticsearchMapping, TestElasticsearchSettings)
	if err != nil {
		return nil, err
	}

	etrdb := TestElasticsearchDB{ResourceDB: rdb, mapping: TestElasticsearchMapping}

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

	return &etrdb, nil
}

func (db *TestElasticsearchDB) PostData(obj interface{}, id piazza.Ident) (piazza.Ident, error) {
	var p *TestElasticsearchBody
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

	indexResult, err := db.Esi.PostData(db.mapping, id.String(), p)
	if err != nil {
		return piazza.NoIdent, LoggedError("TestElasticsearchDB.PostData failed: %s\n%#v\n%#v", err, db.mapping, p)
	}
	if !indexResult.Created {
		return piazza.NoIdent, LoggedError("TestElasticsearchDB.PostData failed: not created")
	}

	return id, nil
}

func (db *TestElasticsearchDB) GetOne(id piazza.Ident) (*TestElasticsearchBody, bool, error) {
	getResult, err := db.Esi.GetByID(db.mapping, id.String())
	if err != nil {
		return nil, getResult.Found, LoggedError("TestElasticsearchDB.GetOne failed: %s", err.Error())
	}
	if getResult == nil {
		return nil, true, LoggedError("TestElasticsearchDB.GetOne failed: no getResult")
	}

	src := getResult.Source
	var body TestElasticsearchBody
	err = json.Unmarshal(*src, &body)
	if err != nil {
		return nil, getResult.Found, err
	}

	return &body, getResult.Found, nil
}

func (db *TestElasticsearchDB) GetVersion() (string, error) {
	v := db.Esi.GetVersion()
	return v, nil
}
