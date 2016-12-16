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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	syslogger "github.com/venicegeo/pz-gocommon/syslog"
)

type Service struct {
	sync.Mutex

	stats  Stats
	origin string

	logWriters   []syslogger.Writer
	auditWriters []syslogger.Writer

	esIndex elasticsearch.IIndex
}

func (service *Service) Init(sys *piazza.SystemConfig, logWriters []syslogger.Writer, auditWriters []syslogger.Writer, esi elasticsearch.IIndex) error {

	service.stats.CreatedOn = time.Now()

	service.logWriters = logWriters
	service.auditWriters = auditWriters

	service.esIndex = esi

	service.origin = string(sys.Name)

	return nil
}

func (service *Service) newInternalErrorResponse(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusInternalServerError,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

func (service *Service) newBadRequestResponse(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusBadRequest,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

func (service *Service) GetRoot() *piazza.JsonResponse {
	resp := &piazza.JsonResponse{
		StatusCode: 200,
		Data:       "Hi. I'm pz-logger.",
	}

	err := resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	return resp
}

func (service *Service) GetStats() *piazza.JsonResponse {
	service.Lock()
	t := service.stats
	service.Unlock()

	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       t,
	}

	err := resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	return resp
}

func createQueryDslAsString(
	pagination *piazza.JsonPagination,
	params *piazza.HttpQueryParams) (string, error) {

	must := []map[string]interface{}{}

	service, err := params.GetAsString("service", "")
	if err != nil {
		return "", err
	}

	contains, err := params.GetAsString("contains", "")
	if err != nil {
		return "", err
	}

	before, err := params.GetBefore(time.Time{})
	if err != nil {
		return "", err
	}

	after, err := params.GetAfter(time.Time{})
	if err != nil {
		return "", err
	}

	if service != "" {
		must = append(must, map[string]interface{}{
			"match": map[string]interface{}{
				"service": service,
			},
		})
	}
	if contains != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  contains,
				"fields": []string{"address", "message", "service", "severity"},
			},
		})
	}

	if !after.IsZero() || !before.IsZero() {
		rangeParams := map[string]time.Time{}

		if !after.IsZero() {
			rangeParams["gte"] = after
		}

		if !before.IsZero() {
			rangeParams["lte"] = before
		}

		must = append(must, map[string]interface{}{
			"range": map[string]interface{}{
				"timeStamp": rangeParams,
			},
		})
	}

	if len(must) == 0 {
		return "", nil
	}

	dsl := map[string]interface{}{
		"query": map[string]interface{}{
			"filtered": map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": must,
					},
				},
			},
		},
		"size": pagination.PerPage,
		"from": pagination.PerPage * pagination.Page,
	}

	dsl["sort"] = map[string]string{
		pagination.SortBy: string(pagination.Order),
	}

	output, err := json.Marshal(dsl)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (service *Service) PostSyslog(mNew *syslogger.Message) *piazza.JsonResponse {
	err := mNew.Validate()
	if err != nil {
		return service.newBadRequestResponse(err)
	}

	err = service.postSyslog(mNew)
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
	}
	return resp
}

func (service *Service) toOldStyle(mNew *syslogger.Message) (*Message, error) {
	var severity Severity

	switch mNew.Severity {
	case syslogger.Debug:
		severity = SeverityDebug
	case syslogger.Informational:
		severity = SeverityInfo
	case syslogger.Warning:
		severity = SeverityWarning
	case syslogger.Error:
		severity = SeverityError
	case syslogger.Fatal:
		severity = SeverityFatal
	default:
		severity = SeverityError
	}

	text := mNew.String()
	application := piazza.ServiceName(mNew.Application)

	mssgOld := &Message{
		CreatedOn: time.Now(),
		Service:   application,
		Address:   mNew.HostName,
		Severity:  severity,
		Message:   text,
	}
	if err := mssgOld.Validate(); err != nil {
		return mssgOld, err
	}
	return mssgOld, nil
}

