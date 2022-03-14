package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/porter-dev/porter-agent/pkg/server/handlers"
)

func NewRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/incidents", handlers.GetAllIncidents)
	router.GET("/incidents/:incidentID", handlers.GetIncidentEventsByID)
	router.GET("/incidents/release/:releaseName", handlers.GetIncidentsByReleaseName)
	router.GET("/incidents/logs/:logID", handlers.GetLogs)

	return router
}
