package client

import (
	"strconv"
)

//---------------------------------------------------------------------------

var conditionID = 1


func NewConditionID() string {
	id := strconv.Itoa(conditionID)
	conditionID++
	return id
}
