package client

import (
)

//---------------------------------------------------------------------------
//---------------------------------------------------------------------------

var alertID = 1

// newAlert makes an Alert, setting the ID for you.
func NewAlert(actionID Ident) Alert {

	id := NewIdentFromInt(alertID)
	alertID++
	s := "A" + string(id)

	return Alert{
		ID:          Ident(s),
		Action: actionID,
	}
}
