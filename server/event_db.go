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

type EventDB struct {
	*ResourceDB
}

func NewEventDB(es *piazza.EsClient, index string) (*EventDB, error) {

	esi := piazza.NewEsIndexClient(es, index)

	rdb, err := NewResourceDB(es, esi)
	if err != nil {
		return nil, err
	}
	erdb := EventDB{ResourceDB: rdb}
	return &erdb, nil
}

func (db *EventDB) PercolateEventData(eventType string, data map[string]interface{}, id common.Ident, alertDB *AlertDB) (*[]common.Ident, error) {

	resp, err := db.Esi.AddPercolationDocument(eventType, data)
	if err != nil {
		return nil, err
	}

	// add the triggers to the alert queue
	ids := make([]common.Ident, len(resp.Matches))
	for i, v := range resp.Matches {
		ids[i] = common.Ident(v.Id)
		alert := common.Alert{ID: common.NewIdent(), EventId: id, TriggerId: common.Ident(v.Id)}
		_, err = alertDB.PostData("Alert", &alert, alert.ID)
		if err != nil {
			return nil, err
		}
	}

	return &ids, nil
}

func (db *EventDB) GetByMapping(mapping string) ([]common.Event, error) {

	searchResult, err := db.Esi.SearchByMatchAllWithMapping(mapping)
	if err != nil {
		return nil, err
	}

	if searchResult.Hits == nil {
		return nil, nil
	}

	ary := make([]common.Event, searchResult.TotalHits())

	for i, hit := range searchResult.Hits.Hits {
		var tmp common.Event
		err = json.Unmarshal([]byte(*hit.Source), tmp)
		if err != nil {
			return nil, err
		}
		ary[i] = tmp
	}
	return ary, nil
}
