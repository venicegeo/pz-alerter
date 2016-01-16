package main

import (
	"flag"
	"fmt"
	"log"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
	"github.com/venicegeo/pz-gocommon"
	"net/http"
)

var database map[int]Alert
var id int

type Alert struct {
	id int
	Name string `json:"name" binding:"required"`
	Condition string `json:"condition" binding:"required"`
}

func runAlertServer(discoveryURL string, port string) error {

	database = make(map[int]Alert)

	myAddress := fmt.Sprintf("%s:%s", "localhost", port)
	myURL := fmt.Sprintf("http://%s/alerts", myAddress)

	piazza.RegistryInit(discoveryURL)
	err := piazza.RegisterService("pz-alerter", "core-service", myURL)
	if err != nil {
		return err
	}

	// TODO
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.POST("/alerts", func(c *gin.Context) {
		var js Alert
		err := c.BindJSON(&js)
		if err != nil {
			c.Error(err)
			return
		}
		js.id = id
		database[id] = js
		id++

		c.JSON(http.StatusCreated, gin.H{"id": fmt.Sprintf("%d",id)})
		//		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})


	})

	// Listen and server on 0.0.0.0:8080
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
