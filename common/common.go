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

package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/venicegeo/pz-gocommon"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
)

type WorkflowIdResponse struct {
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


type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func NewErrorResponseFromHttp(resp *http.Response) *ErrorResponse {
	defer resp.Body.Close()
	mssgBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ErrorResponse{Status: 500, Message: "failed to parse error response: " + err.Error()}
	}

	var e ErrorResponse
	err = json.Unmarshal(mssgBytes, &e)
	if err != nil {
		return &ErrorResponse{Status: 500, Message: "failed to parse error response: " + err.Error()}
	}

	return &e
}

func NewErrorFromHttp(resp *http.Response) error {
	errResp := NewErrorResponseFromHttp(resp)
	return errors.New(errResp.String())
}

func (e *ErrorResponse) String() string {
	return fmt.Sprintf("[error response: %d: %s]", e.Status, e.Message)
}

func (e *ErrorResponse) Error() error {
	return errors.New(e.String())
}

type Ident string

const NoIdent Ident = ""

var globalIdLock sync.Mutex
var globalID = 1
var debugIds = true

func NewIdent() Ident {
	if debugIds {
		globalIdLock.Lock()
		s := strconv.Itoa(globalID)
		globalID++
		globalIdLock.Unlock()
		id := "W" + s
		return Ident(id)
	} else {
		panic(12345)
	}
}

func (id Ident) String() string {
	return string(id)
}

type IIdentable interface {
	GetId() Ident
}

//---------------------------------------------------------------------------

// expresses the idea of "this ES query returns an event"
// Query is specific to the event type
type Condition struct {
	EventType Ident  `json:"type" binding:"required"`
	Query     map[string]interface{} `json:"query" binding:"required"`
}

type Job struct {
	Task string
}

//---------------------------------------------------------------------------

// when the and'ed set of Conditions all are true, do Something
// Events are the results of the Conditions queries
// Job is the JobMessage to submit back to Pz
// TODO: some sort of mapping from the event info into the Job string
type Trigger struct {
	ID        Ident     `json:"id"`
	Title     string    `json:"title" binding:"required"`
	Condition Condition `json:"condition" binding:"required"`
	Job       Job       `json:"job" binding:"required"`
	PercolationID Ident `json:"percolation_id"`
}

type TriggerList []Trigger

//---------------------------------------------------------------------------

// posted by some source (service, user, etc) to indicate Something Happened
// Data is specific to the event type
// TODO: use the delayed-parsing, raw-message json thing?
type Event struct {
	ID        Ident                  `json:"id"`
	EventType Ident                  `json:"type" binding:"required"`
	Date      time.Time              `json:"date" binding:"required"`
	Data      map[string]interface{} `json:"data"`
}

type EventList []Event

//---------------------------------------------------------------------------

type EventType struct {
	ID      Ident                                    `json:"id"`
	Name    string                                   `json:"name" binding:"required"`
	Mapping map[string]piazza.MappingElementTypeName `json:"mapping" binding:"required"`
}

type EventTypeList []EventType

//---------------------------------------------------------------------------

// a notification, automatically created when an Trigger happens
type Alert struct {
	ID        Ident `json:"id"`
	TriggerId Ident `json:"trigger_id"`
	EventId   Ident `json:"event_id"`
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
	return a[i].ID < a[j].ID
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
	Date          time.Time `json:"date"`
	NumAlerts     int       `json:"num_alerts"`
	NumConditions int       `json:"num_conditions"`
	NumEvents     int       `json:"num_events"`
	NumTriggers   int       `json:"num_triggers"`
}

type WorkflowAdminSettings struct {
	Debug bool `json:"debug"`
}
