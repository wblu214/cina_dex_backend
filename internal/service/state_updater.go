package service

import (
	"context"
	"log"
	"time"

	"github.com/cina_dex_backend/internal/onchain"
)

// StartStateUpdater launches a background goroutine that periodically refreshes
// pool state and native price into the given cache.
func StartStateUpdater(ctx context.Context, client onchain.Client, cache *StateCache, interval time.Duration) {
	if cache == nil {
		return
	}

	go func() {
		// initial run
		refreshOnce(ctx, client, cache)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("state updater stopped: context cancelled")
				return
			case <-ticker.C:
				refreshOnce(ctx, client, cache)
			}
		}
	}()
}

func refreshOnce(ctx context.Context, client onchain.Client, cache *StateCache) {
	if cache == nil {
		return
	}

	if ps, err := client.GetPoolState(ctx); err != nil {
		log.Printf("state updater: get pool state: %v", err)
	} else {
		cache.SetPoolState(ps)
	}

	if price, err := client.GetNativePrice(ctx); err != nil {
		log.Printf("state updater: get native price: %v", err)
	} else {
		cache.SetNativePrice(price)
	}
}
