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

package logger

import (
	"errors"
	"fmt"
	"time"

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
)

const LogSchema = "LogData"
const SecuritySchema = "AuditData"

// Message represents the contents of a message for the logger service.
// All fields are required.
type Message struct {
	Service   piazza.ServiceName `json:"service"`
	Address   string             `json:"address"`
	CreatedOn time.Time          `json:"createdOn"`
	Severity  Severity           `json:"severity"`
	Message   string             `json:"message"`
}

//---------------------------------------------------------------------------

type Stats struct {
	CreatedOn time.Time `json:"createdOn"`

	// this is the number of messages since the service was started,
	// not the total number of messages in the system
	NumMessages float64 `json:"numMessages"`
}

// ToString returns a Message as a formatted string.
func (mssg *Message) String() string {
	t := mssg.CreatedOn.Format(time.RFC3339)
	s := fmt.Sprintf("[%s, %s, %s, %s, %s]",
		mssg.Service, mssg.Address, t, mssg.Severity, mssg.Message)
	return s
}

type Severity string

const (
	// SeverityDebug is for log messages that are only used in development.
	SeverityDebug Severity = "Debug"

	// SeverityInfo is for log messages that are only informative, no action needed.
	SeverityInfo Severity = "Info"

	// SeverityWarning is for log messages that indicate possible problems. Execution continues normally.
	SeverityWarning Severity = "Warning"

	// SeverityError is for log messages that indicate something went wrong. The problem is usually handled and execution continues.
	SeverityError Severity = "Error"

	// SeverityFatal is for log messages that indicate an internal error and the system is likely now unstable. These should never happen.
	SeverityFatal Severity = "Fatal"
)

// Validate checks to make sure a Message is properly filled out. If not, a non-nil error is returned.
func (mssg *Message) Validate() error {
	if mssg == nil {
		return errors.New("message is nil")
	}
	if mssg.Service == "" {
		return errors.New("required field 'service' not set")
	}
	if mssg.Address == "" {
		return errors.New("required field 'address' not set")
	}
	if mssg.CreatedOn.IsZero() {
		return errors.New("required field 'createdOn' not set")
	}
	if mssg.Severity == "" {
		return errors.New("required field 'severity' not set")
	}
	if mssg.Message == "" {
		return errors.New("required field 'message' not set")
	}

	return nil
}

//---------------------------------------------------------------------------

func init() {
	piazza.JsonResponseDataTypes["logger.Message"] = "logmessage"
	piazza.JsonResponseDataTypes["*logger.Message"] = "logmessage"
	piazza.JsonResponseDataTypes["[]logger.Message"] = "logmessage-list"
	piazza.JsonResponseDataTypes["[]syslog.Message"] = "syslogMessage-list"
	piazza.JsonResponseDataTypes["logger.Stats"] = "logstats"
	piazza.JsonResponseDataTypes["*logger.Stats"] = "logstats"
}

func paginationCreatedOnToTimeStamp(pagination *piazza.JsonPagination) {
	if pagination.SortBy == "createdOn" {
		pagination.SortBy = "timeStamp"
	}
}

var LogMapping = `
	{
	    "dynamic": "strict",
	    "properties": {
	    	"facility": {
        		"type": "integer"
      		},
      		"severity": {
        		"type": "integer"
      		},
      		"version": {
        		"type": "integer"
      		},
      		"timeStamp": {
        		"type": "string",
        		"index": "not_analyzed"
      		},
      		"hostName": {
        		"type": "string",
        		"index": "not_analyzed"
      		},
      		"application": {
        		"type": "string",
        		"index": "not_analyzed"
      		},
      		"process": {
        		"type": "string",
        		"index": "not_analyzed"
      		},
      		"messageId": {
        		"type": "string",
        		"index": "not_analyzed"
      		},
      		"auditData": {
        		"dynamic": "strict",
        		"properties": {
          			"actor": {
            			"type": "string",
            			"index": "not_analyzed"
          			},
          			"action": {
            			"type": "string",
            			"index": "not_analyzed"
          			},
          			"actee": {
            			"type": "string",
            			"index": "not_analyzed"
          			}
        		}
      		},
     		"metricData": {
        		"dynamic": "strict",
        		"properties": {
          			"name": {
            			"type": "string",
            			"index": "not_analyzed"
          			},
          			"value": {
            			"type": "double"
          			},
          			"object": {
            			"type": "string",
            			"index": "not_analyzed"
          			}
        		}
      		},
      		"sourceData": {
        		"dynamic": "strict",
        		"properties": {
          			"file": {
            			"type": "string",
            			"index": "not_analyzed"
          			},
          			"function": {
            			"type": "string",
            			"index": "not_analyzed"
          			},
          			"line": {
            			"type": "integer"
          			}
        		}
      		},
      		"message": {
        		"type": "string",
        		"index": "not_analyzed"
      		}
    	}
	}`
