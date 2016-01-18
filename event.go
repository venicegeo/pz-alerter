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

// Event types
const EventDataIngested = "DataIngested"
const EventDataAccessed = "DataAccessed"
const EventUSDataFound = "USDataFound"
const EventFoo = "Foo"
const EventBar = "Bar"

type EventType string

type Event struct {
	ID   string            `json:"id"`
	Type EventType         `json:"type" binding:"required"`
	Date string            `json:"date" binding:"required"`
	Data map[string]string `json:"data"` // specific to event type
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
