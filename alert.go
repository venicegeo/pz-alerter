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
	ID        string `json:"id"`
	Condition string `json:"condition" binding:"required"`
}

type AlertDB struct {
	data map[string]Alert
}

func newAlertDB() *AlertDB {
	db := new(AlertDB)
	db.data = make(map[string]Alert)
	return db
}

func (db *AlertDB) write(alert *Alert) error {
	alert.ID = newAlertID()
	db.data[alert.ID] = *alert
	return nil
}
