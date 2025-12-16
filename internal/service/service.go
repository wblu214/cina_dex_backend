package service

import (
	"context"
	"errors"

	"github.com/cina_dex_backend/internal/model"
)

// PoolService defines read operations related to the lending pool.
type PoolService interface {
	GetPoolState(ctx context.Context) (*model.PoolState, error)
	GetUserPosition(ctx context.Context, address string) (*model.UserPosition, error)
}

// LoanService defines operations related to individual loans.
type LoanService interface {
	ListUserLoans(ctx context.Context, address string) ([]*model.Loan, error)
	GetLoan(ctx context.Context, id uint64) (*model.Loan, error)
	GetLoanHealth(ctx context.Context, id uint64) (*model.LoanHealth, error)
}

// NewPoolService returns a default implementation placeholder.
// TODO: wire this to the on-chain client once implemented.
func NewPoolService() PoolService {
	return &noopPoolService{}
}

// NewLoanService returns a default implementation placeholder.
// TODO: wire this to the on-chain client once implemented.
func NewLoanService() LoanService {
	return &noopLoanService{}
}

type noopPoolService struct{}

func (s *noopPoolService) GetPoolState(ctx context.Context) (*model.PoolState, error) {
	return nil, errors.New("GetPoolState not implemented")
}

func (s *noopPoolService) GetUserPosition(ctx context.Context, address string) (*model.UserPosition, error) {
	return nil, errors.New("GetUserPosition not implemented")
}

type noopLoanService struct{}

func (s *noopLoanService) ListUserLoans(ctx context.Context, address string) ([]*model.Loan, error) {
	return nil, errors.New("ListUserLoans not implemented")
}

func (s *noopLoanService) GetLoan(ctx context.Context, id uint64) (*model.Loan, error) {
	return nil, errors.New("GetLoan not implemented")
}

func (s *noopLoanService) GetLoanHealth(ctx context.Context, id uint64) (*model.LoanHealth, error) {
	return nil, errors.New("GetLoanHealth not implemented")
}
