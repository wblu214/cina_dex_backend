package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// ChainConfig holds on-chain addresses for a specific network.
type ChainConfig struct {
	ChainID         int64  `json:"chainId"`
	RPCUrlEnv       string `json:"rpcUrlEnv"`
	MockUSDT        string `json:"mockUsdt,omitempty"`
	USDT            string `json:"usdt,omitempty"`
	PriceFeed       string `json:"priceFeed"`
	ChainlinkOracle string `json:"chainlinkOracle"`
	FToken          string `json:"fToken"`
	LendingPool     string `json:"lendingPool"`
}

// Addresses represents the address book loaded from go_back/addresses.json.
type Addresses struct {
	BSCTestnet ChainConfig `json:"bscTestnet"`
	BSCMainnet ChainConfig `json:"bscMainnet"`
}

// Config is the top-level application configuration.
type Config struct {
	Env         string
	HTTPPort    string
	ChainEnv    string
	RPCURL      string
	ChainConfig ChainConfig
}

// Load loads configuration from environment variables and addresses.json.
func Load() (*Config, error) {
	env := getEnv("APP_ENV", "dev")
	httpPort := getEnv("HTTP_PORT", "8080")
	chainEnv := getEnv("CHAIN_ENV", "bscTestnet") // bscTestnet | bscMainnet

	addresses, err := loadAddresses("go_back/addresses.json")
	if err != nil {
		return nil, err
	}

	var chainCfg ChainConfig
	switch chainEnv {
	case "bscTestnet":
		chainCfg = addresses.BSCTestnet
	case "bscMainnet":
		chainCfg = addresses.BSCMainnet
	default:
		return nil, fmt.Errorf("unsupported CHAIN_ENV: %s", chainEnv)
	}

	rpcURL := os.Getenv(chainCfg.RPCUrlEnv)
	if rpcURL == "" {
		return nil, fmt.Errorf("missing RPC url env %s", chainCfg.RPCUrlEnv)
	}

	return &Config{
		Env:         env,
		HTTPPort:    httpPort,
		ChainEnv:    chainEnv,
		RPCURL:      rpcURL,
		ChainConfig: chainCfg,
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func loadAddresses(path string) (*Addresses, error) {
	bz, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read addresses file: %w", err)
	}

	var addr Addresses
	if err := json.Unmarshal(bz, &addr); err != nil {
		return nil, fmt.Errorf("unmarshal addresses: %w", err)
	}
	return &addr, nil
}
