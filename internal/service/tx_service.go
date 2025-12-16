package service

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/cina_dex_backend/internal/config"
	"github.com/cina_dex_backend/internal/model"
	"github.com/cina_dex_backend/internal/onchain"
	"github.com/ethereum/go-ethereum/common"
)

// TxService builds transaction payloads (to/data/value) for frontend wallets
// to sign and send. It does not hold or use user private keys.
type TxService interface {
	BuildDepositTx(ctx context.Context, amount string) (*model.DepositTx, error)
	BuildBorrowTx(ctx context.Context, amount string, duration uint64, collateralWei string) (*model.BorrowTx, error)
	BuildRepayTx(ctx context.Context, loanID uint64) (*model.RepayTx, error)
	BuildLiquidateTx(ctx context.Context, loanID uint64) (*model.LiquidateTx, error)
}

// txService is the default implementation of TxService.
type txService struct {
	cfg       *config.Config
	client    onchain.Client
	tokenAddr common.Address
	poolAddr  common.Address
}

// Precomputed selectors for write methods.
var (
	selectorERC20Approve = []byte{0x09, 0x5e, 0xa7, 0xb3} // approve(address,uint256)
	selectorDeposit      = []byte{0xb6, 0xb5, 0x5f, 0x25} // deposit(uint256)
	selectorBorrow       = []byte{0x0e, 0xcb, 0xcd, 0xab} // borrow(uint256,uint256)
	selectorRepay        = []byte{0x37, 0x1f, 0xd8, 0xe6} // repay(uint256)
	selectorLiquidate    = []byte{0x41, 0x5f, 0x12, 0x40} // liquidate(uint256)
)

// NewTxService constructs a TxService; it infers the USDT/MockUSDT address
// from the chain config.
func NewTxService(cfg *config.Config, c onchain.Client) (TxService, error) {
	token := cfg.ChainConfig.USDT
	if token == "" {
		token = cfg.ChainConfig.MockUSDT
	}
	if !isHexAddress(token) {
		return nil, fmt.Errorf("invalid token address in chain config: %s", token)
	}
	if !isHexAddress(cfg.ChainConfig.LendingPool) {
		return nil, fmt.Errorf("invalid lendingPool address in chain config: %s", cfg.ChainConfig.LendingPool)
	}

	return &txService{
		cfg:       cfg,
		client:    c,
		tokenAddr: common.HexToAddress(token),
		poolAddr:  common.HexToAddress(cfg.ChainConfig.LendingPool),
	}, nil
}

