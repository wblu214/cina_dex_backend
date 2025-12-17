package model

// PoolState mirrors LendingPool.getPoolState().
// All numeric fields are encoded as decimal strings to avoid precision loss.
type PoolState struct {
	TotalAssets        string `json:"totalAssets"`
	TotalBorrowed      string `json:"totalBorrowed"`
	AvailableLiquidity string `json:"availableLiquidity"`
	ExchangeRate       string `json:"exchangeRate"`
	TotalFTokenSupply  string `json:"totalFTokenSupply"`
}

// Loan represents a single on-chain loan position.
type Loan struct {
	ID               uint64 `json:"id"`
	Borrower         string `json:"borrower"`
	CollateralAmount string `json:"collateralAmount"`
	Principal        string `json:"principal"`
	RepaymentAmount  string `json:"repaymentAmount"`
	StartTime        uint64 `json:"startTime"`
	Duration         uint64 `json:"duration"`
	IsActive         bool   `json:"isActive"`
}

// LoanHealth is derived from getLoanHealth.
type LoanHealth struct {
	LTV            string `json:"ltv"`
	IsLiquidatable bool   `json:"isLiquidatable"`
}

// UserPosition mirrors LendingPool.getUserPosition().
type UserPosition struct {
	Address         string   `json:"address"`
	LoanIDs         []uint64 `json:"loanIds"`
	TotalPrincipal  string   `json:"totalPrincipal"`
	TotalRepayment  string   `json:"totalRepayment"`
	TotalCollateral string   `json:"totalCollateral"`
}

// TxCall describes a single Ethereum transaction for the frontend to sign.
type TxCall struct {
	To    string `json:"to"`
	Data  string `json:"data"`
	Value string `json:"value"`
}

// DepositTx bundles the approve + deposit calls needed for a deposit flow.
type DepositTx struct {
	Approve *TxCall `json:"approve"`
	Deposit *TxCall `json:"deposit"`
}

// BorrowTx contains the single borrow call (BNB as msg.value).
type BorrowTx struct {
	Borrow *TxCall `json:"borrow"`
}

// RepayTx bundles the approve + repay calls.
type RepayTx struct {
	Approve *TxCall `json:"approve"`
	Repay   *TxCall `json:"repay"`
}

// LiquidateTx bundles the approve + liquidate calls.
type LiquidateTx struct {
	Approve   *TxCall `json:"approve"`
	Liquidate *TxCall `json:"liquidate"`
}

// WithdrawTx contains the single withdraw call for LP redemptions.
type WithdrawTx struct {
	Withdraw *TxCall `json:"withdraw"`
}

// LenderPosition mirrors LendingPool.getLenderPosition(address).
// All numeric fields are encoded as decimal strings.
type LenderPosition struct {
	Address           string `json:"address"`
	FTokenBalance     string `json:"fTokenBalance"`
	ExchangeRate      string `json:"exchangeRate"`
	UnderlyingBalance string `json:"underlyingBalance"`
	// NetDeposited is the user's historical net deposit amount (total deposits - total withdrawals),
	// denominated in USDT smallest units (6 decimals). It is tracked off-chain.
	NetDeposited string `json:"netDeposited"`
	// Interest is the current realized interest = underlyingBalance - netDeposited.
	Interest string `json:"interest"`
}

// BorrowQuote describes the required collateral for a desired borrow amount.
// It is computed off-chain using the on-chain price oracle and risk parameters.
type BorrowQuote struct {
	// BorrowAmount is the requested USDT principal in smallest units (6 decimals).
	BorrowAmount string `json:"borrowAmount"`
	// CollateralWei is the required BNB collateral in wei (18 decimals).
	CollateralWei string `json:"collateralWei"`
	// BnbUsdPrice is the BNB/USD price used for the quote, 18 decimals.
	BnbUsdPrice string `json:"bnbUsdPrice"`
	// MaxLTVPercent is the max LTV used in this quote, e.g. "75" for 75%.
	MaxLTVPercent string `json:"maxLtvPercent"`
}
