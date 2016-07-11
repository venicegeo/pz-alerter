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
}

//---------------------------------------------------------------------------

func (server *WorkflowServer) Init(service *WorkflowService) error {

	server.service = service

	server.Routes = []piazza.RouteData{
		{"GET", "/", server.handleGetRoot},

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

	return nil
}

//---------------------------------------------------------------------------

func (server *WorkflowServer) handleGetRoot(c *gin.Context) {
	type T struct {
		Message string
	}
	message := "Hi! I'm pz-workflow."
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: message}
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetStats(c *gin.Context) {
	resp := server.service.GetAdminStats()
	c.IndentedJSON(resp.StatusCode, resp)
}

//---------------------------------------------------------------------------

func (server *WorkflowServer) handleGetEventType(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetEventType(id)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetAllEventTypes(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetAllEventTypes(params)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handlePostEventType(c *gin.Context) {
	eventType := &EventType{}
	err := c.BindJSON(eventType)
	if err != nil {
		resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Message: err.Error()}
		c.IndentedJSON(resp.StatusCode, resp)
		return
	}
	resp := server.service.PostEventType(eventType)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleDeleteEventType(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteEventType(id)
	c.IndentedJSON(resp.StatusCode, resp)
}

//---------------------------------------------------------------------------

func (server *WorkflowServer) handleGetEvent(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetEvent(id)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetAllEvents(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetAllEvents(params)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handlePostEvent(c *gin.Context) {
	event := &Event{}
	err := c.BindJSON(event)
	if err != nil {
		resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Message: err.Error()}
		c.IndentedJSON(resp.StatusCode, resp)
		return
	}
	resp := server.service.PostEvent(event)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleDeleteEvent(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteEvent(id)
	c.IndentedJSON(resp.StatusCode, resp)
}

//---------------------------------------------------------------------------

func (server *WorkflowServer) handleGetTrigger(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetTrigger(id)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetAllTriggers(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetAllTriggers(params)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handlePostTrigger(c *gin.Context) {
	trigger := &Trigger{}
	err := c.BindJSON(trigger)
	if err != nil {
		resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Message: err.Error()}
		c.IndentedJSON(resp.StatusCode, resp)
		return
	}
	resp := server.service.PostTrigger(trigger)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleDeleteTrigger(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteTrigger(id)
	c.IndentedJSON(resp.StatusCode, resp)
}

//---------------------------------------------------------------------------

func (server *WorkflowServer) handleGetAlert(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.GetAlert(id)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetAllAlerts(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.GetAllAlerts(params)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handlePostAlert(c *gin.Context) {
	alert := &Alert{}
	err := c.BindJSON(alert)
	if err != nil {
		resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Message: err.Error()}
		c.IndentedJSON(resp.StatusCode, resp)
		return
	}
	resp := server.service.PostAlert(alert)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleDeleteAlert(c *gin.Context) {
	id := piazza.Ident(c.Param("id"))
	resp := server.service.DeleteAlert(id)
	c.IndentedJSON(resp.StatusCode, resp)
}
