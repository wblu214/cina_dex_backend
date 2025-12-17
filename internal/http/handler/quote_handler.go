package handler

import (
	"net/http"

	"github.com/cina_dex_backend/internal/service"
	"github.com/cina_dex_backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// QuoteHandler exposes endpoints that help frontend calculate collateral, etc.
type QuoteHandler struct {
	quoteSvc service.QuoteService
}

func NewQuoteHandler(quoteSvc service.QuoteService) *QuoteHandler {
	return &QuoteHandler{quoteSvc: quoteSvc}
}

type borrowQuoteRequest struct {
	// Amount is the desired USDT borrow amount in smallest units (6 decimals).
	Amount string `json:"amount" binding:"required"`
}

// QuoteBorrow computes the required BNB collateral (wei) for a given borrow amount.
// 它会调用测试网的价格预言机（ChainlinkOracle.getPrice(address(0))）拿到 BNB/USD 价格，
// 再结合 Max LTV（目前 75%）给出需要抵押的 BNB 数量。
func (h *QuoteHandler) QuoteBorrow(c *gin.Context) {
	var req borrowQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(4001, err.Error()))
		return
	}

	quote, err := h.quoteSvc.QuoteBorrowCollateral(c.Request.Context(), req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(1001, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(quote))
}
