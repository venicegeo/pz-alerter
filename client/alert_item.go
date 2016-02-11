package client

import (
	"sync"
)

var alertIdLock sync.Mutex
var alertID = 1

func NewAlertIdent() Ident {
	alertIdLock.Lock()
	id := NewIdentFromInt(alertID)
	alertID++
	alertIdLock.Unlock()
	s := "A" + id.String()
	return Ident(s)
}

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
