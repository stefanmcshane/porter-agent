package incident

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/porter-dev/porter-agent/api/server/types"
)

func ListIncidents(c *gin.Context) {
	req := &types.ListIncidentsRequest{}

	err := c.BindQuery(req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}
