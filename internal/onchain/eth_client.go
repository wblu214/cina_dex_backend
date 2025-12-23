package onchain

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/cina_dex_backend/internal/config"
	"github.com/cina_dex_backend/internal/model"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthClient is a lightweight on-chain client that talks directly to the
// LendingPool contract using precomputed function selectors.
//
// NOTE: The current implementation assumes that the LendingPool ABI matches
// the description in go_back/docs/api.md. It manually encodes/decodes a
// small subset of view functions needed by the REST API.
type EthClient struct {
	rpc         *ethclient.Client
	lendingPool common.Address
	oracle      common.Address
}

// NewEthClient dials the configured RPC endpoint and prepares a client that
// can read from the on-chain LendingPool contract.
func NewEthClient(ctx context.Context, cfg *config.Config) (*EthClient, error) {
	if cfg.ChainConfig.LendingPool == "" {
		return nil, fmt.Errorf("missing lendingPool address in chain config")
	}

	rpc, err := ethclient.DialContext(ctx, cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("dial rpc: %w", err)
	}

	client := &EthClient{
		rpc:         rpc,
		lendingPool: common.HexToAddress(cfg.ChainConfig.LendingPool),
	}

	// Oracle address is optional; some environments (e.g. mainnet during early setup)
	// may not have it configured yet.
	if cfg.ChainConfig.ChainlinkOracle != "" && isHexAddress(cfg.ChainConfig.ChainlinkOracle) {
		client.oracle = common.HexToAddress(cfg.ChainConfig.ChainlinkOracle)
	}

	return client, nil
}

// Ensure EthClient implements Client.
var _ Client = (*EthClient)(nil)

// Precomputed 4-byte selectors for the view functions we need.
var (
	selectorGetPoolState    = []byte{0x21, 0x7a, 0xc2, 0x37} // getPoolState()
	selectorGetUserPosition = []byte{0x5b, 0x7c, 0x2d, 0xad} // getUserPosition(address)
	selectorGetUserLoans    = []byte{0x02, 0xbf, 0x32, 0x1f} // getUserLoans(address)
	selectorGetLoanHealth   = []byte{0xb6, 0xe0, 0x76, 0x88} // getLoanHealth(uint256)
	selectorLoans           = []byte{0xe1, 0xec, 0x3c, 0x68} // loans(uint256)
	selectorGetLenderPos    = []byte{0x5d, 0x41, 0x3f, 0xa2} // getLenderPosition(address)
	selectorGetPrice        = []byte{0x41, 0x97, 0x6e, 0x09} // getPrice(address)
)

// GetPoolState calls LendingPool.getPoolState() and maps the result to model.PoolState.
func (c *EthClient) GetPoolState(ctx context.Context) (*model.PoolState, error) {
	data := make([]byte, len(selectorGetPoolState))
	copy(data, selectorGetPoolState)

	out, err := c.call(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("call getPoolState: %w", err)
	}

	// Expect 5 * 32 bytes: (totalAssets, totalBorrowed, availableLiquidity, exchangeRate, totalFTokenSupply)
	words, err := splitWords(out, 5)
	if err != nil {
		return nil, fmt.Errorf("decode getPoolState: %w", err)
	}

	return &model.PoolState{
		TotalAssets:        words[0].String(),
		TotalBorrowed:      words[1].String(),
		AvailableLiquidity: words[2].String(),
		ExchangeRate:       words[3].String(),
		TotalFTokenSupply:  words[4].String(),
	}, nil
}

