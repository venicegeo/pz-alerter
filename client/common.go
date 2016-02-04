package client

import (
	"time"
)

type AlerterClient interface {
	// low-level interfaces
	PostToEvents(*Event) (*AlerterIdResponse, error)
	GetFromEvents() (*EventList, error)
	GetFromAlerts() (*AlertList, error)
	PostToConditions(*Condition) (*AlerterIdResponse, error)
	GetFromConditions() (*ConditionList, error)
	GetFromCondition(id string) (*Condition, error)
	DeleteOfCondition(id string) error

	GetFromAdminStats() (*AlerterAdminStats, error)
	GetFromAdminSettings() (*AlerterAdminSettings, error)
	PostToAdminSettings(*AlerterAdminSettings) error
}

type AlerterIdResponse struct {
	ID string `json:"id"`
}

/////////////////

const EventDataIngested = "DataIngested"
const EventDataAccessed = "DataAccessed"
const EventUSDataFound = "USDataFound"
const EventFoo = "Foo"
const EventBar = "Bar"

type EventType string

type Event struct {
	ID   string            `json:"id"`
	Type EventType         `json:"type" binding:"required"`
	Date string            `json:"date" binding:"required"`
	Data map[string]string `json:"data"` // specific to event type
}

type EventList map[string]Event

////////////////

type Alert struct {
	ID          string `json:"id"`
	ConditionID string `json:"condition_id" binding:"required"`
	EventID     string `json:"event_id" binding:"required"`
}

type AlertList map[string]Alert

//////////////

type Condition struct {
	ID string `json:"id"`

	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Type        EventType `json:"type" binding:"required"`
	UserID      string    `json:"user_id" binding:"required"`
	Date        string    `json:"start_date" binding:"required"`
	//ExpirationDate string `json:"expiration_date"`
	//IsEnabled      bool   `json:"is_enabled" binding:"required"`
	//HitCount int `json:"hit_count"`
}

type ConditionList map[string]Condition

/////////////////

type AlerterAdminStats struct {
	StartTime   time.Time `json:"starttime"`
	NumRequests int       `json:"num_requests"`
	NumUUIDs    int       `json:"num_uuids"`
}

type AlerterAdminSettings struct {
	Debug bool `json:"debug"`
}
