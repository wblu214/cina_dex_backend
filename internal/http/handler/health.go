package handler

import (
	"net/http"

	"github.com/cina_dex_backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// Health is a simple liveness endpoint.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"status": "ok",
	}))
}