// BuildDepositTx builds approve + deposit calls given an amount of USDT.
// amount is a decimal string in the token's smallest unit (6 decimals for USDT).
func (s *txService) BuildDepositTx(ctx context.Context, amount string) (*model.DepositTx, error) {
	amt, err := parseBig(amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	approve := buildApproveCall(s.tokenAddr, s.poolAddr, amt)

	// deposit(uint256 amount)
	data := make([]byte, len(selectorDeposit)+32)
	copy(data, selectorDeposit)
	copy(data[len(selectorDeposit):], packUint256(amt))

	deposit := &model.TxCall{
		To:    s.poolAddr.Hex(),
		Data:  bytesToHex(data),
		Value: "0",
	}

	return &model.DepositTx{
		Approve: approve,
		Deposit: deposit,
	}, nil
}

// BuildBorrowTx builds a single borrow call. Collateral is sent as msg.value.
func (s *txService) BuildBorrowTx(ctx context.Context, amount string, duration uint64, collateralWei string) (*model.BorrowTx, error) {
	amt, err := parseBig(amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}
	collateral, err := parseBig(collateralWei)
	if err != nil {
		return nil, fmt.Errorf("invalid collateralWei: %w", err)
	}

	// borrow(uint256 amount, uint256 duration)
	data := make([]byte, len(selectorBorrow)+64)
	copy(data, selectorBorrow)
	copy(data[len(selectorBorrow):], packUint256(amt))

	dur := new(big.Int).SetUint64(duration)
	copy(data[len(selectorBorrow)+32:], packUint256(dur))

	borrow := &model.TxCall{
		To:    s.poolAddr.Hex(),
		Data:  bytesToHex(data),
		Value: collateral.String(),
	}

	return &model.BorrowTx{
		Borrow: borrow,
	}, nil
}

// BuildRepayTx builds approve + repay for a given loanId, using on-chain
// repaymentAmount from loans(loanId).
func (s *txService) BuildRepayTx(ctx context.Context, loanID uint64) (*model.RepayTx, error) {
	loan, err := s.client.GetLoan(ctx, loanID)
	if err != nil {
		return nil, fmt.Errorf("read loan: %w", err)
	}

	repAmount, err := parseBig(loan.RepaymentAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid repaymentAmount on-chain: %w", err)
	}

	approve := buildApproveCall(s.tokenAddr, s.poolAddr, repAmount)

	// repay(uint256 loanId)
	data := make([]byte, len(selectorRepay)+32)
	copy(data, selectorRepay)
	copy(data[len(selectorRepay):], packUint64(loanID))

	repay := &model.TxCall{
		To:    s.poolAddr.Hex(),
		Data:  bytesToHex(data),
		Value: "0",
	}

	return &model.RepayTx{
		Approve: approve,
		Repay:   repay,
	}, nil
}

// BuildLiquidateTx builds approve + liquidate for a given loanId, using the
// current repaymentAmount on-chain as the amount the liquidator needs to pay.
func (s *txService) BuildLiquidateTx(ctx context.Context, loanID uint64) (*model.LiquidateTx, error) {
	loan, err := s.client.GetLoan(ctx, loanID)
	if err != nil {
		return nil, fmt.Errorf("read loan: %w", err)
	}

	repAmount, err := parseBig(loan.RepaymentAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid repaymentAmount on-chain: %w", err)
	}

	approve := buildApproveCall(s.tokenAddr, s.poolAddr, repAmount)

	// liquidate(uint256 loanId)
	data := make([]byte, len(selectorLiquidate)+32)
	copy(data, selectorLiquidate)
	copy(data[len(selectorLiquidate):], packUint64(loanID))

	liq := &model.TxCall{
		To:    s.poolAddr.Hex(),
		Data:  bytesToHex(data),
		Value: "0",
	}

	return &model.LiquidateTx{
		Approve:   approve,
		Liquidate: liq,
	}, nil
}

// buildApproveCall creates an ERC20 approve(spender, amount) TxCall.
func buildApproveCall(token, spender common.Address, amt *big.Int) *model.TxCall {
	data := make([]byte, len(selectorERC20Approve)+64)
	copy(data, selectorERC20Approve)
	copy(data[len(selectorERC20Approve):], packAddress(spender))
	copy(data[len(selectorERC20Approve)+32:], packUint256(amt))

	return &model.TxCall{
		To:    token.Hex(),
		Data:  bytesToHex(data),
		Value: "0",
	}
}

func parseBig(s string) (*big.Int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty string")
	}
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("not a valid decimal: %s", s)
	}
	return v, nil
}

// packUint256 encodes a uint256 as a 32-byte big-endian word.
func packUint256(v *big.Int) []byte {
	word := make([]byte, 32)
	b := v.Bytes()
	copy(word[32-len(b):], b)
	return word
}

// packAddress encodes an address as a 32-byte ABI word.
func packAddress(addr common.Address) []byte {
	word := make([]byte, 32)
	copy(word[12:], addr.Bytes())
	return word
}

// packUint64 encodes a uint64 as a 32-byte ABI uint256.
func packUint64(v uint64) []byte {
	word := make([]byte, 32)
	b := new(big.Int).SetUint64(v).Bytes()
	copy(word[32-len(b):], b)
	return word
}

func bytesToHex(b []byte) string {
	const hex = "0123456789abcdef"
	out := make([]byte, 2+len(b)*2)
	out[0], out[1] = '0', 'x'
	for i, v := range b {
		out[2+i*2] = hex[v>>4]
		out[2+i*2+1] = hex[v&0x0f]
	}
	return string(out)
}

func isHexAddress(s string) bool {
	if len(s) != 42 || !strings.HasPrefix(s, "0x") && !strings.HasPrefix(s, "0X") {
		return false
	}
	return true
}
