package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon"
	"log"
	"net/http"
	"os"
	"strings"
	"strconv"
)

var alertId int = 1

type Alert struct {
	Id        string `json:"id"`
	Name      string `json:"name" binding:"required"`
	Condition string `json:"condition" binding:"required"`
}

type AlertDB struct {
	data map[string]Alert
}

func newAlertDB() *AlertDB {
	db := new(AlertDB)
	db.data = make(map[string]Alert)
	return db
}

func (db *AlertDB) write(alert* Alert) error {
	alert.Id = strconv.Itoa(alertId)
	db.data[alert.Id] = *alert
	alertId++
	return nil
}

func (db *AlertDB) readById(id string) *Alert {
	v, ok := db.data[id]
	if !ok {
		return nil
	}
	return &v
}

func (db *AlertDB) deleteById(id string) bool {
	_, ok := db.data[id]
	if !ok {
		return false
	}
	delete(db.data, id)
	return true
}

func runAlertServer(discoveryURL string, port string) error {

	db := newAlertDB()

	myAddress := fmt.Sprintf("%s:%s", "localhost", port)
	myURL := fmt.Sprintf("http://%s/alerts", myAddress)

	piazza.RegistryInit(discoveryURL)
	err := piazza.RegisterService("pz-alerter", "core-service", myURL)
	if err != nil {
		return err
	}

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.POST("/alerts", func(c *gin.Context) {
		var alert Alert
		err := c.BindJSON(&alert)
		if err != nil {
			c.Error(err)
			return
		}
		err = db.write(&alert)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": alert.Id})
	})

	router.GET("/alerts", func(c *gin.Context) {
		c.JSON(http.StatusOK, db.data)
	})

	router.GET("/alerts/:id", func(c *gin.Context) {
		id := c.Param("id")
		v := db.readById(id)
		if v == nil {
			c.JSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		c.JSON(http.StatusOK, v)
	})

	router.DELETE("/alerts/:id", func(c *gin.Context) {
		id := c.Param("id")
		ok := db.deleteById(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"id": id})
			return
		}
		c.JSON(http.StatusOK, nil)
	})

	err = router.Run("localhost:" + port)
	return err
}


func app() int {
	var discoveryURL = flag.String("discovery", "http://localhost:3000", "URL of pz-discovery")
	var port = flag.String("port", "12342", "port number of this pz-alerter")

	flag.Parse()

	log.Printf("starting: discovery=%s, port=%s", *discoveryURL, *port)

	err := runAlertServer(*discoveryURL, *port)
	if err != nil {
		fmt.Print(err)
		return 1
	}

	// not reached
	return 1
}

func main2(cmd string) int {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = strings.Fields("main_tester " + cmd)
	return app()
}

func main() {
	os.Exit(app())
}
