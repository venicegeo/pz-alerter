package client

import (
	"sync"
)

//---------------------------------------------------------------------------

var conditionID = 1

var conditionIdLock sync.Mutex

func NewConditionIdent() Ident {
	conditionIdLock.Lock()
	id := NewIdentFromInt(conditionID)
	conditionID++
	conditionIdLock.Unlock()
	s := "C" + id.String()
	return Ident(s)
}
