package main

import (
	"log"

	"github.com/cina_dex_backend/internal/config"
	apihttp "github.com/cina_dex_backend/internal/http"
	"github.com/cina_dex_backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	poolSvc := service.NewPoolService()
	loanSvc := service.NewLoanService()

	r := apihttp.NewRouter(cfg, poolSvc, loanSvc)

	addr := ":" + cfg.HTTPPort
	log.Printf("starting API server on %s (env=%s, chain=%s)", addr, cfg.Env, cfg.ChainEnv)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