func (service *Service) postSyslog(mssg *syslogger.Message) error {

	var err error
	isAudit := mssg.AuditData != nil

	for _, writer := range service.logWriters {
		err = writer.Write(mssg)
		if err != nil {
			log.Printf("message post: %s", err.Error())
			if !isAudit {
				return errors.New(fmt.Sprintf("Service.postSyslog: %s", err.Error()))
			}
		}
	}

	if isAudit {
		for _, writer := range service.auditWriters {
			err = writer.Write(mssg)
			if err != nil {
				log.Printf("message audit post: %s", err.Error())
				return errors.New(fmt.Sprintf("Service.postSyslog: %s", err.Error()))
			}
		}
	}

	service.Lock()
	service.stats.NumMessages++

	service.Unlock()

	return nil
}

func (service *Service) getMessageCommon(params *piazza.HttpQueryParams) (*elasticsearch.SearchResult, *piazza.JsonPagination, *piazza.JsonResponse) {
	pagination, err := piazza.NewJsonPagination(params)
	if err != nil {
		return nil, nil, service.newBadRequestResponse(err)
	}

	paginationCreatedOnToTimeStamp(pagination)

	dsl, err := createQueryDslAsString(pagination, params)
	if err != nil {
		return nil, pagination, service.newBadRequestResponse(err)
	}

	var searchResult *elasticsearch.SearchResult

	if dsl == "" {
		searchResult, err = service.esIndex.FilterByMatchAll(LogSchema, pagination)
	} else {
		searchResult, err = service.esIndex.SearchByJSON(LogSchema, dsl)
	}
	if err != nil {
		return nil, pagination, service.newInternalErrorResponse(err)
	}

	return searchResult, pagination, nil
}

func (service *Service) GetSyslog(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	var err error

	format, err := params.GetAsString("format", "json")
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	searchResult, pagination, jErr := service.getMessageCommon(params)
	if jErr != nil {
		return jErr
	}

	var lines = make([]syslogger.Message, 0)

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			if hit.Source == nil {
				log.Printf("null source hit")
				continue
			}

			var msg syslogger.Message
			err = json.Unmarshal(*hit.Source, &msg)
			if err != nil {
				log.Printf("UNABLE TO PARSE: %s", string(*hit.Source))
				return service.newInternalErrorResponse(err)
			}

			// just in case
			err = msg.Validate()
			if err != nil {
				log.Printf("UNABLE TO VALIDATE: %s", string(*hit.Source))
				continue
			}

			lines = append(lines, msg)
		}
	}

	var data interface{}

	switch format {
	case "string":
		temp := []string{}
		for _, line := range lines {
			temp = append(temp, line.String())
		}
		data = temp
	default:
		data = lines
	}

	pagination.Count = int(searchResult.TotalHits())
	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       data,
		Pagination: pagination,
	}

	err = resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	return resp
}

func (service *Service) GetMessage(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	var err error

	searchResult, pagination, jErr := service.getMessageCommon(params)
	if jErr != nil {
		return jErr
	}
	paginationTimeStampToCreateOn(pagination)

	var lines = make([]Message, 0)

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			if hit.Source == nil {
				log.Printf("null source hit")
				continue
			}

			var sysMsg syslogger.Message
			msg := &Message{}
			err = json.Unmarshal(*hit.Source, &sysMsg)
			if err != nil {
				log.Printf("UNABLE TO PARSE: %s", string(*hit.Source))
				return service.newInternalErrorResponse(err)
			}

			msg, err = service.toOldStyle(&sysMsg)
			if err != nil {
				log.Printf("UNABLE TO CONVERT TO OLD TYPE: %s", string(*hit.Source))
				continue
			}
			// just in case
			err = msg.Validate()
			if err != nil {
				log.Printf("UNABLE TO VALIDATE: %s", string(*hit.Source))
				continue
			}

			lines = append(lines, *msg)
		}
	}

	pagination.Count = int(searchResult.TotalHits())
	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       lines,
		Pagination: pagination,
	}

	err = resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}
	return resp
}
