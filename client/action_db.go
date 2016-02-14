package client

import (
	"encoding/json"
	"github.com/venicegeo/pz-gocommon"
	"log"
	"sync"
)

//---------------------------------------------------------------------------

var actionID = 1

var actionIdLock sync.Mutex

func NewActionIdent() Ident {
	actionIdLock.Lock()
	id := NewIdentFromInt(actionID)
	actionID++
	actionIdLock.Unlock()
	s := "X" + id.String()
	return Ident(s)
}

func NewAction(conditions []Ident, events []Ident, job string) Action {

	id := NewActionIdent()

	return Action{
		ID:         id,
		Conditions: conditions,
		Events:     events,
		Job:        job,
	}
}

//---------------------------------------------------------------------------

type ActionRDB struct {
	*ResourceDB
}

func NewActionRDB(es *piazza.ElasticSearchService, index string, typename string) (*ActionRDB, error) {
	rdb, err := NewResourceDB(es, index, typename)
	if err != nil {
		return nil, err
	}
	ardb := ActionRDB{ResourceDB: rdb}
	return &ardb, nil
}

func ConvertRawsToActions(raws []*json.RawMessage) ([]Action, error) {
	objs := make([]Action, len(raws))
	for i, _ := range raws {
		err := json.Unmarshal(*raws[i], &objs[i])
		if err != nil {
			return nil, err
		}
	}
	return objs, nil
}

func (db *ActionRDB) CheckActions(event Event, conditionDB *ResourceDB, alertDB *AlertRDB) error {
	tmp, err := db.GetAll()
	if err != nil {
		return err
	}

	actions, err := ConvertRawsToActions(tmp)
	if err != nil {
		return err
	}

	for _, action := range actions {
		//log.Printf("e%s.%s ==? c%s.%s", e.ID, e.Type, cond.ID, cond.Type)
		match := true
		var cond Condition
		for _, condId := range(action.Conditions) {
			ok, err := conditionDB.GetById(condId, &cond)
			if err != nil {
				return nil
			}
			if !ok {
				// TODO: this is actually an internal error
				match = false
				break
			}
			if cond.Type != event.Type {
				match = false
				break
			}
		}
		if match {
			alert := NewAlert(NewAlertIdent())
			alert.ActionId = action.ID
			alert.EventId = event.ID
			_, err := alertDB.PostData(&alert, alert.ID)
			if err != nil {
				return err
			}
			log.Printf("INFO: Hit! event %s has triggered action %s, causing alert %s", event.ID, action.ID, alert.ID)
		}
	}
	return nil
}
