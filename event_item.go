package main

import (
	"strconv"
)

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
