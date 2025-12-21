启动命令

```bash
set -a
source .env
set +a
env | grep BSC_TESTNET_RPC
go run ./cmd/api
```
# CINA Dex On‑Chain API (Go 后端调用说明)

## 0. 这个项目在做什么？

这是一个**简单版的 DeFi 借贷池**，目前部署在 BSC 上：

- 抵押物：当前链的**原生币**（在 BSC 上是 **BNB**）  
- 借款资产：USDT（BSC 主网用真实 USDT，测试网用 MockUSDT）
- 用户角色：
  - **存款人（LP）**：存入 USDT，获得 LP 份额 `FToken(cUSDT)`，赚利息；
  - **借款人**：抵押 BNB，按固定利率借出 USDT；
  - **清算人**：当借款人风险过高时，清算人代还借款，获得一部分抵押品奖励。

协议的核心逻辑：

1. **存款 / 提取**  
   - 用户调用 `LendingPool.deposit(amount)` 存入 USDT。  
   - 池子按当前汇率铸造 FToken 给用户，记录他在池中的份额。  
   - 汇率通过 `getExchangeRate()` 动态变化，利息收入体现在 FToken 对 USDT 的兑换比例上涨。

2. **抵押借款**  
   - 借款人发送 BNB（`msg.value`），并调用 `borrow(amount, duration)`。  
   - 合约通过链上预言机（`ChainlinkOracle` + BNB/USD 价格）计算抵押物的美元价值，按最大 LTV（75%）判断能借多少 USDT。  
   - 借款成功后，合约把 USDT 转给借款人，并在内部记录一条 `Loan`（包含本金、利息、抵押量、期限等）。

3. **还款**  
   - 借款人在到期前后任何时候可以 `repay(loanId)`。  
   - 需要先对 USDT `approve` 足够额度。  
   - 合约收到应还总额（本金 + 利息）后，将抵押的 BNB 全部返还给借款人，并把贷款标记为已结清。

4. **清算**  
   - 如果价格下跌导致贷款的 **债务/抵押价值** 超过清算阈值（80%），任何人都可以调用 `liquidate(loanId)`：  
     - 清算人代借款人一次性还清全部 USDT；  
     - 清算人按“债务价值 × 104%”的规模获得一部分抵押 BNB；  
     - 剩余抵押，如果还有，退给借款人。  
   - 这样 LP 的资金安全有保障，清算人也有激励。

5. **后端的主要职责**

Go 后端在这个系统里主要做两件事：

- **读数据**：  
  - 获取池子整体状态（总资产、总借出、本地汇率等）给前端展示；  
  - 获取某个用户的贷款列表和整体风险状况。

- **辅助写操作**：  
  - 构造、签名并发送交易（如果你托管私钥）；  
  - 或者生成给前端用的钱包签名参数（如果你让前端自己发 tx）。

后面章节是给 Go 后端看的「接口说明」：包括链上地址、ABI 所对应的函数、以及在 Go 中如何调用的建议。

> 说明：抵押物为当前链的原生币（在 BSC 上为 BNB），借款资产为 USDT（测试网为 MockUSDT）。

---

## 1. 网络 & 地址

集中配置在 `go_back/addresses.json`：

- BSC Testnet（chainId 97）
  - `rpcUrlEnv`: `BSC_TESTNET_RPC`
  - `mockUsdt`: `0xBd8627a3b43d45488e6f15c92Ec3A8A277B1f79d`
  - `priceFeed` (BNB/USD, Binance Oracle): `0x1A26d803C2e796601794f8C5609549643832702C`
  - `chainlinkOracle`: `0x91D2f77c0Cf3D2A2b59F4D6B09314453Bfa63357`
  - `fToken` (cUSDT): `0x35724F5AD969153846189B19bd4A76309EFCE768`
  - `lendingPool`: `0x423B4DA844Fd57A2D93CD25E927eE43A9A9A4a4a`

