package main

import (
	"strconv"
)

//---------------------------------------------------------------------------
//---------------------------------------------------------------------------

var alertID = 1

func newAlertID() string {
	s := strconv.Itoa(alertID)
	alertID++
	return s
}

type Alert struct {
	ID          string `json:"id"`
	ConditionID string `json:"condition_id" binding:"required"`
	EventID     string `json:"event_id" binding:"required"`
}

// newAlert makes an Alert, setting the ID for you.
func newAlert(conditionID string, eventID string) Alert {
	return Alert{
		ID:          newAlertID(),
		ConditionID: conditionID,
		EventID: eventID,
	}
}

func (item *Alert) id() string {
	return item.ID
}
