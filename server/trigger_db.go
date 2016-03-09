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
	"errors"
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-workflow/common"
)

type TriggerDB struct {
	*ResourceDB
}

func NewTriggerDB(es *piazza.EsClient, index string) (*TriggerDB, error) {

	esi := piazza.NewEsIndexClient(es, index)

	rdb, err := NewResourceDB(es, esi)
	if err != nil {
		return nil, err
	}
	ardb := TriggerDB{ResourceDB: rdb}
	return &ardb, nil
}

func (db *TriggerDB) PostTrigger(mapping string, trigger *common.Trigger, id common.Ident, eventDB *EventDB) (common.Ident, error) {

	ifaceObj := trigger.Condition.Query
	body, err := json.Marshal(ifaceObj)
	if err != nil {
		return common.NoIdent, err
	}

	indexResult, err := eventDB.Esi.AddPercolationQuery(string(trigger.ID), piazza.JsonString(body))
	if err != nil {
		return common.NoIdent, err
	}

	trigger.PercolationID = common.Ident(indexResult.Id)

	_, err = db.Esi.PostData(mapping, id.String(), trigger)
	if err != nil {
		return common.NoIdent, err
	}

	err = db.Esi.Flush()
	if err != nil {
		return common.NoIdent, err
	}

	return id, nil
}

func (db *TriggerDB) DeleteTrigger(mapping string, id common.Ident, eventDB *EventDB) (bool, error) {

	var obj common.Trigger
	ok, err := db.GetById(mapping, id, &obj)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	res, err := db.Esi.DeleteById(mapping, string(id))
	if err != nil {
		return false, err
	}

	deleteResult, err := eventDB.Esi.DeletePercolationQuery(string(obj.PercolationID))
	err = db.Esi.Flush()
	if err != nil {
		return false, err
	}
	if !deleteResult.Found {
		return false, errors.New("unable to delete percolation")
	}

	return res.Found, nil
}
