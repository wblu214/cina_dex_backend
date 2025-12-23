package service

import (
	"context"
	"fmt"
	"math/big"

	"github.com/cina_dex_backend/internal/model"
	"github.com/cina_dex_backend/internal/onchain"
)

// QuoteService exposes read-only risk / quote related helpers.
type QuoteService interface {
	// QuoteBorrowCollateral computes the required BNB collateral (wei)
	// for a desired USDT borrow amount (6 decimals, as decimal string).
	QuoteBorrowCollateral(ctx context.Context, amount string) (*model.BorrowQuote, error)
}

// quoteService is the default implementation of QuoteService.
type quoteService struct {
	client onchain.Client
	cache  *StateCache
}

// NewQuoteService constructs a QuoteService backed by the on-chain client.
func NewQuoteService(c onchain.Client, cache *StateCache) QuoteService {
	return &quoteService{
		client: c,
		cache:  cache,
	}
}

// Risk parameters mirrored from the LendingPool contract / documentation.
// If these are changed on-chain, they should be updated here accordingly.
const (
	maxLTVPercent = 75 // 75%
)

// QuoteBorrowCollateral computes the required BNB collateral for a given borrow amount.
// - amount: USDT principal in smallest units (6 decimals), as a decimal string.
// The calculation uses:
//
//	collateralWei >= amountUsd / LTV / priceBnbUsd
//
// where:
//   - amountUsd: 18-decimals USD value of the borrow amount (by scaling 6 -> 18)
//   - LTV: maxLTVPercent%
//   - priceBnbUsd: BNB/USD price from ChainlinkOracle.getPrice(address(0)), 18 decimals.
func (s *quoteService) QuoteBorrowCollateral(ctx context.Context, amount string) (*model.BorrowQuote, error) {
	if amount == "" {
		return nil, fmt.Errorf("amount is required")
	}

	amt, err := parseBig(amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}
	if amt.Sign() <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	var (
		price *big.Int
		err   error
	)

	// Prefer cached price if available, fall back to on-chain call.
	if s.cache != nil {
		if p, ok := s.cache.GetNativePrice(); ok {
			price = p
		}
	}
	if price == nil {
		price, err = s.client.GetNativePrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("get native price: %w", err)
		}
	}
	if price.Sign() <= 0 {
		return nil, fmt.Errorf("oracle returned non-positive price")
	}

	// Convert 6-decimal USDT amount to 18-decimal USD amount.
	ten := big.NewInt(10)
	scaleTo18 := new(big.Int).Exp(ten, big.NewInt(12), nil) // 1e12
	amountUSD := new(big.Int).Mul(amt, scaleTo18)

	// We want: collateralWei >= amountUSD / LTV / price
	// Use integers with:
	//   LTV = maxLTVPercent / 100
	// => collateralWei >= amountUSD * 1e18 * 100 / (maxLTVPercent * price)
	oneEth := new(big.Int).Exp(ten, big.NewInt(18), nil) // 1e18

	numerator := new(big.Int).Mul(amountUSD, oneEth)
	numerator.Mul(numerator, big.NewInt(100)) // * 100

	denominator := new(big.Int).Mul(price, big.NewInt(maxLTVPercent))

	if denominator.Sign() == 0 {
		return nil, fmt.Errorf("invalid parameters: zero denominator")
	}

	// ceil(numerator / denominator)
	quotient, remainder := new(big.Int).QuoRem(numerator, denominator, new(big.Int))
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}

	return &model.BorrowQuote{
		BorrowAmount:  amt.String(),
		CollateralWei: quotient.String(),
		BnbUsdPrice:   price.String(),
		MaxLTVPercent: fmt.Sprintf("%d", maxLTVPercent),
	}, nil
}
