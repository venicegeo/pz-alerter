package main

import (
	"strconv"
)

//---------------------------------------------------------------------------
//---------------------------------------------------------------------------

var alertID = 1

type Alert struct {
	ID          string `json:"id"`
	ConditionID string `json:"condition_id" binding:"required"`
	EventID     string `json:"event_id" binding:"required"`
}

// newAlert makes an Alert, setting the ID for you.
func newAlert(conditionID string, eventID string) Alert {

	id := strconv.Itoa(alertID)
	alertID++
	id = "A" + id

	return Alert{
		ID:          id,
		ConditionID: conditionID,
		EventID: eventID,
	}
}

func (item *Alert) id() string {
	return item.ID
}
