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

	r := apihttp.NewRouter(cfg, poolSvc, loanSvc)

	addr := ":" + cfg.HTTPPort
	log.Printf("starting API server on %s (env=%s, chain=%s)", addr, cfg.Env, cfg.ChainEnv)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
