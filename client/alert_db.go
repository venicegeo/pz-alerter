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

package client

import (
	"encoding/json"
	"github.com/venicegeo/pz-gocommon"
	"sync"
)

var alertIdLock sync.Mutex
var alertID = 1

func NewAlertIdent() Ident {
	alertIdLock.Lock()
	id := NewIdentFromInt(alertID)
	alertID++
	alertIdLock.Unlock()
	s := "A" + id.String()
	return Ident(s)
}

// newAlert makes an Alert, setting the ID for you.
func NewAlert(triggerId Ident) Alert {

	id := NewIdentFromInt(alertID)
	alertID++
	s := "A" + string(id)

	return Alert{
		ID:        Ident(s),
		TriggerId: triggerId,
	}
}

//---------------------------------------------------------------------------

type AlertRDB struct {
	*ResourceDB
}

func NewAlertDB(es *piazza.EsClient, index string, typename string) (*AlertRDB, error) {

	esi := piazza.NewEsIndexClient(es, index)

	rdb, err := NewResourceDB(es, esi, typename)
	if err != nil {
		return nil, err
	}
	ardb := AlertRDB{ResourceDB: rdb}
	return &ardb, nil
}

func ConvertRawsToAlerts(raws []*json.RawMessage) ([]Alert, error) {
	objs := make([]Alert, len(raws))
	for i, _ := range raws {
		err := json.Unmarshal(*raws[i], &objs[i])
		if err != nil {
			return nil, err
		}
	}
	return objs, nil
}

func (db *AlertRDB) GetByConditionID(conditionID string) ([]Alert, error) {
	searchResult, err := db.Esi.SearchByTermQuery("condition_id", conditionID)
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
