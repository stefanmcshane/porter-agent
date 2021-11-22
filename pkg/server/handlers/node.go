package handlers

import (
	"log"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

func ListNodes(c *gin.Context) {
	keys, err := redisClient.GetNodes(c.Copy())
	if err != nil {
		httpLogger.Error(err, "cannot get node list")
		log.Println("cannot get node list. error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if len(keys) == 0 {
		httpLogger.Info("no nodes found in the cluster")
		c.JSON(http.StatusNoContent, gin.H{
			"error": "cannot find any node entries",
		})

		return
	}

	sort.Strings(keys)

	c.JSON(http.StatusOK, gin.H{
		"nodes": keys,
	})
}