- BSC Mainnet（chainId 56）
  - `rpcUrlEnv`: `BSC_MAINNET_RPC`
  - `usdt`: `0x55d398326f99059fF775485246999027B3197955`
  - `priceFeed` (BNB/USD, Chainlink): `0x0567F2323251f0Aab15c8dFb1967E4e8A7D42aeE`
  - `chainlinkOracle`: `TBD`
  - `fToken` (cUSDT): `TBD`
  - `lendingPool`: `TBD`

Go 后端建议：

- 从环境变量读取 RPC：
  - `BSC_TESTNET_RPC`
  - `BSC_MAINNET_RPC`
- 解析 `addresses.json`，按环境选择 testnet/mainnet 的地址。

---

## 2. LendingPool 合约接口

ABI 文件：`go_back/abi/LendingPool.json`  
合约地址：

- Testnet: `0x423B4DA844Fd57A2D93CD25E927eE43A9A9A4a4a`
- Mainnet: (部署后填入)

### 2.1 只读方法（view）

- `function usdt() external view returns (address)`
- `function fToken() external view returns (address)`
- `function oracle() external view returns (address)`
- `function nextLoanId() external view returns (uint256)`
- `function totalBorrowed() external view returns (uint256)`
- `function loans(uint256 loanId) external view returns (address borrower, uint256 collateralAmount, uint256 principal, uint256 repaymentAmount, uint256 startTime, uint256 duration, bool isActive)`
- `function getExchangeRate() external view returns (uint256)`  
  - 18 位精度，`1e18` 表示 1:1。
- `function getPoolState() external view returns (uint256 totalAssets, uint256 totalBorrowed_, uint256 availableLiquidity, uint256 exchangeRate, uint256 totalFTokenSupply)`
- `function getUserLoans(address user) external view returns (uint256[] loanIds)`
- `function getUserPosition(address user) external view returns (uint256[] loanIds, uint256 totalPrincipal, uint256 totalRepayment, uint256 totalCollateral)`
- `function getLoanHealth(uint256 loanId) external view returns (uint256 ltv, bool isLiquidatable)`
- `function getLenderPosition(address user) external view returns (uint256 fTokenBalance, uint256 exchangeRate, uint256 underlyingBalance)`

### 2.2 状态改变方法

#### deposit

```solidity
function deposit(uint256 amount) external;
```

- 参数：
  - `amount`：USDT 数量，6 位小数（例如 1000 USDT = `1000 * 1e6`）。
- 需要用户先对 `usdt` 调用 `approve(pool, amount)`。
- Go 里需要用调用用户私钥发起交易。

#### borrow

```solidity
function borrow(uint256 amount, uint256 duration) external payable;
```

- 参数：
  - `amount`：借款 USDT 数量（6 decimals）
  - `duration`：借款期限（秒）
- `msg.value`：原生币抵押数量（比如 1 BNB = `1e18`）。
- 约束（会 revert）：
  - `amount > 0`
  - `msg.value > 0`
  - `duration > 0`
  - 抵押物价值 × LTV >= 借款额（以 18 decimals 比较）

Go 里发交易时要设置：

- `Value: collateralAmount`（以 wei 表示 BNB）

#### repay

```solidity
function repay(uint256 loanId) external;
```

- 参数：
  - `loanId`：贷款 ID（从 0 开始）
- 行为：
  - 从调用者转入 `repaymentAmount` USDT
  - 归还全部抵押原生币给借款人
  - 将 loan 标记为 inactive

需要先对 USDT 调用 `approve(pool, repaymentAmount)`。

#### liquidate

```solidity
function liquidate(uint256 loanId) external;
```

- 调用者 = 清算人
- 行为：
  - 检查 LTV 是否超过 `LIQUIDATION_THRESHOLD`，否则 revert `"Health factor ok"`
  - 清算人代还全部 `repaymentAmount` USDT（需要先 approve）
  - 按 104% 债务价值计算清算可得原生币，给清算人
  - 剩余抵押返还给借款人

Go 里：清算机器人可以周期性扫描：

- `getUserLoans` + `getLoanHealth`  
对于 `isLiquidatable == true` 的 loan 调用 `liquidate`.

---

## 3. FToken (LP Token) 合约接口

ABI 文件：`go_back/abi/FToken.json`  
地址：

- Testnet: `0x35724F5AD969153846189B19bd4A76309EFCE768`
- Mainnet: (部署后填入)

