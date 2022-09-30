package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/porter-dev/porter-agent/api/server/handlers/incident"
)

func main() {
	router := gin.Default()

	router.GET("/incidents", incident.ListIncidents)
	router.GET("/incidents/:uid", incident.GetIncident)

	if err := router.Run(":50051"); err != nil {
		log.Fatalln(err)
	}
}
