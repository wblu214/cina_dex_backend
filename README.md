启动命令
solidity
set -a
source .env
set +a
env | grep BSC_TESTNET_RPC
go run ./cmd/api
、、、