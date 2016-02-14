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

func NewAlertDB(es *piazza.ElasticSearchService, index string, typename string) (*AlertRDB, error) {
	rdb, err := NewResourceDB(es, index, typename)
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
	searchResult, err := db.es.SearchByTermQuery(db.index, "condition_id", conditionID)
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
