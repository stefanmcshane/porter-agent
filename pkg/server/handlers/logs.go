package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/porter-dev/porter-agent/pkg/models"
)

func ListLogBuckets(c *gin.Context) {
	podName := c.Param("podName")
	namespace := c.Param("namespace")

	keys, err := redisClient.GetKeysForResource(c.Copy(), models.PodResource, namespace, podName)
	if err != nil {
		httpLogger.Error(err, "cannot get keys for the resource")
		log.Println("cannot get keys for resource. error:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if len(keys) == 0 {
		httpLogger.Info("not log buckets found for the requested resource")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "cannot find any log buckets for the requested resource",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"availableLogBuckets": keys,
	})
}

func GetLogBucket(c *gin.Context) {
	podName := c.Param("podName")
	namespace := c.Param("namespace")
	bucket := c.Param("bucket")

	keys, err := redisClient.SearchBestMatchForBucket(c.Copy(), models.PodResource, namespace, podName, bucket)
	if err != nil {
		httpLogger.Error(err, "cannot get the best match for the log bucket")
		c.JSON(http.StatusNoContent, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": keys,
	})
}
