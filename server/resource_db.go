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
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-workflow/common"
)

var resourceID = 1

func NewResourceID() common.Ident {
	id := common.NewIdentFromInt(resourceID)
	resourceID++
	return common.Ident("R" + string(id))
}

//type Resource interface {
//	GetId() Ident
//	SetId(Ident)
//}

type ResourceDB struct {
	Es       *piazza.EsClient
	Esi      *piazza.EsIndexClient
	Typename string
}

func NewResourceDB(es *piazza.EsClient, esi *piazza.EsIndexClient, typename string) (*ResourceDB, error) {
	db := &ResourceDB{
		Es:       es,
		Esi:       esi,
		Typename: typename,
	}

	err := esi.Delete()
	if err != nil {
		return nil, err
	}

	err = esi.Create()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *ResourceDB) PostData(obj interface{}, id common.Ident) (common.Ident, error) {

	_, err := db.Esi.PostData(db.Typename, id.String(), obj)
	if err != nil {
		return common.NoIdent, err
	}

	err = db.Esi.Flush()
	if err != nil {
		return common.NoIdent, err
	}

	return id, nil
}

func (db *ResourceDB) GetAll() ([]*json.RawMessage, error) {
	searchResult, err := db.Esi.SearchByMatchAll()
	if err != nil {
		return nil, err
	}

	if searchResult.Hits == nil {
		return nil, nil
	}

	raws := make([]*json.RawMessage, searchResult.TotalHits())

	for i, hit := range searchResult.Hits.Hits {
		raws[i] = hit.Source
	}

	return raws, nil
}

func (db *ResourceDB) GetById(id common.Ident, obj interface{}) (bool, error) {

	getResult, err := db.Esi.GetById(id.String())
	if err != nil {
		return false, err
	}

	if !getResult.Found {
		return false, nil
	}

	src := getResult.Source
	err = json.Unmarshal(*src, obj)
	if err != nil {
		return true, err
	}
	return true, nil
}

func (db *ResourceDB) DeleteByID(id string) (bool, error) {
	res, err := db.Esi.DeleteById(db.Typename, id)
	if err != nil {
		return false, err
	}

	err = db.Esi.Flush()
	if err != nil {
		return false, err
	}

	return res.Found, nil
}

func (db *ResourceDB) AddMapping(name string, jsn piazza.JsonString) (error) {
	err := db.Esi.SetMapping(name, jsn)
	if err != nil {
		return err
	}
	return nil
}
