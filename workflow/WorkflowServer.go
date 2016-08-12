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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

//---------------------------------------------------------------------------

type WorkflowServer struct {
	sysConfig *piazza.SystemConfig

	Routes  []piazza.RouteData
	service *WorkflowService
	origin  string
}

const Version = "1.0.0"

//---------------------------------------------------------------------------

func (server *WorkflowServer) Init(service *WorkflowService) error {

	server.service = service

	server.Routes = []piazza.RouteData{
		{"GET", "/", server.handleGetRoot},
		{"GET", "/version", server.handleGetVersion},

		{"GET", "/eventType", server.handleGetAllEventTypes},
		{"GET", "/eventType/:id", server.handleGetEventType},
		{"POST", "/eventType", server.handlePostEventType},
		// TODO: PUT
		{"DELETE", "/eventType/:id", server.handleDeleteEventType},

		{"GET", "/event/:id", server.handleGetEvent},
		{"GET", "/event", server.handleGetAllEvents},
		{"POST", "/event", server.handlePostEvent},
		// TODO: PUT
		{"DELETE", "/event/:id", server.handleDeleteEvent},

		{"GET", "/trigger/:id", server.handleGetTrigger},
		{"GET", "/trigger", server.handleGetAllTriggers},
		{"POST", "/trigger", server.handlePostTrigger},
		// TODO: PUT
		{"DELETE", "/trigger/:id", server.handleDeleteTrigger},

		{"GET", "/alert/:id", server.handleGetAlert},
		{"GET", "/alert", server.handleGetAllAlerts},
		{"POST", "/alert", server.handlePostAlert},
		// TODO: PUT
		{"DELETE", "/alert/:id", server.handleDeleteAlert},

		{"GET", "/admin/stats", server.handleGetStats},
	}

	server.origin = service.origin

	return nil
}

//---------------------------------------------------------------------------

func (server *WorkflowServer) handleGetRoot(c *gin.Context) {
	message := "Hi! I'm pz-workflow."
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: message}
	piazza.GinReturnJson(c, resp)
}

func (server *WorkflowServer) handleGetVersion(c *gin.Context) {
	version := piazza.Version{Version: Version}
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: version}
	piazza.GinReturnJson(c, resp)
}

func (server *WorkflowServer) handleGetStats(c *gin.Context) {
	resp := server.service.GetAdminStats()
	piazza.GinReturnJson(c, resp)
}

//---------------------------------------------------------------------------

func (server *WorkflowServer) handleGetEventType(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetEventType(id)
	piazza.GinReturnJson(c, resp)
}

func (server *WorkflowServer) handleGetAllEventTypes(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetAllEventTypes(params)
	piazza.GinReturnJson(c, resp)
}

func (server *WorkflowServer) handlePostEventType(c *gin.Context) {
	eventType := &EventType{}
	err := c.BindJSON(eventType)
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

func (server *WorkflowServer) handleDeleteEventType(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteEventType(id)
	piazza.GinReturnJson(c, resp)
}

//---------------------------------------------------------------------------

func (server *WorkflowServer) handleGetEvent(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetEvent(id)
	piazza.GinReturnJson(c, resp)
}

func (server *WorkflowServer) handleGetAllEvents(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetAllEvents(params)
	piazza.GinReturnJson(c, resp)
}

func (server *WorkflowServer) handlePostEvent(c *gin.Context) {
	event := &Event{}
	err := c.BindJSON(event)
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

func (server *WorkflowServer) handleDeleteEvent(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteEvent(id)
	piazza.GinReturnJson(c, resp)
}

//---------------------------------------------------------------------------

func (server *WorkflowServer) handleGetTrigger(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetTrigger(id)
	piazza.GinReturnJson(c, resp)
}

func (server *WorkflowServer) handleGetAllTriggers(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetAllTriggers(params)
	piazza.GinReturnJson(c, resp)
}

func (server *WorkflowServer) handlePostTrigger(c *gin.Context) {
	trigger := &Trigger{Enabled: true}
	err := c.BindJSON(trigger)
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

func (server *WorkflowServer) handleDeleteTrigger(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteTrigger(id)
	piazza.GinReturnJson(c, resp)
}

//---------------------------------------------------------------------------

func (server *WorkflowServer) handleGetAlert(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetAlert(id)
	piazza.GinReturnJson(c, resp)
}

func (server *WorkflowServer) handleGetAllAlerts(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetAllAlerts(params)
	piazza.GinReturnJson(c, resp)
}

func (server *WorkflowServer) handlePostAlert(c *gin.Context) {
	alert := &Alert{}
	err := c.BindJSON(alert)
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

func (server *WorkflowServer) handleDeleteAlert(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteAlert(id)
	piazza.GinReturnJson(c, resp)
}
