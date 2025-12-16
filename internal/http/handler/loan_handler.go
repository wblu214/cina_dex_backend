package handler

import (
	"net/http"
	"strconv"

	"github.com/cina_dex_backend/internal/service"
	"github.com/cina_dex_backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// LoanHandler exposes loan-level read APIs.
type LoanHandler struct {
	loanSvc service.LoanService
}

func NewLoanHandler(loanSvc service.LoanService) *LoanHandler {
	return &LoanHandler{loanSvc: loanSvc}
}

// GetLoan returns details for a specific loan.
func (h *LoanHandler) GetLoan(c *gin.Context) {
	loanID, ok := parseLoanID(c)
	if !ok {
		return
	}

	loan, err := h.loanSvc.GetLoan(c.Request.Context(), loanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(1001, err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(loan))
}

// GetLoanHealth returns LTV and liquidation status for a loan.
func (h *LoanHandler) GetLoanHealth(c *gin.Context) {
	loanID, ok := parseLoanID(c)
	if !ok {
		return
	}

	health, err := h.loanSvc.GetLoanHealth(c.Request.Context(), loanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(1001, err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(health))
}

func parseLoanID(c *gin.Context) (uint64, bool) {
	raw := c.Param("loanId")
	if raw == "" {
		c.JSON(http.StatusBadRequest, response.Error(4001, "loanId is required"))
		return 0, false
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(4002, "loanId must be a uint"))
		return 0, false
	}
	return id, true
}
