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

package piazza

import (
	"errors"
	"log"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

const ginHammerTime = 3 * time.Second

// GenericServer is the basic framework for standing up a gin-based server. It has methods
// to Start and Stop as well as to define the routing paths.
type GenericServer struct {
	Sys    *SystemConfig
	pid    int
	router http.Handler
}

// RouteData describes one server route: which http verb, the path string, and the handler to use.
type RouteData struct {
	Verb    string
	Path    string
	Handler gin.HandlerFunc
}

// Stop will request the server to shutdown. It will wait for the service to die before returning.
func (server *GenericServer) Stop() error {
	sys := server.Sys

	err := syscall.Kill(server.pid, syscall.SIGINT)
	if err != nil {
		return err
	}

	err = sys.WaitForServiceToDieByAddress(sys.Name, sys.BindTo)
	return err
}

// Start will start the service. You must call Configure first.
func (server *GenericServer) Start() (chan error, error) {

	sys := server.Sys

	done := make(chan error)

	ready := make(chan bool)

	endless.DefaultHammerTime = ginHammerTime

	ginServer := endless.NewServer(server.Sys.BindTo, server.router)

	ginServer.BeforeBegin = func(_ string) {
		server.pid = syscall.Getpid()
		//log.Printf("Actual pid is %d", server.pid)

		sys.BindTo = ginServer.EndlessListener.Addr().(*net.TCPAddr).String()

		ready <- true
	}

	go func() {
		err := ginServer.ListenAndServe()
		done <- err
	}()

	<-ready

	err := sys.WaitForServiceByAddress(sys.Name, sys.BindTo)
	if err != nil {
		return nil, err
	}

	log.Printf("Server %s started on %s (%s)", sys.Name, sys.Address, sys.BindTo)

	sys.AddService(sys.Name, sys.BindTo)

	return done, nil
}

// Configure will take the give RouteData and register them with the server.
func (server *GenericServer) Configure(routeData []RouteData) error {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	//router.Use(gin.Logger())
	//router.Use(gin.Recovery())

	for _, data := range routeData {
		switch data.Verb {
		case "GET":
			router.GET(data.Path, data.Handler)
		case "POST":
			router.POST(data.Path, data.Handler)
		case "PUT":
			router.PUT(data.Path, data.Handler)
		case "DELETE":
			router.DELETE(data.Path, data.Handler)
		default:
			return errors.New("Invalid verb: " + data.Verb)
		}
	}

	server.router = router

	return nil
}
