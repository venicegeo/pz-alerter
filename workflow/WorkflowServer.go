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

		{"GET", "/eventType", server.handleGetEventTypes},
		{"GET", "/eventType/:id", server.handleGetEventTypeByID},
		{"POST", "/eventType", server.handlePostEventType},
		// TODO: PUT
		{"DELETE", "/eventType/:id", server.handleDeleteEventTypeByID},

		{"GET", "/event", server.handleGetEvents},
		{"GET", "/event/:id", server.handleGetEventByID},
		{"POST", "/event", server.handlePostEvent},
		// TODO: PUT
		{"DELETE", "/event/:id", server.handleDeleteEventByID},

		{"GET", "/trigger", server.handleGetTriggers},
		{"GET", "/trigger/:id", server.handleGetTriggerByID},
		{"POST", "/trigger", server.handlePostTrigger},
		// TODO: PUT
		{"DELETE", "/trigger/:id", server.handleDeleteTriggerByID},

		{"GET", "/alert", server.handleGetAlerts},
		{"GET", "/alert/:id", server.handleGetAlertByID},
		{"POST", "/alert", server.handlePostAlert},
		// TODO: PUT
		{"DELETE", "/alert/:id", server.handleDeleteAlertByID},

		{"GET", "/admin/stats", server.handleGetAdminStats},
	}

	return nil
}

//---------------------------------------------------------------------------

func (server *WorkflowServer) handleGetAdminStats(c *gin.Context) {
	resp := server.service.GetAdminStats()
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetEvents(c *gin.Context) {
	resp := server.service.GetEvents(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetEventByID(c *gin.Context) {
	resp := server.service.GetEventByID(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleDeleteAlertByID(c *gin.Context) {
	resp := server.service.DeleteAlertByID(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handlePostAlert(c *gin.Context) {
	resp := server.service.PostAlert(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetAlertByID(c *gin.Context) {
	resp := server.service.GetAlertByID(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetAlerts(c *gin.Context) {
	resp := server.service.GetAlerts(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleDeleteTriggerByID(c *gin.Context) {
	resp := server.service.DeleteTriggerByID(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetTriggerByID(c *gin.Context) {
	resp := server.service.GetTriggerByID(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetTriggers(c *gin.Context) {
	resp := server.service.GetTriggers(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handlePostTrigger(c *gin.Context) {
	resp := server.service.PostTrigger(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleDeleteEventTypeByID(c *gin.Context) {
	resp := server.service.DeleteEventTypeByID(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetEventTypeByID(c *gin.Context) {
	resp := server.service.GetEventTypeByID(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetEventTypes(c *gin.Context) {
	resp := server.service.GetEventTypes(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handlePostEventType(c *gin.Context) {
	resp := server.service.PostEventType(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleDeleteEventByID(c *gin.Context) {
	resp := server.service.DeleteEventByID(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetEventsByEventType(c *gin.Context) {
	resp := server.service.GetEventsByEventType(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handlePostEvent(c *gin.Context) {
	resp := server.service.PostEvent(c)
	c.JSON(resp.StatusCode, resp)
}

func (server *WorkflowServer) handleGetRoot(c *gin.Context) {
	type T struct {
		Message string
	}
	message := "Hi! I'm pz-workflow."
	resp := &piazza.JsonResponse{StatusCode: http.StatusOK, Data: message}
	c.JSON(resp.StatusCode, resp)
}
