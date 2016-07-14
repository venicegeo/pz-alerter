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
	"time"

	uuidpkg "github.com/pborman/uuid"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

const AlertDBMapping string = "Alert"
const TriggerDBMapping string = "Trigger"
const EventTypeDBMapping string = "EventType"

//---------------------------------------------------------------------------

const (
	TriggerIndexSettings = `
{
	"mappings": {
		"Trigger": {
			"properties": {
				"triggerId": {
					"type": "string",
					"index": "not_analyzed"
				},
				"title": {
					"type": "string",
					"index": "not_analyzed"
				},
				"createdOn": {
					"type": "date",
					"index": "not_analyzed"
				},
				"createdBy": {
					"type": "string",
					"index": "not_analyzed"
				},
				"disabled": {
					"type": "boolean",
					"index": "not_analyzed"
				},
				"condition": {
					"properties": {
						"eventTypeIds": {
							"type": "string",
							"index": "not_analyzed"
						},
						"query": {
							"dynamic": true,
							"properties": {}
						}
					}
				},
				"job": {
					"properties": {
						"createdBy": {
							"type": "string",
							"index": "not_analyzed"
						},
						"jobType": {
							"dynamic": true,
							"properties": {}
						}
					}
				},
				"percolationId": {
					"type": "string",
					"index": "not_analyzed"
				}
			}
		}
	}
}
`
)

// expresses the idea of "this ES query returns an event"
// Query is specific to the event type
type Condition struct {
	EventTypeIds []piazza.Ident         `json:"eventTypeIds" binding:"required"`
	Query        map[string]interface{} `json:"query" binding:"required"`
}

// Job JSON struct
type Job struct {
	CreatedBy string                 `json:"createdBy" binding:"required"`
	JobType   map[string]interface{} `json:"jobType" binding:"required"`
}

// when the and'ed set of Conditions all are true, do Something
// Events are the results of the Conditions queries
// Job is the JobMessage to submit back to Pz
type Trigger struct {
	TriggerId     piazza.Ident `json:"triggerId"`
	Title         string       `json:"title" binding:"required"`
	Condition     Condition    `json:"condition" binding:"required"`
	Job           Job          `json:"job" binding:"required"`
	PercolationId piazza.Ident `json:"percolationId"`
	CreatedBy     string       `json:"createdBy"`
	CreatedOn     time.Time    `json:"createdOn"`
	Disabled      bool         `json:"disabled"`
}

type TriggerList []Trigger

//---------------------------------------------------------------------------

const (
	EventIndexSettings = `
{
	"mappings": {
		"_default_": {
			"properties": {
				"eventTypeId": {
					"type": "string",
					"index": "not_analyzed"
				},
				"eventId": {
					"type": "string",
					"index": "not_analyzed"
				},
				"data": {
					"dynamic": true,
					"properties": {}
				},
				"createdBy": {
					"type": "string",
					"index": "not_analyzed"
				},
				"createdOn": {
					"type": "date",
					"index": "not_analyzed"
				}
			}
		}
	}
}
`
)

// posted by some source (service, user, etc) to indicate Something Happened
// Data is specific to the event type
type Event struct {
	EventId     piazza.Ident           `json:"eventId"`
	EventTypeId piazza.Ident           `json:"eventTypeId" binding:"required"`
	Data        map[string]interface{} `json:"data"`
	CreatedBy   string                 `json:"createdBy"`
	CreatedOn   time.Time              `json:"createdOn"`
}

type EventList []Event

//---------------------------------------------------------------------------

const (
	EventTypeIndexSettings = `
{
	"mappings": {
		"EventType": {
			"properties": {
				"eventTypeId": {
					"type": "string",
					"index": "not_analyzed"
				},
				"name": {
					"type": "string",
					"index": "not_analyzed"
				},
				"createdOn": {
					"type": "date",
					"index": "not_analyzed"
				},
				"createdBy": {
					"type": "string",
					"index": "not_analyzed"
				},
				"mapping": {
					"dynamic": true,
					"properties": {}
				}
			}
		}
	}
}
`
)

type EventType struct {
	EventTypeId piazza.Ident                                    `json:"eventTypeId"`
	Name        string                                          `json:"name" binding:"required"`
	Mapping     map[string]elasticsearch.MappingElementTypeName `json:"mapping" binding:"required"`
	CreatedBy   string                                          `json:"createdBy"`
	CreatedOn   time.Time                                       `json:"createdOn"`
}

type EventTypeList []EventType

//---------------------------------------------------------------------------

// The default settings for our Elasticsearch alerts index
// Explanation:
//   "index": "not_analyzed"
//     This means that these properties are not analyzed by Elasticsearch.
//     Previously, these ids were analyzed by ES and thus broken up into chunks;
//     in the case of a UUID this would happen via break-up by the "-" character.
//     For example, the UUID "ab3142cd-1a8e-44f8-6a01-5ce8a9328fb2" would be broken
//     into "ab3142cd", "1a8e", "44f8", "6a01" and "5ce8a9328fb2", and queries would
//     match on all of these separate strings, which was undesired behavior.
const (
	AlertIndexSettings = `
{
	"mappings": {
		"Alert": {
			"properties": {
				"alertId": {
					"type": "string",
					"index": "not_analyzed"
				},
				"triggerId": {
					"type": "string",
					"index": "not_analyzed"
				},
				"jobId": {
					"type": "string",
					"index": "not_analyzed"
				},
				"eventId": {
					"type": "string",
					"index": "not_analyzed"
				},
				"createdBy": {
					"type": "string",
					"index": "not_analyzed"
				},
				"createdOn": {
					"type": "date",
					"index": "not_analyzed"
				}
			}
		}
	}
}
`
)

// a notification, automatically created when a Trigger happens
type Alert struct {
	AlertId   piazza.Ident `json:"alertId"`
	TriggerId piazza.Ident `json:"triggerId"`
	EventId   piazza.Ident `json:"eventId"`
	JobId     piazza.Ident `json:"jobId"`
	CreatedBy string       `json:"createdBy"`
	CreatedOn time.Time    `json:"createdOn"`
}

//---------------------------------------------------------------------------

type WorkflowAdminStats struct {
	CreatedOn     time.Time `json:"createdOn"`
	NumAlerts     int       `json:"numAlerts"`
	NumConditions int       `json:"numConditions"`
	NumEvents     int       `json:"numEvents"`
	NumTriggers   int       `json:"numTriggers"`
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

//---------------------------------------------------------------------------

func init() {
	piazza.JsonResponseDataTypes["*workflow.EventType"] = "eventtype"
	piazza.JsonResponseDataTypes["*[]workflow.EventType"] = "eventtype-list"
	piazza.JsonResponseDataTypes["*workflow.Event"] = "event"
	piazza.JsonResponseDataTypes["*[]workflow.Event"] = "event-list"
	piazza.JsonResponseDataTypes["*workflow.Trigger"] = "trigger"
	piazza.JsonResponseDataTypes["*[]workflow.Trigger"] = "trigger-list"
	piazza.JsonResponseDataTypes["*workflow.Alert"] = "alert"
	piazza.JsonResponseDataTypes["*[]workflow.Alert"] = "alert-list"
	piazza.JsonResponseDataTypes["workflow.WorkflowAdminStats"] = "workflowstats"
}