继承自 OpenZeppelin `ERC20` + `Ownable`：

- 常用只读方法：
  - `function name() external view returns (string)`
  - `function symbol() external view returns (string)`
  - `function decimals() external view returns (uint8)` (默认 18)
  - `function totalSupply() external view returns (uint256)`
  - `function balanceOf(address) external view returns (uint256)`
- 只读 LP 汇总数据推荐直接从 `LendingPool.getPoolState()` 获取。

mint/burn 只允许 `LendingPool` 调用，后端不会直接操作。

---

## 4. ChainlinkOracle 合约接口

ABI 文件：`go_back/abi/ChainlinkOracle.json`  
地址：

- Testnet: `0x91D2f77c0Cf3D2A2b59F4D6B09314453Bfa63357`
- Mainnet: (部署后填入)

主要函数：

```solidity
function priceFeeds(address asset) external view returns (address);
function setPriceFeed(address asset, address feed) external onlyOwner;
function getPrice(address asset) external view returns (uint256);
```

约定：

- `asset = address(0)` 表示原生币（BNB）
- 返回值 `getPrice(address(0))` 为 18 位精度的 BNB/USD 价格（例如 2000 美元 ≈ `2000e18`）。

Go 端一般只需要读取价格：

- 使用 `Call` 调用 `getPrice(address(0))`。

`setPriceFeed` 只在部署/运维阶段通过脚本或运维账号调用。

---

## 5. MockUSDT / 主网 USDT 合约接口

ABI 文件（测试网 MockUSDT）：`go_back/abi/MockUSDT.json`  
测试网地址：`0xBd8627a3b43d45488e6f15c92Ec3A8A277B1f79d`  
主网地址（真实 USDT）：`0x55d398326f99059fF775485246999027B3197955`

MockUSDT 扩展了一个 `mint`，但生产环境下你只会用标准 ERC20 接口：

- `function name() external view returns (string)`
- `function symbol() external view returns (string)`
- `function decimals() external view returns (uint8)`（6）
- `function totalSupply() external view returns (uint256)`
- `function balanceOf(address) external view returns (uint256)`
- `function allowance(address owner, address spender) external view returns (uint256)`
- `function approve(address spender, uint256 amount) external returns (bool)`
- `function transfer(address to, uint256 amount) external returns (bool)`
- `function transferFrom(address from, address to, uint256 amount) external returns (bool)`

在 Go 中，所有对 `LendingPool.deposit/repay` 相关的 USDT 流转都需要先做 `approve`。

---

## 6. Go 后端建议的封装接口（示例）

建议在 Go 项目中抽象出服务层，例如：

- `PoolService`：
  - `GetPoolState(ctx) (*PoolState, error)` → 调用 `getPoolState`
  - `GetUserPosition(ctx, userAddr) (*UserPosition, error)` → 调用 `getUserPosition`
  - `Deposit(ctx, userKey, amount)` → 调用 `usdt.approve` + `pool.deposit`
  - `Borrow(ctx, userKey, amount, duration, collateralWei)` → 调用 `pool.borrow`（带 value）
  - `Repay(ctx, userKey, loanId)` → 先读 loan、然后 approve，最后调用 `repay`
  - `Liquidate(ctx, botKey, loanId)` → `usdt.approve` + `liquidate`

- `OracleService`：
  - `GetNativePrice(ctx) (big.Int, error)` → `getPrice(address(0))`

常用流程：

1. 前端请求 → Go 后端 → 查询链上数据（`Call`）返回只读信息。
2. 需要用户签名的操作（存款 / 借款 / 还款 / 清算），建议：
   - 要么由前端直接签名发起交易；
   - 要么由后端托管私钥（风险更大），通过 Go 的 `bind.NewKeyedTransactorWithChainID` 构造交易。

ABI JSON 已生成在：

- `go_back/abi/LendingPool.json`
- `go_back/abi/ChainlinkOracle.json`
- `go_back/abi/FToken.json`
- `go_back/abi/MockUSDT.json`

可以直接用 `abigen` 或 `bind.NewBoundContract` 生成/使用 Go 绑定。
