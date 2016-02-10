package client

import (
	piazza "github.com/venicegeo/pz-gocommon"
	"time"
	"strconv"
)

type IAlerterService interface {
	GetName() piazza.ServiceName
	GetAddress() string

	// low-level interfaces
	PostToEvents(*Event) (*AlerterIdResponse, error)
	GetFromEvents() (*EventList, error)
	GetFromAlerts() (*AlertList, error)
	PostToConditions(*Condition) (*AlerterIdResponse, error)
	GetFromConditions() (*ConditionList, error)
	GetFromCondition(id Ident) (*Condition, error)
	DeleteOfCondition(id Ident) error
	PostToActions(*Action) (*AlerterIdResponse, error)
	GetFromActions() (*ActionList, error)
	GetFromAction(id Ident) (*Action, error)

	GetFromAdminStats() (*AlerterAdminStats, error)
	GetFromAdminSettings() (*AlerterAdminSettings, error)
	PostToAdminSettings(*AlerterAdminSettings) error
}

type AlerterIdResponse struct {
	ID Ident `json:"id"`
}

type Ident string

func (id Ident) String() string {
	return string(id)
}

func NewIdentFromInt(id int) Ident {
	s := strconv.Itoa(id)
	return Ident(s)
}

/////////////////

type EventType string

const (
	EventDataIngested EventType = "DataIngested"
	EventDataAccessed EventType = "DataAccessed"
	EventUSDataFound  EventType = "USDataFound"
	EventFoo          EventType = "Foo"
	EventBar          EventType = "Bar"
)

/////////////////

// expresses the idea of "this ES query returns an event"
// Query is specific to the event type
type Condition struct {
	ID    Ident     `json:"id"`
	Title string    `json:"title" binding:"required"`
	Type  EventType `json:"type" binding:"required"`
	Query string    `json:"query" binding:"required"`
}

type ConditionList map[Ident]Condition

/////////////////

// when the and'ed set of Conditions all are true, do Something
// Events are the results of the Conditions queries
// Job is the JobMessage to submit back to Pz
// TODO: some sort of mapping from the event info into the Job string
type Action struct {
	ID         Ident   `json:"id"`
	Conditions []Ident `json:"conditions" binding:"required"`
	Events     []Ident `json:"events"`
	Job        string  `json:job`
}

type ActionList map[Ident]Action

/////////////////

// posted by some source (service, user, etc) to indicate Something Happened
// Data is specific to the event type
// TODO: use the delayed-parsing, raw-message json thing?
type Event struct {
	ID   Ident             `json:"id"`
	Type EventType         `json:"type" binding:"required"`
	Date time.Time         `json:"date" binding:"required"`
	Data map[string]string `json:"data"`
}

type EventList map[Ident]Event

////////////////

// a notification, automatically created when an Action happens
type Alert struct {
	ID     Ident `json:"id"`
	Action Ident `json:"action_id"`
}

type AlertList map[Ident]Alert

//////////////

type AlerterAdminStats struct {
	Date          time.Time `json:"date"`
	NumAlerts     int       `json:"num_alerts"`
	NumConditions int       `json:"num_conditions"`
	NumEvents     int       `json:"num_events"`
	NumActions    int       `json:"num_actions"`
}

type AlerterAdminSettings struct {
	Debug bool `json:"debug"`
}