func (c *EthClient) GetUserPosition(ctx context.Context, address string) (*model.UserPosition, error) {
	addr := common.HexToAddress(address)

	data := make([]byte, len(selectorGetUserPosition)+32)
	copy(data, selectorGetUserPosition)
	copy(data[len(selectorGetUserPosition):], packAddress(addr))

	out, err := c.call(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("call getUserPosition: %w", err)
	}

	// Returns (uint256[] loanIds, uint256 totalPrincipal, uint256 totalRepayment, uint256 totalCollateral)
	if len(out) < 4*32 {
		return nil, fmt.Errorf("getUserPosition output too short")
	}

	offset := new(big.Int).SetBytes(out[0:32]).Int64()
	if offset < 0 || int(offset)+32 > len(out) {
		return nil, fmt.Errorf("invalid loanIds offset %d", offset)
	}

	totalPrincipal := new(big.Int).SetBytes(out[32:64])
	totalRepayment := new(big.Int).SetBytes(out[64:96])
	totalCollateral := new(big.Int).SetBytes(out[96:128])

	loanIDs, err := decodeUint256Array(out, int(offset))
	if err != nil {
		return nil, fmt.Errorf("decode loanIds: %w", err)
	}

	return &model.UserPosition{
		Address:         addr.Hex(),
		LoanIDs:         loanIDs,
		TotalPrincipal:  totalPrincipal.String(),
		TotalRepayment:  totalRepayment.String(),
		TotalCollateral: totalCollateral.String(),
	}, nil
}

// GetLenderPosition calls LendingPool.getLenderPosition(address).
func (c *EthClient) GetLenderPosition(ctx context.Context, address string) (*model.LenderPosition, error) {
	addr := common.HexToAddress(address)

	data := make([]byte, len(selectorGetLenderPos)+32)
	copy(data, selectorGetLenderPos)
	copy(data[len(selectorGetLenderPos):], packAddress(addr))

	out, err := c.call(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("call getLenderPosition: %w", err)
	}

	// (uint256 fTokenBalance, uint256 exchangeRate, uint256 underlyingBalance)
	words, err := splitWords(out, 3)
	if err != nil {
		return nil, fmt.Errorf("decode getLenderPosition: %w", err)
	}

	return &model.LenderPosition{
		Address:           addr.Hex(),
		FTokenBalance:     words[0].String(),
		ExchangeRate:      words[1].String(),
		UnderlyingBalance: words[2].String(),
		NetDeposited:      "0", // filled by service layer if off-chain tracking is implemented
		Interest:          "0", // filled by service layer
	}, nil
}

// ListUserLoans calls getUserLoans(address) to get IDs then loans(id) for each.
func (c *EthClient) ListUserLoans(ctx context.Context, address string) ([]*model.Loan, error) {
	addr := common.HexToAddress(address)

	data := make([]byte, len(selectorGetUserLoans)+32)
	copy(data, selectorGetUserLoans)
	copy(data[len(selectorGetUserLoans):], packAddress(addr))

	out, err := c.call(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("call getUserLoans: %w", err)
	}

	if len(out) < 32 {
		return nil, fmt.Errorf("getUserLoans output too short")
	}

	offset := new(big.Int).SetBytes(out[0:32]).Int64()
	if offset < 0 || int(offset)+32 > len(out) {
		return nil, fmt.Errorf("invalid loanIds offset %d", offset)
	}

	ids, err := decodeUint256Array(out, int(offset))
	if err != nil {
		return nil, fmt.Errorf("decode getUserLoans: %w", err)
	}

	loans := make([]*model.Loan, 0, len(ids))
	for _, id := range ids {
		loan, err := c.GetLoan(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get loan %d: %w", id, err)
		}
		loans = append(loans, loan)
	}

	return loans, nil
}

// GetLoan calls loans(uint256) and maps to model.Loan.
func (c *EthClient) GetLoan(ctx context.Context, id uint64) (*model.Loan, error) {
	data := make([]byte, len(selectorLoans)+32)
	copy(data, selectorLoans)
	copy(data[len(selectorLoans):], packUint64(id))

	out, err := c.call(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("call loans(%d): %w", id, err)
	}

	// (address borrower, uint256 collateralAmount, uint256 principal,
	//  uint256 repaymentAmount, uint256 startTime, uint256 duration, bool isActive)
	if len(out) < 7*32 {
		return nil, fmt.Errorf("loans output too short")
	}

	borrower := common.BytesToAddress(out[12:32]).Hex()
	collateral := new(big.Int).SetBytes(out[32:64])
	principal := new(big.Int).SetBytes(out[64:96])
	repayment := new(big.Int).SetBytes(out[96:128])
	startTime := new(big.Int).SetBytes(out[128:160])
	duration := new(big.Int).SetBytes(out[160:192])
	isActive := out[192+31] == 1

	return &model.Loan{
		ID:               id,
		Borrower:         borrower,
		CollateralAmount: collateral.String(),
		Principal:        principal.String(),
		RepaymentAmount:  repayment.String(),
		StartTime:        startTime.Uint64(),
		Duration:         duration.Uint64(),
		IsActive:         isActive,
	}, nil
}

