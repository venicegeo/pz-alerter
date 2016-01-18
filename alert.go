package main

import (
	"strconv"
)

//---------------------------------------------------------------------------
//---------------------------------------------------------------------------

var alertID = 1

func newAlertID() string {
	s := strconv.Itoa(alertID)
	alertID++
	return s
}

type Alert struct {
	ID          string `json:"id"`
	ConditionID string `json:"condition_id" binding:"required"`
	EventID     string `json:"condition_id" binding:"required"`
}

// newAlert makes an Alert, setting the ID for you.
func newAlert(conditionID string, eventID string) Alert {
	return Alert{
		ID:          newAlertID(),
		ConditionID: conditionID,
		EventID: eventID,
	}
}

type AlertDB struct {
	data map[string]Alert
}

func (item *Alert) id() string {
	return item.ID
}

func newAlertDB() *AlertDB {
	db := new(AlertDB)
	db.data = make(map[string]Alert)
	return db
}

func (db *AlertDB) write(alert *Alert) error {
	db.data[alert.ID] = *alert
	return nil
}

func (db *AlertDB) getByID(conditionID string) *Alert {
	for _, v := range db.data {
		if v.ID == conditionID {
			return &v
		}
	}
	return nil
}
