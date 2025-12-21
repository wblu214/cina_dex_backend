package handler

import (
	"net/http"

	"github.com/cina_dex_backend/internal/service"
	"github.com/cina_dex_backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// PoolHandler exposes pool-related read APIs.
type PoolHandler struct {
	poolSvc service.PoolService
}

func NewPoolHandler(poolSvc service.PoolService) *PoolHandler {
	return &PoolHandler{poolSvc: poolSvc}
}



// GetPoolState returns aggregated pool state for the frontend dashboard.
func (h *PoolHandler) GetPoolState(c *gin.Context) {
	state, err := h.poolSvc.GetPoolState(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(1001, err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(state))
}
