// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
)

type WorkflowIDResponse struct {
	ID Ident `json:"id"`
}

func SuperConvert(src interface{}, dst interface{}) error {
	jsn, err := json.Marshal(src)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsn, dst)
	if err != nil {
		return err
	}

	return nil
}

type Ident string

const NoIdent Ident = ""

func (id Ident) String() string {
	return string(id)
}

//---------------------------------------------------------------------------

// expresses the idea of "this ES query returns an event"
// Query is specific to the event type
type Condition struct {
	EventTypeIDs []Ident                `json:"eventtype_ids" binding:"required"`
	Query        map[string]interface{} `json:"query" binding:"required"`
}

//---------------------------------------------------------------------------

// Job JSON struct
type Job struct {
	Username string                 `json:"userName" binding:"required"`
	JobType  map[string]interface{} `json:"jobType" binding:"required"`
}

//---------------------------------------------------------------------------

// when the and'ed set of Conditions all are true, do Something
// Events are the results of the Conditions queries
// Job is the JobMessage to submit back to Pz
type Trigger struct {
	ID            Ident     `json:"id"`
	Title         string    `json:"title" binding:"required"`
	Condition     Condition `json:"condition" binding:"required"`
	Job           Job       `json:"job" binding:"required"`
	PercolationID Ident     `json:"percolation_id"`
}

type TriggerList []Trigger

//---------------------------------------------------------------------------

// posted by some source (service, user, etc) to indicate Something Happened
// Data is specific to the event type
// TODO: use the delayed-parsing, raw-message json thing?
type Event struct {
	ID          Ident                  `json:"id"`
	EventTypeID Ident                  `json:"eventtype_id" binding:"required"`
	Date        time.Time              `json:"date" binding:"required"`
	Data        map[string]interface{} `json:"data"`
}

type EventList []Event

//---------------------------------------------------------------------------

type EventType struct {
	ID      Ident                                           `json:"id"`
	Name    string                                          `json:"name" binding:"required"`
	Mapping map[string]elasticsearch.MappingElementTypeName `json:"mapping" binding:"required"`
}

type EventTypeList []EventType

//---------------------------------------------------------------------------

// a notification, automatically created when an Trigger happens
type Alert struct {
	ID        Ident `json:"id"`
	TriggerID Ident `json:"trigger_id"`
	EventID   Ident `json:"event_id"`
	JobID     Ident `json:"job_id"`
}

type AlertList []Alert

type AlertListByID []Alert

func (a AlertListByID) Len() int {
	return len(a)
}
func (a AlertListByID) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a AlertListByID) Less(i, j int) bool {
	return a[i].ID < a[j].ID
}

func (list AlertList) ToSortedArray() []Alert {
	array := make([]Alert, len(list))
	i := 0
	for _, v := range list {
		array[i] = v
		i++
	}
	sort.Sort(AlertListByID(array))
	return array
}

//---------------------------------------------------------------------------

type WorkflowAdminStats struct {
	Date          time.Time `json:"date"`
	NumAlerts     int       `json:"num_alerts"`
	NumConditions int       `json:"num_conditions"`
	NumEvents     int       `json:"num_events"`
	NumTriggers   int       `json:"num_triggers"`
}

type WorkflowAdminSettings struct {
	Debug bool `json:"debug"`
}

func LoggedError(mssg string, args ...interface{}) error {
	str := fmt.Sprintf(mssg, args)
	log.Printf(str)
	return errors.New(str)
}
