package server

import (
	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-alerter/client"
	"github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	uuidgenPkg "github.com/venicegeo/pz-uuidgen/client"
	"log"
	"net/http"
	"sync"
	"time"
)

type LockedAdminSettings struct {
	sync.Mutex
	client.AlerterAdminSettings
}

var settings LockedAdminSettings

type LockedAdminStats struct {
	sync.Mutex
	client.AlerterAdminStats
}

var stats LockedAdminStats

func init() {
	stats.Date = time.Now()
}

///////////////////////////////////////////////////////////

func handleGetAdminStats(c *gin.Context) {
	stats.Lock()
	t := stats.AlerterAdminStats
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
	var s client.AlerterAdminSettings
	err := c.BindJSON(&s)
	if err != nil {
		c.Error(err)
		return
	}
	settings.Lock()
	settings.AlerterAdminSettings = s
	settings.Unlock()
	c.JSON(http.StatusOK, s)
}

func handlePostAdminShutdown(c *gin.Context) {
	piazza.HandlePostAdminShutdown(c)
}

func CreateHandlers(sys *piazza.System, logger loggerPkg.ILoggerService, uuidgenner uuidgenPkg.IUuidGenService) (http.Handler, error) {

	es := sys.ElasticSearchService

	alertDB, err := client.NewAlertDB(es, "alerts", "Alert")
	if err != nil {
		return nil, err
	}

	triggerDB, err := client.NewTriggerDB(es, "triggers", "Triggers")
	if err != nil {
		return nil, err
	}

	eventDB, err := client.NewResourceDB(es, "events", "Events")
	if err != nil {
		return nil, err
	}

	err = es.FlushIndex("events")
	if err != nil {
		return nil,err
	}
	err = es.FlushIndex("alerts")
	if err != nil {
		return nil,err
	}
	err = es.FlushIndex("triggers")
	if err != nil {
		return nil,err
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//router.Use(gin.Logger())
	//router.Use(gin.Recovery())

	//---------------------------------

	router.GET("/", func(c *gin.Context) {
		log.Print("got health-check request")
		c.String(http.StatusOK, "Hi. I'm pz-alerter.")
	})

	//---------------------------------

	router.POST("/v1/events", func(c *gin.Context) {
		event := &client.Event{}
		err := c.BindJSON(event)
		if err != nil {
			//pzService.Error("POST to /v1/events", err)
			log.Printf("POST to /v1/events: %v", err)
			c.Error(err)
			return
		}

		event.ID = client.NewEventID()
		id, err := eventDB.PostData(event, event.ID)
		if err != nil {
			c.Error(err)
			return
		}

		a := client.AlerterIdResponse{ID: id}
		c.IndentedJSON(http.StatusCreated, a)

		triggerDB.CheckTriggers(*event, alertDB)
	})

	router.GET("/v1/events", func(c *gin.Context) {
		m, err := eventDB.GetAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.IndentedJSON(http.StatusOK, m)
	})

	router.DELETE("/v1/events/:id", func(c *gin.Context) {
		id := c.Param("id")
		ok, err := eventDB.DeleteByID(id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"id": id})
			return
		}
		if !ok {
			c.IndentedJSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		c.IndentedJSON(http.StatusOK, nil)
	})

	//---------------------------------

	router.POST("/v1/triggers", func(c *gin.Context) {
		trigger := &client.Trigger{}
		err := c.BindJSON(trigger)
		if err != nil {
			//pzService.Error("POST to /v1/events", err)
			log.Printf("POST to /v1/triggers: %v", err)
			c.Error(err)
			return
		}

		trigger.ID = client.NewTriggerIdent()

		_, err = triggerDB.PostData(trigger, trigger.ID)
		if err != nil {
			c.Error(err)
			return
		}

		a := client.AlerterIdResponse{ID: trigger.ID}
		c.IndentedJSON(http.StatusCreated, a)
	})

	router.GET("/v1/triggers", func(c *gin.Context) {
		m, err := triggerDB.GetAll()
		if err != nil {
			c.Error(err)
			return
		}

		c.IndentedJSON(http.StatusOK, m)
	})

	router.GET("/v1/triggers/:id", func(c *gin.Context) {
		s := c.Param("id")

		id := client.Ident(s)
		var v client.Trigger
		ok, err := triggerDB.GetById(id, &v)
		if err != nil {
			c.Error(err)
			return
		}
		if !ok {
			c.IndentedJSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		c.IndentedJSON(http.StatusOK, v)
	})

	router.DELETE("/v1/triggers/:id", func(c *gin.Context) {
		id := c.Param("id")
		ok, err := triggerDB.DeleteByID(id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"id": id})
			return
		}
		if !ok {
			c.IndentedJSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		c.IndentedJSON(http.StatusOK, nil)
	})

	//---------------------------------

	router.GET("/v1/alerts", func(c *gin.Context) {

		conditionID := c.Query("condition")
		if conditionID != "" {
			v, err := alertDB.GetByConditionID(conditionID)
			if err != nil {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"condition_id": conditionID})
				return
			}
			if v == nil {
				c.IndentedJSON(http.StatusNotFound, gin.H{"condition_id": conditionID})
				return
			}
			c.IndentedJSON(http.StatusOK, v)
			return
		}

		all, err := alertDB.GetAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.IndentedJSON(http.StatusOK, all)
	})

	router.GET("/v1/alerts/:id", func(c *gin.Context) {
		s := c.Param("id")

		id := client.Ident(s)
		var alert client.Alert
		ok, err := alertDB.GetById(id, &alert)
		if err != nil {
			c.Error(err)
			return
		}
		if !ok {
			c.IndentedJSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		c.IndentedJSON(http.StatusOK, alert)
	})

	router.POST("/v1/alerts", func(c *gin.Context) {
		var alert client.Alert
		err := c.BindJSON(&alert)
		if err != nil {
			c.Error(err)
			log.Printf("ERROR: POST to /v1/alerts %v", err)
			return
		}

		alert.ID = client.NewAlertIdent()

		_, err = alertDB.PostData(&alert, alert.ID)
		if err != nil {
			c.AbortWithError(499, err)
			return
		}
		c.IndentedJSON(http.StatusCreated, gin.H{"id": alert.ID})
	})

	router.DELETE("/v1/alerts/:id", func(c *gin.Context) {
		id := c.Param("id")
		ok, err := alertDB.DeleteByID(id)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"id": id})
			return
		}
		if !ok {
			c.IndentedJSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		c.IndentedJSON(http.StatusOK, nil)
	})

	//---------------------------------

	router.GET("/v1/admin/stats", func(c *gin.Context) { handleGetAdminStats(c) })

	router.GET("/v1/admin/settings", func(c *gin.Context) { handleGetAdminSettings(c) })
	router.POST("/v1/admin/settings", func(c *gin.Context) { handlePostAdminSettings(c) })

	router.POST("/v1/admin/shutdown", func(c *gin.Context) { handlePostAdminShutdown(c) })

	return router, nil
}
