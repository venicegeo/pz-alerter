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

package workflow

import (
	"errors"
	"fmt"
	"sort"
	"time"

	uuidpkg "github.com/pborman/uuid"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

//---------------------------------------------------------------------------

// expresses the idea of "this ES query returns an event"
// Query is specific to the event type
type Condition struct {
	EventTypeIds []piazza.Ident         `json:"eventTypeIds" binding:"required"`
	Query        map[string]interface{} `json:"query" binding:"required"`
}

//---------------------------------------------------------------------------

// Job JSON struct
type Job struct {
	UserName string                 `json:"userName" binding:"required"`
	JobType  map[string]interface{} `json:"jobType" binding:"required"`
}

//---------------------------------------------------------------------------

// when the and'ed set of Conditions all are true, do Something
// Events are the results of the Conditions queries
// Job is the JobMessage to submit back to Pz
type Trigger struct {
	TriggerId     piazza.Ident `json:"triggerId"`
	Title         string       `json:"title" binding:"required"`
	Condition     Condition    `json:"condition" binding:"required"`
	Job           Job          `json:"job" binding:"required"`
	PercolationId piazza.Ident `json:"percolationId"`
	UserName      string       `json:"userName"`
	CreatedOn     time.Time    `json:"createdOn"`
}

type TriggerList []Trigger

//---------------------------------------------------------------------------

// posted by some source (service, user, etc) to indicate Something Happened
// Data is specific to the event type
type Event struct {
	EventId     piazza.Ident           `json:"eventId"`
	EventTypeId piazza.Ident           `json:"eventTypeId" binding:"required"`
	CreatedOn   time.Time              `json:"createdOn"`
	Data        map[string]interface{} `json:"data"`
	UserName    string                 `json:"userName"`
}

type EventList []Event

//---------------------------------------------------------------------------

type EventType struct {
	EventTypeId piazza.Ident                                    `json:"eventTypeId"`
	Name        string                                          `json:"name" binding:"required"`
	Mapping     map[string]elasticsearch.MappingElementTypeName `json:"mapping" binding:"required"`
	UserName    string                                          `json:"userName"`
	CreatedOn   time.Time                                       `json:"createdOn"`
}

type EventTypeList []EventType

//---------------------------------------------------------------------------

// a notification, automatically created when an Trigger happens
type Alert struct {
	AlertId   piazza.Ident `json:"alertId"`
	TriggerId piazza.Ident `json:"triggerId"`
	EventId   piazza.Ident `json:"eventId"`
	JobId     piazza.Ident `json:"jobId"`
	CreatedOn time.Time    `json:"createdOn"`
	UserName  string       `json:"userName"`
}

type AlertList []Alert

type AlertListById []Alert

func (a AlertListById) Len() int {
	return len(a)
}
func (a AlertListById) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a AlertListById) Less(i, j int) bool {
	return a[i].AlertId < a[j].AlertId
}

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

//---------------------------------------------------------------------------

type WorkflowAdminStats struct {
	CreatedOn     time.Time `json:"createdOn"`
	NumAlerts     int       `json:"numAlerts"`
	NumConditions int       `json:"numConditions"`
	NumEvents     int       `json:"numEvents"`
	NumTriggers   int       `json:"numTriggers"`
}

type WorkflowAdminSettings struct {
	Debug bool `json:"debug"`
}

func LoggedError(mssg string, args ...interface{}) error {
	str := fmt.Sprintf(mssg, args)
	//log.Printf(str)
	return errors.New(str)
}

// Checks to see if the Uuid is valid
func isUuid(uuid string) bool {
	check := uuidpkg.Parse(uuid)
	return check != nil
}
