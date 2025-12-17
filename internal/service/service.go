package service

import (
	"context"

	"github.com/cina_dex_backend/internal/model"
	"github.com/cina_dex_backend/internal/onchain"
)

// PoolService defines read operations related to the lending pool.
type PoolService interface {
	GetPoolState(ctx context.Context) (*model.PoolState, error)
	GetUserPosition(ctx context.Context, address string) (*model.UserPosition, error)
	// GetLenderPosition returns LP position and earnings info for a given address.
	GetLenderPosition(ctx context.Context, address string) (*model.LenderPosition, error)
}

// LoanService defines operations related to individual loans.
type LoanService interface {
	ListUserLoans(ctx context.Context, address string) ([]*model.Loan, error)
	GetLoan(ctx context.Context, id uint64) (*model.Loan, error)
	GetLoanHealth(ctx context.Context, id uint64) (*model.LoanHealth, error)
}

// NewPoolService constructs a PoolService backed by the on-chain client.
func NewPoolService(c onchain.Client) PoolService {
	return &poolService{client: c}
}

// NewLoanService constructs a LoanService backed by the on-chain client.
func NewLoanService(c onchain.Client) LoanService {
	return &loanService{client: c}
}

type poolService struct {
	client onchain.Client
}

func (s *poolService) GetPoolState(ctx context.Context) (*model.PoolState, error) {
	return s.client.GetPoolState(ctx)
}

func (s *poolService) GetUserPosition(ctx context.Context, address string) (*model.UserPosition, error) {
	return s.client.GetUserPosition(ctx, address)
}

func (s *poolService) GetLenderPosition(ctx context.Context, address string) (*model.LenderPosition, error) {
	lp, err := s.client.GetLenderPosition(ctx, address)
	if err != nil {
		return nil, err
	}

	// TODO: replace this stub with real off-chain tracking of netDeposited (total deposits - total withdrawals).
	lp.NetDeposited = "0"
	lp.Interest = lp.UnderlyingBalance // until netDeposited is wired, treat全部为收益的上界展示

	return lp, nil
}

type loanService struct {
	client onchain.Client
}

func (s *loanService) ListUserLoans(ctx context.Context, address string) ([]*model.Loan, error) {
	return s.client.ListUserLoans(ctx, address)
}

func (s *loanService) GetLoan(ctx context.Context, id uint64) (*model.Loan, error) {
	return s.client.GetLoan(ctx, id)
}

func (s *loanService) GetLoanHealth(ctx context.Context, id uint64) (*model.LoanHealth, error) {
	return s.client.GetLoanHealth(ctx, id)
}
