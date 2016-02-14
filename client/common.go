package client

import (
	piazza "github.com/venicegeo/pz-gocommon"
	"sort"
	"strconv"
	"time"
)

type IAlerterService interface {
	GetName() piazza.ServiceName
	GetAddress() string

	// low-level interfaces
	PostToEvents(*Event) (*AlerterIdResponse, error)
	GetFromEvents() (*[]Event, error)
	DeleteOfEvent(id Ident) error

	GetFromAlerts() (*[]Alert, error)
	GetFromAlert(id Ident) (*Alert, error)
	PostToAlerts(*Alert) (*AlerterIdResponse, error)
	DeleteOfAlert(id Ident) error

	PostToTriggers(*Trigger) (*AlerterIdResponse, error)
	GetFromTriggers() (*[]Trigger, error)
	GetFromTrigger(id Ident) (*Trigger, error)
	DeleteOfTrigger(id Ident) error

	GetFromAdminStats() (*AlerterAdminStats, error)
	GetFromAdminSettings() (*AlerterAdminSettings, error)
	PostToAdminSettings(*AlerterAdminSettings) error
}

type AlerterIdResponse struct {
	ID Ident `json:"id"`
}

type Ident string

const NoIdent Ident = ""

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
	EventBaz          EventType = "Baz"
	EventBuz          EventType = "Buz"
)

/////////////////

// expresses the idea of "this ES query returns an event"
// Query is specific to the event type
type Condition struct {
	Type  EventType `json:"type" binding:"required"`
	Query string    `json:"query" binding:"required"`
}

type Job struct {
	Task string
}

/////////////////

// when the and'ed set of Conditions all are true, do Something
// Events are the results of the Conditions queries
// Job is the JobMessage to submit back to Pz
// TODO: some sort of mapping from the event info into the Job string
type Trigger struct {
	ID        Ident     `json:"id"`
	Title     string    `json:"title" binding:"required"`
	Condition Condition `json:"conditions" binding:"required"`
	Job       Job       `json:"job" binding:"required"`
}

type TriggerList []Trigger

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

type EventList []Event

////////////////

// a notification, automatically created when an Trigger happens
type Alert struct {
	ID        Ident `json:"id"`
	TriggerId Ident `json:"trigger_id"`
	EventId   Ident `json:"event_id"`
}

type AlertList []Alert

type AlertListById []Alert

func (a AlertListById) Len() int           { return len(a) }
func (a AlertListById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AlertListById) Less(i, j int) bool { return a[i].ID < a[j].ID }

func (list AlertList) ToSortedArray() []Alert {
	array := make([]Alert, len(list))
	i := 0
	for _, v := range list {
		array[i] = v
		i++
	}
	sort.Sort(AlertListById(array))
	return array
}

//////////////

type AlerterAdminStats struct {
	Date          time.Time `json:"date"`
	NumAlerts     int       `json:"num_alerts"`
	NumConditions int       `json:"num_conditions"`
	NumEvents     int       `json:"num_events"`
	NumTriggers   int       `json:"num_triggers"`
}

type AlerterAdminSettings struct {
	Debug bool `json:"debug"`
}
