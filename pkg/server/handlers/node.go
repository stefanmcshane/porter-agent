package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"

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
	nodes := make(map[string]bool)

	for _, key := range keys {
		k := strings.Split(key, ":")
		stripped := strings.Join(k[1:len(k)-1], ":")
		nodes[stripped] = true
	}

	strippedNodes := []string{}
	for k := range nodes {
		strippedNodes = append(strippedNodes, k)
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": strippedNodes,
	})
}

func ListNodeHistory(c *gin.Context) {
	nodeName := c.Param("node")

	keys, err := redisClient.GetNodeHistory(c.Copy(), nodeName)
	if err != nil {
		httpLogger.Error(err, "cannot get node event history")
		log.Println("unable to get node history", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if len(keys) == 0 {
		httpLogger.Info("no event history found for the given node")
		c.JSON(http.StatusNoContent, gin.H{
			"error": "cannot find event history for given node",
		})

		return
	}

	timestamps := []string{}
	for _, k := range keys {
		ts := strings.Split(k, ":")
		timestamps = append(timestamps, ts[2:]...)
	}

	c.JSON(http.StatusOK, gin.H{
		"node":    nodeName,
		"history": timestamps,
	})
}

func GetNodeCondition(c *gin.Context) {
	nodeName := c.Param("node")
	timestamp := c.Param("timestamp")

	rawConditions, err := redisClient.GetNodeCondition(c.Copy(), nodeName, timestamp)
	if err != nil {
		httpLogger.Error(err, "cannot get node condition")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if len(rawConditions) == 0 {
		httpLogger.Info("no conditions found for the given node and timestamp")
		c.JSON(http.StatusNoContent, gin.H{})

		return
	}

	var conditions interface{}

	err = json.Unmarshal([]byte(rawConditions[0]), &conditions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conditions": conditions,
	})
}
