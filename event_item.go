package main

import (
	"strconv"
)

//---------------------------------------------------------------------------

var eventID = 1

func newEventID() string {
	id := strconv.Itoa(eventID)
	eventID++
	return "E" + id
}
