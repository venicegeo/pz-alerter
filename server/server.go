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

package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
)

type LockedAdminSettings struct {
	sync.Mutex
	WorkflowAdminSettings
}

var settings LockedAdminSettings

type LockedAdminStats struct {
	sync.Mutex
	WorkflowAdminStats
}

var stats LockedAdminStats

func init() {
	stats.Date = time.Now()
}

func Status(c *gin.Context, code int, mssg string) {
	e := ErrorResponse{Status: code, Message: mssg}
	c.JSON(code, e)
}

//---------------------------------------------------------------------------

type Server struct {
	eventTypeDB *EventTypeDB
	eventDB     *EventDB
	triggerDB   *TriggerDB
	alertDB     *AlertDB

	uuidgen uuidgenPkg.IUuidGenService
}

var server *Server

func NewServer(es *elasticsearch.Client, uuidgen uuidgenPkg.IUuidGenService) (*Server, error) {
	var s Server
	var err error

	s.uuidgen = uuidgen

	s.eventTypeDB, err = NewEventTypeDB(&s, es, "eventtypes")
	if err != nil {
		return nil, err
	}
	err = s.eventTypeDB.Esi.Flush()
	if err != nil {
		return nil, err
	}

	s.eventDB, err = NewEventDB(&s, es, "events")
	if err != nil {
		return nil, err
	}
	err = s.eventDB.Esi.Flush()
	if err != nil {
		return nil, err
	}

	s.triggerDB, err = NewTriggerDB(&s, es, "triggers")
	if err != nil {
		return nil, err
	}
	err = s.triggerDB.Esi.Flush()
	if err != nil {
		return nil, err
	}

	s.alertDB, err = NewAlertDB(&s, es, "alerts")
	if err != nil {
		return nil, err
	}
	err = s.alertDB.Esi.Flush()
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Server) NewIdent() Ident {
	var debugIds = true

	var uuid string
	var err error

	if debugIds {
		uuid, err = s.uuidgen.GetDebugUuid("W")
		if err != nil {
			panic("uuidgen failed")
		}
	} else {
		uuid, err = s.uuidgen.GetUuid()
		if err != nil {
			panic("uuidgen failed")
		}
	}
	return Ident(uuid)
}

//---------------------------------------------------------------------------

func handleGetAdminStats(c *gin.Context) {
	stats.Lock()
	t := stats.WorkflowAdminStats
	stats.Unlock()
	c.JSON(http.StatusOK, t)
}

func handleGetAdminSettings(c *gin.Context) {
	settings.Lock()
	t := settings
	settings.Unlock()
	c.JSON(http.StatusOK, t)
}

func handlePostAdminSettings(c *gin.Context) {
	var s WorkflowAdminSettings
	err := c.BindJSON(&s)
	if err != nil {
		c.Error(err)
		return
	}
	settings.Lock()
	settings.WorkflowAdminSettings = s
	settings.Unlock()
	c.JSON(http.StatusOK, s)
}

func handlePostAdminShutdown(c *gin.Context) {
	piazza.HandlePostAdminShutdown(c)
}

func handeGetEventByID(c *gin.Context) {
	eventType := c.Param("eventType")
	s := c.Param("id")

	id := Ident(s)
	event, err := server.eventDB.GetOne(eventType, id)
	if err != nil {
		Status(c, 400, err.Error())
		return
	}
	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"id": id})
		return
	}
	c.JSON(http.StatusOK, event)
}

func handleDeleteAlertByID(c *gin.Context) {
	id := c.Param("id")
	ok, err := server.alertDB.DeleteByID("Alert", Ident(id))
	if err != nil {
		Status(c, 400, err.Error())
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"id": id})
		return
	}

	err = server.alertDB.Flush()
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	c.JSON(http.StatusOK, nil)
}

func handlePostAlert(c *gin.Context) {
	var alert Alert
	err := c.BindJSON(&alert)
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	alert.ID = server.NewIdent()

	id, err := server.alertDB.PostData("Alert", &alert, alert.ID)
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	retID := WorkflowIDResponse{ID: id}

	err = server.alertDB.Flush()
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	c.JSON(http.StatusCreated, retID)
}

