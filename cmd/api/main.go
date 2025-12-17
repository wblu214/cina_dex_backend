package main

import (
	"context"
	"log"

	"github.com/cina_dex_backend/internal/config"
	apihttp "github.com/cina_dex_backend/internal/http"
	"github.com/cina_dex_backend/internal/onchain"
	"github.com/cina_dex_backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()

	chainClient, err := onchain.NewEthClient(ctx, cfg)
	if err != nil {
		log.Fatalf("init on-chain client: %v", err)
	}

	poolSvc := service.NewPoolService(chainClient)
	loanSvc := service.NewLoanService(chainClient)
	quoteSvc := service.NewQuoteService(chainClient)
	txSvc, err := service.NewTxService(cfg, chainClient)
	if err != nil {
		log.Fatalf("init tx service: %v", err)
	}

	r := apihttp.NewRouter(cfg, poolSvc, loanSvc, txSvc, quoteSvc)

	addr := ":" + cfg.HTTPPort
	log.Printf("starting API server on %s (env=%s, chain=%s)", addr, cfg.Env, cfg.ChainEnv)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
