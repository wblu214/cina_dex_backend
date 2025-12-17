package onchain

import (
	"context"
	"math/big"

	"github.com/cina_dex_backend/internal/model"
)

// Client abstracts read access to the on-chain lending protocol.
// A concrete implementation should wrap go-ethereum bindings generated from
// the ABI files in go_back/abi.
type Client interface {
	GetPoolState(ctx context.Context) (*model.PoolState, error)
	GetUserPosition(ctx context.Context, address string) (*model.UserPosition, error)
	ListUserLoans(ctx context.Context, address string) ([]*model.Loan, error)
	GetLoan(ctx context.Context, id uint64) (*model.Loan, error)
	GetLoanHealth(ctx context.Context, id uint64) (*model.LoanHealth, error)
	// GetNativePrice returns the BNB/USD price with 18 decimals from ChainlinkOracle.getPrice(address(0)).
	GetNativePrice(ctx context.Context) (*big.Int, error)
}
