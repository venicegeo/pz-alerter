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

import "github.com/venicegeo/pz-gocommon/elasticsearch"

type ResourceDB struct {
	Es  *elasticsearch.Client
	Esi *elasticsearch.Index
}

func NewResourceDB(es *elasticsearch.Client, esi *elasticsearch.Index) (*ResourceDB, error) {
	db := &ResourceDB{
		Es:  es,
		Esi: esi,
	}

	_ = esi.Delete()

	err := esi.Create()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *ResourceDB) PostData(mapping string, obj interface{}, id Ident) (Ident, error) {

	// TODO: check IndexResult return value
	_, err := db.Esi.PostData(mapping, id.String(), obj)
	if err != nil {
		return NoIdent, err
	}

	err = db.Esi.Flush()
	if err != nil {
		return NoIdent, err
	}

	return id, nil
}

func (db *ResourceDB) DeleteByID(mapping string, id Ident) (bool, error) {
	res, err := db.Esi.DeleteByID(mapping, string(id))
	if err != nil {
		return false, err
	}

	err = db.Esi.Flush()
	if err != nil {
		return false, err
	}

	return res.Found, nil
}

func (db *ResourceDB) AddMapping(name string, mapping map[string]elasticsearch.MappingElementTypeName) error {

	jsn, err := elasticsearch.ConstructMappingSchema(name, mapping)
	err = db.Esi.SetMapping(name, jsn)
	if err != nil {
		return err
	}

	err = db.Esi.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (db *ResourceDB) Flush() error {

	err := db.Esi.Flush()
	if err != nil {
		return err
	}

	return nil
}
