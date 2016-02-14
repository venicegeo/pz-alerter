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

func NewTriggerDB(es *piazza.ElasticSearchService, index string, typename string) (*TriggerDB, error) {
	rdb, err := NewResourceDB(es, index, typename)
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

		match := (cond.Type == event.Type)

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
