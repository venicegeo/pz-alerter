package client

import (
	"strconv"
)

//---------------------------------------------------------------------------

var eventID = 1

func NewEventID() string {
	id := strconv.Itoa(eventID)
	eventID++
	return "E" + id
}
