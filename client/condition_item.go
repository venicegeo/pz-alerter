package client

import (
)

//---------------------------------------------------------------------------

var conditionID = 1

func NewConditionID() Ident {
	id := NewIdentFromInt(conditionID)
	conditionID++
	return id
}
