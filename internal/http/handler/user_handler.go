package handler

import (
	"net/http"

	"github.com/cina_dex_backend/internal/service"
	"github.com/cina_dex_backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// UserHandler exposes user-centric read APIs.
type UserHandler struct {
	poolSvc service.PoolService
	loanSvc service.LoanService
}

func NewUserHandler(poolSvc service.PoolService, loanSvc service.LoanService) *UserHandler {
	return &UserHandler{
		poolSvc: poolSvc,
		loanSvc: loanSvc,
	}
}

// GetUserPosition returns aggregated principal/repayment/collateral info for a user.
func (h *UserHandler) GetUserPosition(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, response.Error(4001, "address is required"))
		return
	}

	pos, err := h.poolSvc.GetUserPosition(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(1001, err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(pos))
}

// ListUserLoans returns all loans for a given user address.
func (h *UserHandler) ListUserLoans(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, response.Error(4001, "address is required"))
		return
	}

	loans, err := h.loanSvc.ListUserLoans(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(1001, err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(loans))
}
