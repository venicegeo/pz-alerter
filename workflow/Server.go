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
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

//---------------------------------------------------------------------------

type Server struct {
	//sysConfig *piazza.SystemConfig

	Routes  []piazza.RouteData
	service *Service
	origin  string
}

const Version = "1.0.0"

//---------------------------------------------------------------------------

func (server *Server) Init(service *Service) error {

	server.service = service

	server.Routes = []piazza.RouteData{
		{Verb: "GET", Path: "/", Handler: server.handleGetRoot},
		{Verb: "GET", Path: "/version", Handler: server.handleGetVersion},

		{Verb: "GET", Path: "/eventType", Handler: server.handleGetAllEventTypes},
		{Verb: "GET", Path: "/eventType/:id", Handler: server.handleGetEventType},
		{Verb: "POST", Path: "/eventType", Handler: server.handlePostEventType},
		{Verb: "POST", Path: "/eventType/query", Handler: server.handleEventTypeQuery},
		{Verb: "DELETE", Path: "/eventType/:id", Handler: server.handleDeleteEventType},

		{Verb: "GET", Path: "/event/:id", Handler: server.handleGetEvent},
		{Verb: "GET", Path: "/event", Handler: server.handleGetAllEvents},
		{Verb: "POST", Path: "/event", Handler: server.handlePostEvent},
		{Verb: "POST", Path: "/event/query", Handler: server.handleEventQuery},
		{Verb: "DELETE", Path: "/event/:id", Handler: server.handleDeleteEvent},

		{Verb: "GET", Path: "/trigger/:id", Handler: server.handleGetTrigger},
		{Verb: "GET", Path: "/trigger", Handler: server.handleGetAllTriggers},
		{Verb: "POST", Path: "/trigger", Handler: server.handlePostTrigger},
		{Verb: "POST", Path: "/trigger/query", Handler: server.handleTriggerQuery},
		{Verb: "PUT", Path: "/trigger/:id", Handler: server.handlePutTrigger},
		{Verb: "DELETE", Path: "/trigger/:id", Handler: server.handleDeleteTrigger},

		{Verb: "GET", Path: "/alert/:id", Handler: server.handleGetAlert},
		{Verb: "GET", Path: "/alert", Handler: server.handleGetAllAlerts},
		{Verb: "POST", Path: "/alert", Handler: server.handlePostAlert},
		{Verb: "POST", Path: "/alert/query", Handler: server.handleAlertQuery},
		{Verb: "DELETE", Path: "/alert/:id", Handler: server.handleDeleteAlert},

		{Verb: "GET", Path: "/admin/stats", Handler: server.handleGetStats},

		{Verb: "GET", Path: "/_test/elasticsearch/version", Handler: server.handleTestElasticsearchVersion},
		{Verb: "GET", Path: "/_test/elasticsearch/data/:id", Handler: server.handleTestElasticsearchGetOne},
		{Verb: "POST", Path: "/_test/elasticsearch/data", Handler: server.handleTestElasticsearchPost},
	}

	server.origin = service.origin

	return nil
}

//---------------------------------------------------------------------------

func (server *Server) handleGetRoot(c *gin.Context) {
	message := "Hi! I'm pz-workflow."
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: message}
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleGetVersion(c *gin.Context) {
	version := piazza.Version{Version: Version}
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: version}
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleGetStats(c *gin.Context) {
	resp := server.service.GetStats()
	piazza.GinReturnJson(c, resp)
}

//---------------------------------------------------------------------------

func (server *Server) handleGetEventType(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetEventType(id)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleGetAllEventTypes(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetAllEventTypes(params)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handlePostEventType(c *gin.Context) {
	eventType := &EventType{}
	decoder := json.NewDecoder(c.Request.Body)
	decoder.UseNumber()
	err := decoder.Decode(&eventType)
	if eventType == nil {
		err = errors.New("EventType is <nil>")
	}
	if err != nil {
		resp := &piazza.JsonResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Origin:     server.origin,
		}
		piazza.GinReturnJson(c, resp)
		return
	}

	resp := server.service.PostEventType(eventType)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleEventTypeQuery(c *gin.Context) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		resp := &piazza.JsonResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Origin:     server.origin,
		}
		piazza.GinReturnJson(c, resp)
		return
	}

	jsonString := buf.String()
	params := piazza.NewQueryParams(c.Request)

	var resp *piazza.JsonResponse

	resp = server.service.QueryEventTypes(jsonString, params)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleDeleteEventType(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteEventType(id)
	piazza.GinReturnJson(c, resp)
}

