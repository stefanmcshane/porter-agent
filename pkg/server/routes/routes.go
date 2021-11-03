package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/porter-dev/porter-agent/pkg/server/handlers"
)

func NewRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/pod/:podName/ns/:namespace/logbuckets", handlers.GetLogBuckets)

	return router
}
