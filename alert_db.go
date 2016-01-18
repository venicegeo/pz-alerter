package main

import (
)

//---------------------------------------------------------------------------

type AlertDB struct {
	data map[string]Alert
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

func (db *AlertDB) getAll() map[string]Alert {
	return db.data
}

func (alertDB *AlertDB) checkConditions(e Event, conditionDB *ConditionDB) {
	for _, cond := range(conditionDB.data) {
		if cond.Type == e.Type {
			a := newAlert(cond.ID, e.ID)
			alertDB.write(&a)
		}
	}
}