// GetLoanHealth calls getLoanHealth(uint256).
func (c *EthClient) GetLoanHealth(ctx context.Context, id uint64) (*model.LoanHealth, error) {
	data := make([]byte, len(selectorGetLoanHealth)+32)
	copy(data, selectorGetLoanHealth)
	copy(data[len(selectorGetLoanHealth):], packUint64(id))

	out, err := c.call(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("call getLoanHealth(%d): %w", id, err)
	}

	if len(out) < 2*32 {
		return nil, fmt.Errorf("getLoanHealth output too short")
	}

	ltv := new(big.Int).SetBytes(out[0:32])
	isLiquidatable := out[32+31] == 1

	return &model.LoanHealth{
		LTV:            ltv.String(),
		IsLiquidatable: isLiquidatable,
	}, nil
}

// GetNativePrice reads BNB/USD price from ChainlinkOracle.getPrice(address(0)).
// It returns a uint256 with 18 decimals, for example 2000e18 for $2000.
func (c *EthClient) GetNativePrice(ctx context.Context) (*big.Int, error) {
	if (c.oracle == common.Address{}) {
		return nil, fmt.Errorf("oracle address not configured")
	}

	// getPrice(address(0))
	data := make([]byte, len(selectorGetPrice)+32)
	copy(data, selectorGetPrice)
	// asset = address(0) encoded as a 32-byte word.
	var zero common.Address
	copy(data[len(selectorGetPrice):], packAddress(zero))

	msg := ethereum.CallMsg{
		To:   &c.oracle,
		Data: data,
	}

	out, err := c.rpc.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("call getPrice(address(0)): %w", err)
	}
	if len(out) < 32 {
		return nil, fmt.Errorf("getPrice output too short")
	}

	price := new(big.Int).SetBytes(out[0:32])
	return price, nil
}

// call executes a read-only call against the LendingPool contract.
func (c *EthClient) call(ctx context.Context, data []byte) ([]byte, error) {
	msg := ethereum.CallMsg{
		To:   &c.lendingPool,
		Data: data,
	}
	return c.rpc.CallContract(ctx, msg, nil)
}

// splitWords splits ABI-encoded static return data into N uint256 words.
func splitWords(data []byte, n int) ([]*big.Int, error) {
	if len(data) < n*32 {
		return nil, fmt.Errorf("need at least %d bytes, got %d", n*32, len(data))
	}
	words := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		start := i * 32
		end := start + 32
		words[i] = new(big.Int).SetBytes(data[start:end])
	}
	return words, nil
}

// decodeUint256Array decodes a dynamic uint256[] whose length word starts at the given offset.
func decodeUint256Array(data []byte, offset int) ([]uint64, error) {
	if offset < 0 || offset+32 > len(data) {
		return nil, fmt.Errorf("invalid offset %d", offset)
	}

	length := new(big.Int).SetBytes(data[offset : offset+32]).Int64()
	if length < 0 {
		return nil, fmt.Errorf("negative length %d", length)
	}

	total := offset + 32 + int(length)*32
	if total > len(data) {
		return nil, fmt.Errorf("array length %d exceeds data size", length)
	}

	res := make([]uint64, length)
	for i := int64(0); i < length; i++ {
		start := offset + 32 + int(i)*32
		end := start + 32
		v := new(big.Int).SetBytes(data[start:end])
		res[i] = v.Uint64()
	}
	return res, nil
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

// isHexAddress performs a minimal check for an Ethereum hex address string.
func isHexAddress(s string) bool {
	if len(s) != 42 {
		return false
	}
	if !strings.HasPrefix(s, "0x") && !strings.HasPrefix(s, "0X") {
		return false
	}
	return true
}
