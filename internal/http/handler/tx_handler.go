package handler

import (
	"net/http"

	"github.com/cina_dex_backend/internal/service"
	"github.com/cina_dex_backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// TxHandler exposes endpoints that build transaction payloads
// (to/data/value) for frontend wallets.
type TxHandler struct {
	txSvc service.TxService
}

func NewTxHandler(txSvc service.TxService) *TxHandler {
	return &TxHandler{txSvc: txSvc}
}

type depositTxRequest struct {
	UserAddress string `json:"userAddress"` // not used in backend, kept for symmetry
	Amount      string `json:"amount" binding:"required"`
}

type borrowTxRequest struct {
	UserAddress   string `json:"userAddress"`
	Amount        string `json:"amount" binding:"required"`
	Duration      uint64 `json:"duration" binding:"required"`
	CollateralWei string `json:"collateralWei" binding:"required"`
}

type repayTxRequest struct {
	UserAddress string `json:"userAddress"`
	LoanID      uint64 `json:"loanId" binding:"required"`
}

type liquidateTxRequest struct {
	UserAddress string `json:"userAddress"`
	LoanID      uint64 `json:"loanId" binding:"required"`
}

// BuildDeposit builds approve + deposit tx for a given USDT amount.
func (h *TxHandler) BuildDeposit(c *gin.Context) {
	var req depositTxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(4001, err.Error()))
		return
	}

	tx, err := h.txSvc.BuildDepositTx(c.Request.Context(), req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(1001, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(tx))
}

// BuildBorrow builds a borrow tx with BNB as collateral (msg.value).
func (h *TxHandler) BuildBorrow(c *gin.Context) {
	var req borrowTxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(4001, err.Error()))
		return
	}

	tx, err := h.txSvc.BuildBorrowTx(c.Request.Context(), req.Amount, req.Duration, req.CollateralWei)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(1001, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(tx))
}

// BuildRepay builds approve + repay tx for a loanId.
func (h *TxHandler) BuildRepay(c *gin.Context) {
	var req repayTxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(4001, err.Error()))
		return
	}

	tx, err := h.txSvc.BuildRepayTx(c.Request.Context(), req.LoanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(1001, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(tx))
}

// BuildLiquidate builds approve + liquidate tx for a loanId.
func (h *TxHandler) BuildLiquidate(c *gin.Context) {
	var req liquidateTxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(4001, err.Error()))
		return
	}

	tx, err := h.txSvc.BuildLiquidateTx(c.Request.Context(), req.LoanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(1001, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(tx))
}
