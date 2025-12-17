package http

import (
	"github.com/cina_dex_backend/internal/config"
	"github.com/cina_dex_backend/internal/http/handler"
	"github.com/cina_dex_backend/internal/service"
	"github.com/gin-gonic/gin"
)

// NewRouter wires routes, handlers, and middlewares.
func NewRouter(cfg *config.Config, poolSvc service.PoolService, loanSvc service.LoanService, txSvc service.TxService, quoteSvc service.QuoteService) *gin.Engine {
	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	poolHandler := handler.NewPoolHandler(poolSvc)
	userHandler := handler.NewUserHandler(poolSvc, loanSvc)
	loanHandler := handler.NewLoanHandler(loanSvc)
	txHandler := handler.NewTxHandler(txSvc)
	quoteHandler := handler.NewQuoteHandler(quoteSvc)

	api := r.Group("/api/v1")
	{
		api.GET("/health", handler.Health)

		api.GET("/pool/state", poolHandler.GetPoolState)

		api.GET("/users/:address/position", userHandler.GetUserPosition)
		api.GET("/users/:address/lender-position", userHandler.GetLenderPosition)
		api.GET("/users/:address/loans", userHandler.ListUserLoans)

		api.GET("/loans/:loanId", loanHandler.GetLoan)
		api.GET("/loans/:loanId/health", loanHandler.GetLoanHealth)

		// risk / quote endpoints
		api.POST("/borrow/quote", quoteHandler.QuoteBorrow)

		// transaction building endpoints
		api.POST("/tx/deposit", txHandler.BuildDeposit)
		api.POST("/tx/withdraw", txHandler.BuildWithdraw)
		api.POST("/tx/borrow", txHandler.BuildBorrow)
		api.POST("/tx/repay", txHandler.BuildRepay)
		api.POST("/tx/liquidate", txHandler.BuildLiquidate)
		// testnet faucet: build MockUSDT mint tx (owner signs & sends on frontend)
		api.POST("/tx/mock-usdt/mint", txHandler.BuildMintMockUSDT)
	}

	// Swagger UI & OpenAPI spec
	r.GET("/swagger", handler.SwaggerUI)
	r.GET("/swagger/openapi.json", handler.SwaggerSpec)

	return r
}
