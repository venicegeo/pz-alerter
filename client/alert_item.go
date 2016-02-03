package client

import (
	"strconv"
)

//---------------------------------------------------------------------------
//---------------------------------------------------------------------------

var alertID = 1

// newAlert makes an Alert, setting the ID for you.
func NewAlert(conditionID string, eventID string) Alert {

	id := strconv.Itoa(alertID)
	alertID++
	id = "A" + id

	return Alert{
		ID:          id,
		ConditionID: conditionID,
		EventID: eventID,
	}
}
