package client

import (
	"sync"
)

//---------------------------------------------------------------------------
//---------------------------------------------------------------------------

var actionID = 1

var actionIdLock sync.Mutex

func NewActionIdent() Ident {
	actionIdLock.Lock()
	id := NewIdentFromInt(actionID)
	actionID++
	actionIdLock.Unlock()
	s := "X" + id.String()
	return Ident(s)
}

// newAlert makes an Alert, setting the ID for you.
func NewAction(conditions []Ident, events []Ident, job string) Action {

	id := NewActionIdent()

	return Action{
		ID:         id,
		Conditions: conditions,
		Events:     events,
		Job:        job,
	}
}
