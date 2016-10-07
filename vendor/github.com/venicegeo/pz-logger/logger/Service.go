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
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

const schema = "LogData7"

type Service struct {
	sync.Mutex

	stats  Stats
	origin string

	esIndex elasticsearch.IIndex
	id      int
}

func (service *Service) Init(sys *piazza.SystemConfig, esIndex elasticsearch.IIndex) error {
	var err error

	service.stats.CreatedOn = time.Now()
	/***
	err = esIndex.Delete()
	if err != nil {
		log.Fatal(err)
	}
	if esIndex.IndexExists() {
		log.Fatal("index still exists")
	}
	err = esIndex.Create()
	if err != nil {
		log.Fatal(err)
	}
	***/

	ok, err := esIndex.IndexExists()
	if err != nil {
		return err
	}
	if !ok {
		log.Printf("Creating index: %s", esIndex.IndexName())
		err = esIndex.Create("")
		if err != nil {
			log.Fatal(err)
		}
	}

	ok, err = esIndex.TypeExists(schema)
	if err != nil {
		return err
	}
	if !ok {
		log.Printf("Creating type: %s", schema)

		mapping :=
			`{
			"LogData7":{
				"dynamic": "strict",
				"properties": {
					"service": {
						"type": "string",
						"store": true,
						"index": "not_analyzed"
					},
					"address": {
						"type": "string",
						"store": true,
						"index": "not_analyzed"
					},
					"createdOn": {
						"type": "date",
						"store": true,
						"index": "not_analyzed"
					},
					"severity": {
						"type": "string",
						"store": true,
						"index": "not_analyzed"
					},
					"message": {
						"type": "string",
						"store": true,
						"index": "analyzed"
					}
				}
			}
		}`

		err = esIndex.SetMapping(schema, piazza.JsonString(mapping))
		if err != nil {
			log.Printf("LoggerService.Init: %s", err.Error())
			return err
		}
	}

	service.esIndex = esIndex

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

func (service *Service) PostMessage(mssg *Message) *piazza.JsonResponse {
	err := mssg.Validate()
	if err != nil {
		return service.newBadRequestResponse(err)
	}

	service.Lock()
	idStr := strconv.Itoa(service.id)
	service.id++
	service.Unlock()

	_, err = service.esIndex.PostData(schema, idStr, mssg)
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	service.stats.NumMessages++

	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       mssg,
	}

	err = resp.SetType()
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

func (service *Service) GetMessage(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	var err error

	pagination, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.newBadRequestResponse(err)
	}

	dsl, err := createQueryDslAsString(pagination, params)
	if err != nil {
		return service.newBadRequestResponse(err)
	}

	var searchResult *elasticsearch.SearchResult

	if dsl == "" {
		searchResult, err = service.esIndex.FilterByMatchAll(schema, pagination)
	} else {
		searchResult, err = service.esIndex.SearchByJSON(schema, dsl)
	}

	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	var lines = make([]Message, 0)

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			if hit.Source == nil {
				log.Printf("null source hit")
				continue
			}

			var msg Message
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
				"createdOn": rangeParams,
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
