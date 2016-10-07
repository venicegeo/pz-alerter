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
	"net/http"

	"github.com/gin-gonic/gin"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
)

//--------------------------------------------------

type Server struct {
	//logger  pzlogger.IClient
	Routes  []piazza.RouteData
	service *Service
}

const Version = "1.0.0"

//--------------------------------------------------

func (server *Server) Init(service *Service) {
	server.Routes = []piazza.RouteData{
		{Verb: "GET", Path: "/", Handler: server.handleGetRoot},
		{Verb: "GET", Path: "/version", Handler: server.handleGetVersion},
		{Verb: "GET", Path: "/admin/stats", Handler: server.handleGetStats},
		{Verb: "POST", Path: "/uuids", Handler: server.handlePostUuids},
	}
	server.service = service
}

func (server *Server) handleGetRoot(c *gin.Context) {
	message := "Hi. I'm pz-uuidgen."
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

// request body is ignored
// we allow a count of zero, for testing
func (server *Server) handlePostUuids(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.PostUuids(params)
	piazza.GinReturnJson(c, resp)
}
