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
	"log"
	"sync"
)

//---------------------------------------------------------------------------

var triggerID = 1

var triggerIdLock sync.Mutex

func NewTriggerIdent() Ident {
	triggerIdLock.Lock()
	id := NewIdentFromInt(triggerID)
	triggerID++
	triggerIdLock.Unlock()
	s := "X" + id.String()
	return Ident(s)
}

func NewTrigger(title string, condition Condition, job Job) Trigger {

	id := NewTriggerIdent()

	return Trigger{
		ID:        id,
		Condition: condition,
		Job:       job,
	}
}

//---------------------------------------------------------------------------

type TriggerDB struct {
	*ResourceDB
}

func NewTriggerDB(es *piazza.EsClient, index string, typename string) (*TriggerDB, error) {

	esi := piazza.NewEsIndexClient(es, index)

	rdb, err := NewResourceDB(es, esi, typename)
	if err != nil {
		return nil, err
	}
	ardb := TriggerDB{ResourceDB: rdb}
	return &ardb, nil
}

func ConvertRawsToTriggers(raws []*json.RawMessage) ([]Trigger, error) {
	objs := make([]Trigger, len(raws))
	for i, _ := range raws {
		err := json.Unmarshal(*raws[i], &objs[i])
		if err != nil {
			return nil, err
		}
	}
	return objs, nil
}

func (db *TriggerDB) CheckTriggers(event Event, alertDB *AlertRDB) error {
	tmp, err := db.GetAll()
	if err != nil {
		return err
	}

	triggers, err := ConvertRawsToTriggers(tmp)
	if err != nil {
		return err
	}

	for _, trigger := range triggers {
		cond := trigger.Condition

		match := (cond.EventType == event.EventType)

		if match {
			alert := NewAlert(NewAlertIdent())
			alert.TriggerId = trigger.ID
			alert.EventId = event.ID
			_, err := alertDB.PostData(&alert, alert.ID)
			if err != nil {
				return err
			}
			log.Printf("INFO: Hit! event %s fired trigger %s", event.ID, trigger.ID)
			log.Printf("      Created alert %s", alert.ID)
			log.Printf("      Started task: %s", trigger.Job.Task)
		}
	}
	return nil
}
