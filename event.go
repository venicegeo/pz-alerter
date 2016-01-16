package main

import (
	"strconv"
)

//---------------------------------------------------------------------------
//---------------------------------------------------------------------------

var eventID = 1
func newEventID() string {
	s := strconv.Itoa(eventID)
	eventID++
	return s
}

const EventDataIngested = "DataIngested"
const EventDataAccessed = "DataAccessed"
const EventUSDataFound = "USDataFound"

type Event struct {
	ID             string `json:"id"`
//	Type           string `json:"type" binding:"required"`
	Condition      string `json:"condition" binding:"required"` // an ES query string
//	Title          string `json:"title" binding:"required"`
//	Description    string `json:"description"`
//	UserID         string `json:"used_id" binding:"required"`
//	StartDate      string `json:"start_date" binding:"required"`
//	ExpirationDate string `json:"expiration_date"`
//	IsEnabled      bool   `json:"is_enabled" binding:"required"`
//	HitCount       int    `json:"hit_count" binding:"required"`
}

type EventDB struct {
	data map[string]Event
}

func newEventDB() *EventDB {
	db := new(EventDB)
	db.data = make(map[string]Event)
	return db
}

func (db *EventDB) write(event *Event) error {
	event.ID = newEventID()
	db.data[event.ID] = *event
	return nil
}
