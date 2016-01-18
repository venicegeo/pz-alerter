package main

import (
	"strconv"
)

//---------------------------------------------------------------------------

var conditionID = 1


type Condition struct {
	ID string `json:"id"`

	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Type        EventType `json:"type" binding:"required"`
	UserID      string    `json:"user_id" binding:"required"`
	Date   string    `json:"start_date" binding:"required"`
	//ExpirationDate string `json:"expiration_date"`
	//IsEnabled      bool   `json:"is_enabled" binding:"required"`
	HitCount int `json:"hit_count"`
}

func newCondition() *Condition {
	id := strconv.Itoa(conditionID)
	conditionID++

	e := &Condition{ID: id}
	return e
}
