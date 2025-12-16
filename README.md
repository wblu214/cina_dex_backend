启动命令

```bash
set -a
source .env
set +a
env | grep BSC_TESTNET_RPC
go run ./cmd/api
```
