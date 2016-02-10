package server

import (
	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-alerter/client"
	piazza "github.com/venicegeo/pz-gocommon"
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

	conditionDB, err := client.NewConditionDB(es, "conditions")
	if err != nil {
		return nil, err
	}
	eventDB, err := client.NewEventDB(es, "events")
	if err != nil {
		return nil, err
	}
	alertDB, err := client.NewAlertDB(es, "alerts")
	if err != nil {
		return nil, err
	}
	actionDB, err := client.NewActionDB(es, "actions")
	if err != nil {
		return nil, err
	}


	//err = es.Flush("conditions")
	//if err != nil {
	//	return nil,err
	//}
	err = es.Flush("events")
	if err != nil {
		return nil,err
	}
	err = es.Flush("alerts")
	if err != nil {
		return nil,err
	}
	err = es.Flush("actions")
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
		err = eventDB.Write(event)
		if err != nil {
			c.Error(err)
			return
		}

		a := client.AlerterIdResponse{ID: event.ID}
		c.IndentedJSON(http.StatusCreated, a)

		alertDB.CheckConditions(*event, conditionDB)
	})

	router.GET("/v1/events", func(c *gin.Context) {
		m, err := eventDB.GetAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.IndentedJSON(http.StatusOK, m)
	})

	//---------------------------------

	router.POST("/v1/actions", func(c *gin.Context) {
		action := &client.Action{}
		err := c.BindJSON(action)
		if err != nil {
			//pzService.Error("POST to /v1/events", err)
			log.Printf("POST to /v1/actions: %v", err)
			c.Error(err)
			return
		}

		action.ID = client.NewActionIdent()

		err = actionDB.Write(action)
		if err != nil {
			c.Error(err)
			return
		}

		a := client.AlerterIdResponse{ID: action.ID}
		c.IndentedJSON(http.StatusCreated, a)
	})

	router.GET("/v1/actions", func(c *gin.Context) {
		m, err := actionDB.GetAll()
		if err != nil {
			c.Error(err)
			return
		}

		c.IndentedJSON(http.StatusOK, m)
		log.Printf("%#v", m)
	})

	router.GET("/v1/actions/:id", func(c *gin.Context) {
		s := c.Param("id")

		id := client.Ident(s)
		v, err := actionDB.GetByID(id)
		if err != nil {
			c.Error(err)
			return
		}
		if v == nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		log.Print("266622 ", s)
		c.IndentedJSON(http.StatusOK, v)
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

	//---------------------------------
	router.POST("/v1/conditions", func(c *gin.Context) {
		var condition client.Condition
		err := c.BindJSON(&condition)
		if err != nil {
			c.Error(err)
			//pzService.Error("POST to /v1/conditions", err)
			log.Printf("ERROR: POST to /v1/conditions %v", err)
			return
		}
		err = conditionDB.Write(&condition)
		if err != nil {
			c.AbortWithError(499, err)
			return
		}
		c.IndentedJSON(http.StatusCreated, gin.H{"id": condition.ID})
	})

	/*router.PUT("/v1/conditions", func(c *gin.Context) {
		var condition Condition
		err := c.BindJSON(&condition)
		if err != nil {
			c.Error(err)
			return
		}
		ok := conditionDB.update(&condition)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"id": condition.ID})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": condition.ID})
	})*/

	router.GET("/v1/conditions", func(c *gin.Context) {
		all, err := conditionDB.GetAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.IndentedJSON(http.StatusOK, all)
	})

	router.GET("/v1/conditions/:id", func(c *gin.Context) {
		id := c.Param("id")
		v, err := conditionDB.ReadByID(client.Ident(id))
		if err != nil {
			log.Printf("+++0")
			c.Error(err)
			return
		}
		if v == nil {
			log.Printf("+++1")
			c.IndentedJSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		log.Printf("+++2")
		c.IndentedJSON(http.StatusOK, v)
	})

	router.DELETE("/v1/conditions/:id", func(c *gin.Context) {
		id := c.Param("id")
		ok, err := conditionDB.DeleteByID(id)
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
