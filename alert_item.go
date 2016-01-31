package main

import (
	"strconv"
	piazza "github.com/venicegeo/pz-gocommon"
)

//---------------------------------------------------------------------------
//---------------------------------------------------------------------------

var alertID = 1

// newAlert makes an Alert, setting the ID for you.
func newAlert(conditionID string, eventID string) piazza.Alert {

	id := strconv.Itoa(alertID)
	alertID++
	id = "A" + id

	return piazza.Alert{
		ID:          id,
		ConditionID: conditionID,
		EventID: eventID,
	}
}
