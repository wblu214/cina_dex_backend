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
