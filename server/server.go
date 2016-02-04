package server

import (
	"github.com/gin-gonic/gin"
	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-alerter/client"
	"log"
	"net/http"
	"time"
)

var startTime = time.Now()

var debugMode bool

///////////////////////////////////////////////////////////

func handleGetAdminStats(c *gin.Context) {
	m := map[string]string{"start_time": startTime.String()}
	c.JSON(http.StatusOK, m)
}

func handleGetAdminSettings(c *gin.Context) {
	s := "false"
	if debugMode {
		s = "true"
	}
	m := map[string]string{"debug": s}
	c.JSON(http.StatusOK, m)
}

func handlePostAdminSettings(c *gin.Context) {
	m := map[string]string{}
	err := c.BindJSON(&m)
	if err != nil {
		c.Error(err)
		return
	}
	for k, v := range m {
		switch k {
		case "debug":
			switch v {
			case "true":
				debugMode = true
				break
			case "false":
				debugMode = false
			default:
				c.String(http.StatusBadRequest, "Illegal value for 'debug': %s", v)
				return
			}
		default:
			c.String(http.StatusBadRequest, "Unknown parameter: %s", k)
			return
		}
	}
	c.JSON(http.StatusOK, m)
}

func handlePostAdminShutdown(c *gin.Context) {
	piazza.HandlePostAdminShutdown(c)
}

func RunAlertServer(config *piazza.ServiceConfig) error {

	es := config.ElasticSearch

	conditionDB, err := client.NewConditionDB(es, "conditions")
	if err != nil {
		return err
	}
	eventDB, err := client.NewEventDB(es, "events")
	if err != nil {
		return err
	}
	alertDB, err := client.NewAlertDB(es, "alerts")
	if err != nil {
		return err
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//router.Use(gin.Logger())
	//router.Use(gin.Recovery())

	//---------------------------------

	router.GET("/", func(c *gin.Context) {
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
			c.Error(err)
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
		v, err := conditionDB.ReadByID(id)
		if err != nil {
			c.Error(err)
			return
		}
		if v == nil {
			c.IndentedJSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
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

	return router.Run(config.BindTo)
}
