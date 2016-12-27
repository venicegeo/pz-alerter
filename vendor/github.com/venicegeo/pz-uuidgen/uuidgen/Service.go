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

package uuidgen

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	pzsyslog "github.com/venicegeo/pz-gocommon/syslog"
)

//---------------------------------------------------------------------

type Service struct {
	sync.Mutex
	stats     Stats
	syslogger *pzsyslog.Logger
	origin    string
}

//---------------------------------------------------------------------

func (service *Service) Init(sys *piazza.SystemConfig, logger *pzsyslog.Logger) error {
	service.stats.CreatedOn = time.Now()

	service.origin = string(sys.Name)

	service.syslogger = logger

	service.syslogger.Info("uuidgen service started")

	return nil
}

func (service *Service) GetStats() *piazza.JsonResponse {
	log.Printf("uuidgen stats service called (1)")
	service.syslogger.Info("uuidgen stats service called (2)")

	service.Lock()
	data := service.stats
	service.Unlock()

	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: data}
	err := resp.SetType()
	if err != nil {
		return &piazza.JsonResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
			Origin:     service.origin,
		}
	}

	return resp
}

// PostUuids generates one or more UUIDs.
//
// The request body is ignored. We allow a count of zero, for testing.
func (service *Service) PostUuids(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	var count int
	var err error

	// ?count=INT
	count, err = params.GetCount(1)
	if err != nil {
		return &piazza.JsonResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
			Origin:     service.origin,
		}
	}

	if count < 0 || count > 255 {
		s := fmt.Sprintf("query argument out of range: %d", count)
		return &piazza.JsonResponse{
			StatusCode: http.StatusBadRequest,
			Message:    s,
			Origin:     service.origin,
		}
	}

	uuids := make([]string, count)
	for i := 0; i < count; i++ {
		uuids[i] = piazza.NewUuid().String()
	}
	// service.syslogger.Audit("pz-uuidgen", "createUUID", "", "UUIDGen created uuids: [%s]", uuids)

	service.Lock()
	service.stats.NumUUIDs += count
	service.stats.NumRequests++
	service.Unlock()

	resp := &piazza.JsonResponse{StatusCode: http.StatusCreated, Data: uuids}
	err = resp.SetType()
	if err != nil {
		return &piazza.JsonResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
			Origin:     service.origin,
		}
	}

	return resp
}