//---------------------------------------------------------------------------

func (server *Server) handleGetEvent(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetEvent(id)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleGetAllEvents(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetAllEvents(params)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handlePostEvent(c *gin.Context) {
	event := &Event{}
	decoder := json.NewDecoder(c.Request.Body)
	decoder.UseNumber()
	err := decoder.Decode(&event)
	if event == nil {
		err = errors.New("Event is <nil>")
	}
	if err != nil {
		resp := &piazza.JsonResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Origin:     server.origin,
		}
		piazza.GinReturnJson(c, resp)
		return
	}

	var resp *piazza.JsonResponse

	if event.CronSchedule != "" {
		resp = server.service.PostRepeatingEvent(event)
	} else {
		resp = server.service.PostEvent(event)
	}
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleEventQuery(c *gin.Context) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		resp := &piazza.JsonResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Origin:     server.origin,
		}
		piazza.GinReturnJson(c, resp)
		return
	}

	jsonString := buf.String()
	params := piazza.NewQueryParams(c.Request)

	var resp *piazza.JsonResponse

	resp = server.service.QueryEvents(jsonString, params)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleDeleteEvent(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteEvent(id)
	piazza.GinReturnJson(c, resp)
}

//---------------------------------------------------------------------------

func (server *Server) handleGetTrigger(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetTrigger(id)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleGetAllTriggers(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetAllTriggers(params)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handlePostTrigger(c *gin.Context) {
	trigger := &Trigger{Enabled: true}
	decoder := json.NewDecoder(c.Request.Body)
	decoder.UseNumber()
	err := decoder.Decode(&trigger)
	if trigger == nil {
		err = errors.New("Trigger is <nil>")
	}
	if err != nil {
		resp := &piazza.JsonResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Origin:     server.origin,
		}
		piazza.GinReturnJson(c, resp)
		return
	}
	resp := server.service.PostTrigger(trigger)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleTriggerQuery(c *gin.Context) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		resp := &piazza.JsonResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Origin:     server.origin,
		}
		piazza.GinReturnJson(c, resp)
		return
	}

	jsonString := buf.String()
	params := piazza.NewQueryParams(c.Request)

	var resp *piazza.JsonResponse

	resp = server.service.QueryTriggers(jsonString, params)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handlePutTrigger(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	update := &TriggerUpdate{}
	err := c.BindJSON(update)
	if err != nil {
		resp := &piazza.JsonResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Origin:     server.origin,
		}
		piazza.GinReturnJson(c, resp)
		return
	}
	resp := server.service.PutTrigger(id, update)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleDeleteTrigger(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteTrigger(id)
	piazza.GinReturnJson(c, resp)
}

//---------------------------------------------------------------------------

func (server *Server) handleGetAlert(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetAlert(id)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleGetAllAlerts(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetAllAlerts(params)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleAlertQuery(c *gin.Context) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		resp := &piazza.JsonResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Origin:     server.origin,
		}
		piazza.GinReturnJson(c, resp)
		return
	}

	jsonString := buf.String()
	params := piazza.NewQueryParams(c.Request)

	var resp *piazza.JsonResponse

	resp = server.service.QueryAlerts(jsonString, params)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handlePostAlert(c *gin.Context) {
	alert := &Alert{}
	decoder := json.NewDecoder(c.Request.Body)
	decoder.UseNumber()
	err := decoder.Decode(&alert)
	if alert == nil {
		err = errors.New("Alert is <nil>")
	}
	if err != nil {
		resp := &piazza.JsonResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Origin:     server.origin,
		}
		piazza.GinReturnJson(c, resp)
		return
	}
	resp := server.service.PostAlert(alert)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleDeleteAlert(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteAlert(id)
	piazza.GinReturnJson(c, resp)
}

//---------------------------------------------------------------------

func (server *Server) handleTestElasticsearchVersion(c *gin.Context) {
	resp := server.service.TestElasticsearchVersion()
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleTestElasticsearchGetOne(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.TestElasticsearchGetOne(id)
	piazza.GinReturnJson(c, resp)
}

func (server *Server) handleTestElasticsearchPost(c *gin.Context) {
	body := &TestElasticsearchBody{}
	err := c.BindJSON(body)
	if err != nil {
		resp := &piazza.JsonResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Origin:     server.origin,
		}
		piazza.GinReturnJson(c, resp)
		return
	}

	resp := server.service.TestElasticsearchPost(body)

	piazza.GinReturnJson(c, resp)
}