func handleGetAlertByID(c *gin.Context) {
	id := c.Param("id")

	alert, err := server.alertDB.GetOne("Alert", Ident(id))
	if err != nil {
		Status(c, 400, err.Error())
		return
	}
	if alert == nil {
		c.JSON(http.StatusNotFound, gin.H{"id": id})
		return
	}
	c.JSON(http.StatusOK, alert)
}

func handleGetAlerts(c *gin.Context) {
	// TODO: conditionID := c.Query("condition")

	all, err := server.alertDB.GetAll("Alert")
	if err != nil {
		Status(c, 400, err.Error())
		return
	}
	c.JSON(http.StatusOK, all)
}

func handleDeleteTriggerByID(c *gin.Context) {
	id := c.Param("id")
	ok, err := server.triggerDB.DeleteTrigger("Trigger", Ident(id))
	if err != nil {
		Status(c, 400, err.Error())
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"id": id})
		return
	}

	err = server.triggerDB.Flush()
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	c.JSON(http.StatusOK, nil)
}

func handleGetTriggerByID(c *gin.Context) {
	id := c.Param("id")

	trigger, err := server.triggerDB.GetOne("Trigger", Ident(id))
	if err != nil {
		Status(c, 400, err.Error())
		return
	}
	if trigger == nil {
		c.JSON(http.StatusNotFound, gin.H{"id": id})
		return
	}
	c.JSON(http.StatusOK, trigger)
}

func handleGetTriggers(c *gin.Context) {
	m, err := server.triggerDB.GetAll("Trigger")
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	c.JSON(http.StatusOK, m)
}

func handlePostTrigger(c *gin.Context) {
	trigger := &Trigger{}
	err := c.BindJSON(trigger)
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	trigger.ID = server.NewIdent()

	_, err = server.triggerDB.PostTrigger("Trigger", trigger, trigger.ID)
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	a := WorkflowIDResponse{ID: trigger.ID}

	err = server.triggerDB.Flush()
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	c.JSON(http.StatusCreated, a)
}

func handleDeleteEventTypeByID(c *gin.Context) {
	id := c.Param("id")
	ok, err := server.eventTypeDB.DeleteByID("EventType", Ident(id))
	if err != nil {
		Status(c, 400, err.Error())
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"id": id})
		return
	}

	err = server.eventTypeDB.Flush()
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	c.JSON(http.StatusOK, nil)
}

func handleGetEventTypeByID(c *gin.Context) {
	id := c.Param("id")

	event, err := server.eventTypeDB.GetOne("EventType", Ident(id))
	if err != nil {
		Status(c, 400, err.Error())
		return
	}
	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"id": id})
		return
	}
	c.JSON(http.StatusOK, event)
}

func handleGetEventTypes(c *gin.Context) {
	ets, err := server.eventTypeDB.GetAll("EventType")
	if err != nil {
		Status(c, 400, err.Error())
		return
	}
	c.JSON(http.StatusOK, ets)
}

func handlePostEventType(c *gin.Context) {
	eventType := &EventType{}
	err := c.BindJSON(eventType)
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	eventType.ID = server.NewIdent()
	id, err := server.eventTypeDB.PostData("EventType", eventType, eventType.ID)
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	err = server.eventDB.AddMapping(eventType.Name, eventType.Mapping)
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	retID := WorkflowIDResponse{ID: id}

	err = server.eventTypeDB.Flush()
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	c.JSON(http.StatusCreated, retID)
}

func handeDeleteEventByID(c *gin.Context) {
	id := c.Param("id")
	eventType := c.Param("eventType")

	ok, err := server.eventDB.DeleteByID(eventType, Ident(id))
	if err != nil {
		Status(c, 400, err.Error())
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"id": id})
		return
	}

	err = server.eventDB.Flush()
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	c.JSON(http.StatusOK, nil)
}

func handleGetEvents(c *gin.Context) {
	m, err := server.eventDB.GetAll("")
	if err != nil {
		Status(c, 400, err.Error())
		return
	}
	c.JSON(http.StatusOK, m)
}

func handleGetEventsByEventType(c *gin.Context) {
	eventType := c.Param("eventType")

	m, err := server.eventDB.GetAll(eventType)
	if err != nil {
		Status(c, 400, err.Error())
		return
	}
	c.JSON(http.StatusOK, m)
}

