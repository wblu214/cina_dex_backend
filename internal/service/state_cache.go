package service

import (
	"math/big"
	"sync"

	"github.com/cina_dex_backend/internal/model"
)

// StateCache stores latest pool state and price in memory, updated by a background job.
type StateCache struct {
	mu          sync.RWMutex
	poolState   *model.PoolState
	nativePrice *big.Int
}

func NewStateCache() *StateCache {
	return &StateCache{}
}

func (c *StateCache) SetPoolState(s *model.PoolState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.poolState = s
}

func (c *StateCache) GetPoolState() (*model.PoolState, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.poolState == nil {
		return nil, false
	}
	return c.poolState, true
}

func (c *StateCache) SetNativePrice(p *big.Int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if p == nil {
		c.nativePrice = nil
		return
	}
	// store a copy to avoid external mutation
	c.nativePrice = new(big.Int).Set(p)
}

func (c *StateCache) GetNativePrice() (*big.Int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.nativePrice == nil {
		return nil, false
	}
	// return a copy so callers cannot mutate internal state
	return new(big.Int).Set(c.nativePrice), true
}
