package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	piazza "github.com/venicegeo/pz-gocommon"
	"net/http"
	"os"
	"strings"
)

var pzService *piazza.PzService

///////////////////////////////////////////////////////////

func runAlertServer() error {

	es := pzService.ElasticSearch

	conditionDB, err := newConditionDB(es, "conditions")
	if err != nil {
		return err
	}
	eventDB, err := newEventDB(es, "events")
	if err != nil {
		return err
	}
	alertDB, err := newAlertDB(es, "alerts")
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

	router.POST("/events", func(c *gin.Context) {
		event := &Event{}
		err := c.BindJSON(event)
		if err != nil {
			pzService.Error("POST to /events", err)
			c.Error(err)
			return
		}
		err = eventDB.write(event)
		if err != nil {
			c.Error(err)
			return
		}
		c.IndentedJSON(http.StatusCreated, gin.H{"id": event.ID})

		alertDB.checkConditions(*event, conditionDB)
	})

	router.GET("/events", func(c *gin.Context) {
		m, err := eventDB.getAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.IndentedJSON(http.StatusOK, m)
	})

	//---------------------------------

	router.GET("/alerts", func(c *gin.Context) {

		conditionID := c.Query("condition")
		if conditionID != "" {
			v, err := alertDB.getByConditionID(conditionID)
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

			all, err := alertDB.getAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.IndentedJSON(http.StatusOK, all)

	})

	//---------------------------------
	router.POST("/conditions", func(c *gin.Context) {
		var condition Condition
		err := c.BindJSON(&condition)
		if err != nil {
			c.Error(err)
			pzService.Error("POST to /conditions", err)
			return
		}
		err = conditionDB.write(&condition)
		if err != nil {
			c.Error(err)
			return
		}
		c.IndentedJSON(http.StatusCreated, gin.H{"id": condition.ID})
	})

	/*router.PUT("/conditions", func(c *gin.Context) {
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

	router.GET("/conditions", func(c *gin.Context) {
		all, err := conditionDB.getAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.IndentedJSON(http.StatusOK, all)
	})

	router.GET("/conditions/:id", func(c *gin.Context) {
		id := c.Param("id")
		v, err := conditionDB.readByID(id)
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

	router.DELETE("/conditions/:id", func(c *gin.Context) {
		id := c.Param("id")
		ok, err := conditionDB.deleteByID(id)
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

	return router.Run(pzService.Address)
}

func app(done chan bool) int {

	var err error

	// handles the command line flags, finds the discover service, registers us,
	// and figures out our own server address
	serviceAddress, discoverAddress, debug, err := piazza.NewDiscoverService("pz-alerter", "localhost:12342", "localhost:3000")
	if err != nil {
		pzService.Fatal(err)
		return 1
	}

	pzService, err = piazza.NewPzService("pz-alerter", serviceAddress, discoverAddress, debug)
	if err != nil {
		pzService.Fatal(err)
		return 1
	}

	err = pzService.WaitForService("pz-logger", 1000)
	if err != nil {
		pzService.Fatal(err)
		return 1
	}

	err = pzService.WaitForService("pz-uuidgen", 1000)
	if err != nil {
		pzService.Fatal(err)
		return 1
	}

	if done != nil {
		done <- true
	}

	err = runAlertServer()
	if err != nil {
		pzService.Fatal(err)
		return 1
	}

	// not reached
	return 1
}

func main2(cmd string, done chan bool) int {
	flag.CommandLine = flag.NewFlagSet("pz-alerter", flag.ExitOnError)
	os.Args = strings.Fields("main_tester " + cmd)
	return app(done)
}

func main() {
	os.Exit(app(nil))
}
