package main

import (
	"strconv"
)

//---------------------------------------------------------------------------

var conditionID = 1


func newConditionID() string {
	id := strconv.Itoa(conditionID)
	conditionID++
	return id
}
