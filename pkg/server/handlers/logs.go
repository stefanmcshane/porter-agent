package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/porter-dev/porter-agent/pkg/models"
)

func GetLogBuckets(c *gin.Context) {
	logger := logr.FromContext(c)

	podName := c.Param("podName")
	namespace := c.Param("namespace")

	keys, err := redisClient.GetKeysForResource(context.Background(), models.PodResource, namespace, podName)
	if err != nil {
		logger.Error(err, "cannot get keys for the resource")
		log.Println("cannot get keys for resource. error:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if len(keys) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "cannot find any keys for that pattern",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"availableLogBuckets": keys,
	})
}
