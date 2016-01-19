package main

import (
	"strconv"
)

//---------------------------------------------------------------------------

var eventID = 1

const EventDataIngested = "DataIngested"
const EventDataAccessed = "DataAccessed"
const EventUSDataFound = "USDataFound"
const EventFoo = "Foo"
const EventBar = "Bar"

type EventType string

type Event struct {
	ID   string            `json:"id" binding:"required"`
	Type EventType         `json:"type" binding:"required"`
	Date string            `json:"date" binding:"required"`
	Data map[string]string `json:"data"` // specific to event type
}

func newEvent() *Event {
	id := strconv.Itoa(eventID)
	eventID++

	e := &Event{ID: id}
	return e
}