func handlePostEvent(c *gin.Context) {

	eventType := c.Param("eventType")

	var event Event
	err := c.BindJSON(&event)
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	event.ID = server.NewIdent()
	_, err = server.eventDB.PostData(eventType, event, event.ID)
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	retID := WorkflowIDResponse{ID: event.ID}

	err = server.eventDB.Flush()
	if err != nil {
		Status(c, 400, err.Error())
		return
	}

	{
		// TODO: this should be done asynchronously
		///log.Printf("event:\n")
		///log.Printf("\tID: %v\n", event.ID)
		///log.Printf("\tType: %v\n", eventType)
		///log.Printf("\tData: %v\n", event.Data)

		// Find triggers associated with event
		triggerIDs, err := server.eventDB.PercolateEventData(eventType, event.Data, event.ID)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		// For each trigger,  apply the event data and submit job
		for _, triggerID := range *triggerIDs {
			go func(triggerID Ident) {

				///log.Printf("\ntriggerID: %v\n", triggerID)
				trigger, err := server.triggerDB.GetOne("Trigger", triggerID)
				if err != nil {
					Status(c, 400, err.Error())
					return
				}
				if trigger == nil {
					c.JSON(http.StatusNotFound, gin.H{"id": triggerID})
					return
				}
				///log.Printf("trigger: %v\n", trigger)
				///log.Printf("\tJob: %v\n\n", trigger.Job.Task)

				var jobInstance = trigger.Job.Task

				//  Not very robust,  need to find a better way
				for key, value := range event.Data {
					jobInstance = strings.Replace(jobInstance, "$"+key, fmt.Sprintf("%v", value), 1)
				}

				//log.Printf("jobInstance: %s\n\n", jobInstance)

				// Figure out how to post the jobInstance to job manager server.

			}(triggerID)
		}
	}

	c.JSON(http.StatusCreated, retID)
}

func handleHealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "Hi. I'm pz-workflow.")
}

func CreateHandlers(sys *piazza.System, logger *loggerPkg.CustomLogger, uuidgen uuidgenPkg.IUuidGenService) (http.Handler, error) {

	var err error

	esx := sys.Services[piazza.PzElasticSearch]
	if esx == nil {
		return nil, errors.New("internal error: elasticsearch not registered")
	}
	es, ok := esx.(*elasticsearch.Client)
	if !ok {
		return nil, errors.New("internl error")
	}

	server, err = NewServer(es, uuidgen)
	if err != nil {
		return nil, errors.New("internal error: server context failed to initialize")
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//router.Use(gin.Logger())
	//router.Use(gin.Recovery())

	//---------------------------------------------------------------

	router.GET("/", handleHealthCheck)

	router.POST("/v1/events/:eventType", handlePostEvent)
	router.GET("/v1/events", handleGetEvents)
	router.GET("/v1/events/:eventType", handleGetEventsByEventType)
	router.GET("/v1/events/:eventType/:id", handeGetEventByID)
	router.DELETE("/v1/events/:eventType/:id", handeDeleteEventByID)

	router.POST("/v1/eventtypes", handlePostEventType)
	router.GET("/v1/eventtypes", handleGetEventTypes)
	router.GET("/v1/eventtypes/:id", handleGetEventTypeByID)
	router.DELETE("/v1/eventtypes/:id", handleDeleteEventTypeByID)

	router.POST("/v1/triggers", handlePostTrigger)
	router.GET("/v1/triggers", handleGetTriggers)
	router.GET("/v1/triggers/:id", handleGetTriggerByID)
	router.DELETE("/v1/triggers/:id", handleDeleteTriggerByID)

	router.POST("/v1/alerts", handlePostAlert)
	router.GET("/v1/alerts", handleGetAlerts)
	router.GET("/v1/alerts/:id", handleGetAlertByID)
	router.DELETE("/v1/alerts/:id", handleDeleteAlertByID)

	router.POST("/v1/admin/settings", handlePostAdminSettings)
	router.POST("/v1/admin/shutdown", handlePostAdminShutdown)
	router.GET("/v1/admin/stats", handleGetAdminStats)
	router.GET("/v1/admin/settings", handleGetAdminSettings)

	logger.Info("handlers set")

	return router, nil
}
