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
	"time"

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
)

const LogSchema = "LogData"
const SecuritySchema = "AuditData"

//---------------------------------------------------------------------------

type Stats struct {
	CreatedOn time.Time `json:"createdOn"`

	// this is the number of messages since the service was started,
	// not the total number of messages in the system
	NumMessages float64 `json:"numMessages"`
}

//---------------------------------------------------------------------------

func init() {
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
