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
	"log"
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

func Status(c *gin.Context, code int, mssg string) {
	e := ErrorResponse{Status: code, Message: mssg}
	c.JSON(code, e)
}

var NewIdent func() Ident

func CreateHandlers(sys *piazza.System, logger *loggerPkg.CustomLogger, uuidgenner uuidgenPkg.IUuidGenService) (http.Handler, error) {

	var debugIds = true

	NewIdent = func() Ident {
		var uuid string
		var err error

		if debugIds {
			uuid, err = uuidgenner.GetDebugUuid("W")
			if err != nil {
				panic("uuidgen failed")
			}
		} else {
			uuid, err = uuidgenner.GetUuid()
			if err != nil {
				panic("uuidgen failed")
			}
		}
		return Ident(uuid)
	}

	esx := sys.Services[piazza.PzElasticSearch]
	if esx == nil {
		return nil, errors.New("internal error: elasticsearch not registered")
	}
	es, ok := esx.(*elasticsearch.ElasticsearchClient)
	if !ok {
		return nil, errors.New("internl error")
	}

	alertDB, err := NewAlertDB(es, "alerts")
	if err != nil {
		return nil, err
	}

	triggerDB, err := NewTriggerDB(es, "triggers")
	if err != nil {
		return nil, err
	}

	eventDB, err := NewEventDB(es, "events")
	if err != nil {
		return nil, err
	}

	eventTypeDB, err := NewEventTypeDB(es, "eventtypes")
	if err != nil {
		return nil, err
	}

	err = alertDB.Esi.Flush()
	if err != nil {
		return nil, err
	}
	err = triggerDB.Esi.Flush()
	if err != nil {
		return nil, err
	}
	err = alertDB.Esi.Flush()
	if err != nil {
		return nil, err
	}
	err = triggerDB.Esi.Flush()
	if err != nil {
		return nil, err
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//router.Use(gin.Logger())
	//router.Use(gin.Recovery())

	//---------------------------------------------------------------

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hi. I'm pz-workflow.")
	})

	// ---------------------- EVENTS ----------------------

	router.POST("/v1/events/:eventType", func(c *gin.Context) {

		eventType := c.Param("eventType")

		var event Event
		err := c.BindJSON(&event)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		event.ID = NewIdent()
		_, err = eventDB.PostData(eventType, event, event.ID)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		retId := WorkflowIdResponse{ID: event.ID}

		err = eventDB.Flush()
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
			triggerIds, err := eventDB.PercolateEventData(eventType, event.Data, event.ID, alertDB)
			if err != nil {
				Status(c, 400, err.Error())
				return
			}

			// For each trigger,  apply the event data and submit job
			for _, triggerId := range *triggerIds {
				go func(triggerId Ident) {

					///log.Printf("\ntriggerId: %v\n", triggerId)
					var trigger Trigger

					ok, err := triggerDB.GetById("Trigger", triggerId, &trigger)
					if err != nil {
						Status(c, 400, err.Error())
						return
					}
					if !ok {
						c.JSON(http.StatusNotFound, gin.H{"id": triggerId})
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

				}(triggerId)
			}
		}

		c.JSON(http.StatusCreated, retId)
	})

	router.GET("/v1/events", func(c *gin.Context) {
		m, err := eventDB.GetAll("")
		if err != nil {
			Status(c, 400, err.Error())
			return
		}
		c.JSON(http.StatusOK, m)
	})

	router.GET("/v1/events/:eventType", func(c *gin.Context) {
		eventType := c.Param("eventType")

		m, err := eventDB.GetAll(eventType)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}
		c.JSON(http.StatusOK, m)
	})

	router.GET("/v1/events/:eventType/:id", func(c *gin.Context) {
		eventType := c.Param("eventType")
		s := c.Param("id")

		id := Ident(s)
		var v Event
		ok, err := eventDB.GetById(eventType, id, &v)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		c.JSON(http.StatusOK, v)
	})

	router.DELETE("/v1/events/:eventType/:id", func(c *gin.Context) {
		id := c.Param("id")
		eventType := c.Param("eventType")

		ok, err := eventDB.DeleteByID(eventType, id)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"id": id})
			return
		}

		err = eventDB.Flush()
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		c.JSON(http.StatusOK, nil)
	})

	// ---------------------- EVENT TYPES ----------------------

	router.POST("/v1/eventtypes", func(c *gin.Context) {
		eventType := &EventType{}
		err := c.BindJSON(eventType)
		if err != nil {
			Status(c, 403, err.Error())
			return
		}

		eventType.ID = NewIdent()
		id, err := eventTypeDB.PostData("EventType", eventType, eventType.ID)
		if err != nil {
			Status(c, 401, err.Error())
			return
		}

		err = eventDB.AddMapping(eventType.Name, eventType.Mapping)
		if err != nil {
			Status(c, 402, err.Error())
			return
		}

		retId := WorkflowIdResponse{ID: id}

		err = eventTypeDB.Flush()
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		{
			id := Ident(retId.ID)
			var v EventType
			ok, err := eventTypeDB.GetById("EventType", id, &v)
			if err != nil {
				Status(c, 400, err.Error())
				return
			}
			if !ok {
				Status(c, 410, err.Error())
				return
			}
			if v.ID != retId.ID {
				log.Printf("*************** %s %s ************", v.ID, retId.ID)
				panic(1)
			}
		}

		c.JSON(http.StatusCreated, retId)
	})

	// returns a list of all IDs
	router.GET("/v1/eventtypes", func(c *gin.Context) {
		m, err := eventTypeDB.GetAll("EventType")
		if err != nil {
			Status(c, 400, err.Error())
			return
		}
		c.JSON(http.StatusOK, m)
	})

	// returns info on a given ID
	router.GET("/v1/eventtypes/:id", func(c *gin.Context) {
		s := c.Param("id")

		id := Ident(s)
		var v Event
		ok, err := eventTypeDB.GetById("EventType", id, &v)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		c.JSON(http.StatusOK, v)
	})

	router.DELETE("/v1/eventtypes/:id", func(c *gin.Context) {
		id := c.Param("id")
		ok, err := eventTypeDB.DeleteByID("EventType", id)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"id": id})
			return
		}

		err = eventTypeDB.Flush()
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		c.JSON(http.StatusOK, nil)
	})

	// ---------------------- TRIGGERS ----------------------

	router.POST("/v1/triggers", func(c *gin.Context) {
		trigger := &Trigger{}
		err := c.BindJSON(trigger)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		trigger.ID = NewIdent()

		_, err = triggerDB.PostTrigger("Trigger", trigger, trigger.ID, eventDB)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		a := WorkflowIdResponse{ID: trigger.ID}

		err = triggerDB.Flush()
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		c.JSON(http.StatusCreated, a)
	})

	router.GET("/v1/triggers", func(c *gin.Context) {
		m, err := triggerDB.GetAll("Trigger")
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		c.JSON(http.StatusOK, m)
	})

	router.GET("/v1/triggers/:id", func(c *gin.Context) {
		s := c.Param("id")

		id := Ident(s)
		var v Trigger
		ok, err := triggerDB.GetById("Trigger", id, &v)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		c.JSON(http.StatusOK, v)
	})

	router.DELETE("/v1/triggers/:id", func(c *gin.Context) {
		id := c.Param("id")
		ok, err := triggerDB.DeleteTrigger("Trigger", Ident(id), eventDB)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"id": id})
			return
		}

		err = triggerDB.Flush()
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		c.JSON(http.StatusOK, nil)
	})

	// ---------------------- ALERTS ----------------------

	router.GET("/v1/alerts", func(c *gin.Context) {

		conditionID := c.Query("condition")
		if conditionID != "" {
			v, err := alertDB.GetByConditionID("Alert", conditionID)
			if err != nil {
				Status(c, 400, err.Error())
				return
			}
			if v == nil {
				Status(c, 400, err.Error())
				return
			}
			c.JSON(http.StatusOK, v)
			return
		}

		all, err := alertDB.GetAll("Alert")
		if err != nil {
			Status(c, 400, err.Error())
			return
		}
		c.JSON(http.StatusOK, all)
	})

	router.GET("/v1/alerts/:id", func(c *gin.Context) {
		s := c.Param("id")

		id := Ident(s)
		var alert Alert
		ok, err := alertDB.GetById("Alert", id, &alert)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		c.JSON(http.StatusOK, alert)
	})

	router.POST("/v1/alerts", func(c *gin.Context) {
		var alert Alert
		err := c.BindJSON(&alert)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		alert.ID = NewIdent()

		id, err := alertDB.PostData("Alert", &alert, alert.ID)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		retId := WorkflowIdResponse{ID: id}

		err = alertDB.Flush()
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		c.JSON(http.StatusCreated, retId)
	})

	router.DELETE("/v1/alerts/:id", func(c *gin.Context) {
		id := c.Param("id")
		ok, err := alertDB.DeleteByID("Alert", id)
		if err != nil {
			Status(c, 400, err.Error())
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"id": id})
			return
		}

		err = alertDB.Flush()
		if err != nil {
			Status(c, 400, err.Error())
			return
		}

		c.JSON(http.StatusOK, nil)
	})

	//-----------------------------------------------------------------------

	router.GET("/v1/admin/stats", func(c *gin.Context) { handleGetAdminStats(c) })

	router.GET("/v1/admin/settings", func(c *gin.Context) { handleGetAdminSettings(c) })
	router.POST("/v1/admin/settings", func(c *gin.Context) { handlePostAdminSettings(c) })

	router.POST("/v1/admin/shutdown", func(c *gin.Context) { handlePostAdminShutdown(c) })

	logger.Info("handlers set")

	return router, nil
}
