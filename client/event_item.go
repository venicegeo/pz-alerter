package client

import (
)

//---------------------------------------------------------------------------

var eventID = 1

func NewEventID() Ident {
	id := NewIdentFromInt(eventID)
	eventID++
	return Ident("E" + string(id))
}
