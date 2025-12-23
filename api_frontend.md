# CINA Dex On-Chain API（前端调用说明 / Frontend API Doc）

> 说明：本文档面向前端开发，列出当前可用的 HTTP 接口、请求示例和返回结构。  
> 后端基于 Gin 实现，统一前缀为 `/api/v1`，所有接口返回统一的 JSON 包装格式。

## 0. 通用约定

- Base URL：`http://{host}:{HTTP_PORT}/api/v1`
- 统一响应结构（`pkg/response/response.go`）：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

- 约定：
  - `code = 0` 表示成功；
  - 常见错误码：
    - `4001`：参数校验失败（JSON 绑定错误 / 必填字段缺失等）；
    - `4002`：路径参数格式错误；
    - `1001`：后端内部错误或链上调用失败。
  - 所有数值型的链上金额/价格都用字符串返回，前端自行做精度处理。

---

## 1. 健康检查

### GET `/health`

- 功能：存活探针，用于前端/监控检查服务是否正常。
- 请求参数：无
- 响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok"
  }
}
```

---

## 2. 借贷池（Pool）相关

### GET `/pool/state`

- 功能：查询借贷池整体状态，用于首页大盘展示。
- 后端实现细节：
  - 有一个后台任务每 **3 分钟** 从链上拉取一次池子状态并缓存；
  - 接口优先返回缓存，缓存不存在时会实时读链。
- 请求参数：无
- 响应 `data` 结构（`model.PoolState`）：

```json
{
  "totalAssets": "string",         // 池子总资产（USDT 最小单位，整数字符串）
  "totalBorrowed": "string",       // 总借出
  "availableLiquidity": "string",  // 当前可用流动性
  "exchangeRate": "string",        // FToken->USDT 汇率，18 位精度，1e18 = 1:1
  "totalFTokenSupply": "string"    // FToken 总供应量
}
```

---

## 3. 用户维度（User）接口

路径中的 `address` 都是用户链上地址（0x 开头的 42 字符串）。

### 3.1 GET `/users/:address/position`

- 功能：查询用户整体借款头寸（总本金/总应还/总抵押）。
- 请求参数：
  - Path：`address`（用户地址）
- 响应 `data` 结构（`model.UserPosition`）：

```json
{
  "address": "0x...",
  "loanIds": [1, 2, 3],
  "totalPrincipal": "string",   // 总本金
  "totalRepayment": "string",   // 总应还（本+息）
  "totalCollateral": "string"   // 总抵押 BNB，单位 wei
}
```

---

### 3.2 GET `/users/:address/lender-position`

- 功能：查询用户作为 LP（存款人）的头寸与收益信息。
- 请求参数：
  - Path：`address`
- 响应 `data` 结构（`model.LenderPosition`）：

```json
{
  "address": "0x...",
  "fTokenBalance": "string",      // FToken 余额（18 位）
  "exchangeRate": "string",       // 当前汇率（18 位）
  "underlyingBalance": "string",  // 按汇率换算后的 USDT 数量
  "netDeposited": "string",       // 当前暂时固定为 "0"（后续会接真实数据）
  "interest": "string"            // 当前暂时等于 underlyingBalance
}
```

> 当前版本中，`netDeposited` 和 `interest` 只是占位逻辑，前端展示时可以标注为“估算”。

---

### 3.3 GET `/users/:address/loans`

- 功能：查询用户所有贷款列表。
- 请求参数：
  - Path：`address`
- 响应 `data` 结构：`Loan[]` 数组（`model.Loan`）：

```json
[
  {
    "id": 1,
    "borrower": "0x...",
    "collateralAmount": "string",  // 抵押 BNB，wei
    "principal": "string",         // 借款本金（USDT 最小单位）
    "repaymentAmount": "string",   // 应还总额（本+息）
    "startTime": 1234567890,       // 借款开始时间，秒级时间戳
    "duration": 3600,              // 借款时长（秒）
    "isActive": true
  }
]
```

---

## 4. 单笔贷款（Loan）接口

### 4.1 GET `/loans/:loanId`

- 功能：查询指定 `loanId` 的详细信息。
- 请求参数：
  - Path：`loanId`（字符串形式的整数，例如 `"1"`）
- 响应 `data`：同上 `Loan` 结构。

---

### 4.2 GET `/loans/:loanId/health`

- 功能：查询指定贷款的健康度（LTV、是否可清算）。
- 请求参数：
  - Path：`loanId`
- 响应 `data` 结构（`model.LoanHealth`）：

```json
{
  "ltv": "string",          // LTV，18 位精度
  "isLiquidatable": true    // 是否达到清算条件
}
```

---

## 5. 报价 / 风险接口

### POST `/borrow/quote`

- 功能：根据指定的 USDT 借款金额，计算需要抵押的 BNB 数量。
- 数据来源：
  - 后台任务每 **3 分钟** 从链上读取一次 BNB/USD 价格并缓存；
  - 接口优先使用缓存价格，没有缓存时实时读链。
- 请求 Body：

```json
{
  "amount": "100000000"
}
```

- 字段说明：
  - `amount`（必填）：想要借的 USDT 数量，最小单位（6 位小数），例如 100 USDT = `"100000000"`.

- 响应 `data` 结构（`model.BorrowQuote`）：

```json
{
  "borrowAmount": "100000000",    // 请求的借款金额（原样返回）
  "collateralWei": "1234567890",  // 所需抵押的 BNB 数量，wei
  "bnbUsdPrice": "2000000000000000000000", // 使用的 BNB/USD 价格，18 位
  "maxLtvPercent": "75"           // 使用的最大 LTV（百分比）
}
```

前端可直接用：

```text
BNB 数量 = collateralWei / 1e18
```

---

## 6. 交易构建（Tx Builder）接口

这些接口都只返回交易调用参数（`to/data/value`），**不会直接发送链上交易**。  
前端需要使用钱包（例如 MetaMask）或者自有 signer 去签名并发送。

通用结构 `TxCall`：

```json
{
  "to": "0x...",     // 合约地址
  "data": "0x...",   // ABI 编码后的 calldata
  "value": "0"       // 发送的原生币数量（wei），字符串
}
```

---

### 6.1 POST `/tx/deposit`

- 功能：构建“存款 USDT”所需的两笔交易：`approve` + `deposit`。
- 请求 Body：

```json
{
  "userAddress": "0x...",      // 当前未在后端使用，仅占位
  "amount": "100000000"        // 必填，USDT 数量（最小单位）
}
```

- 响应 `data` 结构（`model.DepositTx`）：

```json
{
  "approve": {
    "to": "0xToken",     // USDT 合约地址
    "data": "0x...",     // approve(spender, amount)
    "value": "0"
  },
  "deposit": {
    "to": "0xPool",      // LendingPool 合约地址
    "data": "0x...",     // deposit(amount)
    "value": "0"
  }
}
```

前端调用顺序：
1. 先用钱包调用 `approve`；
2. approve 成功后再调用 `deposit`。

---

### 6.2 POST `/tx/borrow`

- 功能：构建“借款 USDT”所需的交易（BNB 作为抵押，通过 `msg.value` 发送）。
- 请求 Body：

```json
{
  "userAddress": "0x...",
  "amount": "100000000",     // 必填，USDT 借款金额，最小单位
  "duration": 3600,          // 必填，借款时长（秒）
  "collateralWei": "123456"  // 必填，抵押 BNB 数量，wei
}
```

- 响应 `data` 结构（`model.BorrowTx`）：

```json
{
  "borrow": {
    "to": "0xPool",
    "data": "0x...",      // borrow(amount, duration)
    "value": "123456"     // 作为 msg.value 发送的 BNB 数量（wei）
  }
}
```

前端发送时需要注意：钱包发送交易时的 `value` 字段要用返回的 `value`。

---

### 6.3 POST `/tx/repay`

- 功能：构建“还款”需要的两笔交易：`approve` + `repay`。
- 请求 Body：

```json
{
  "userAddress": "0x...",
  "loanId": 1
}
```

- 后端逻辑：
  - 从链上读取 `loans(loanId).repaymentAmount`；
  - 用该金额构建：
    - ERC20 `approve(pool, repaymentAmount)`；
    - LendingPool `repay(loanId)`。

- 响应 `data` 结构（`model.RepayTx`）：

```json
{
  "approve": {
    "to": "0xToken",
    "data": "0x...",
    "value": "0"
  },
  "repay": {
    "to": "0xPool",
    "data": "0x...",
    "value": "0"
  }
}
```

---

### 6.4 POST `/tx/liquidate`

- 功能：构建“清算”所需的两笔交易：`approve` + `liquidate`。
- 请求 Body：

```json
{
  "userAddress": "0x...",
  "loanId": 1
}
```

- 后端逻辑与还款类似：
  - 使用当前 `repaymentAmount` 作为清算人需要支付的 USDT 数量；
  - 构建 ERC20 `approve` 与 `liquidate(loanId)` 调用。

- 响应 `data` 结构（`model.LiquidateTx`）：

```json
{
  "approve": {
    "to": "0xToken",
    "data": "0x...",
    "value": "0"
  },
  "liquidate": {
    "to": "0xPool",
    "data": "0x...",
    "value": "0"
  }
}
```

---

### 6.5 POST `/tx/withdraw`

- 功能：构建 LP 赎回 FToken 的交易。
- 请求 Body：

```json
{
  "userAddress": "0x...",
  "fTokenAmount": "1000000000000000000"  // 必填，FToken 数量，18 位
}
```

- 响应 `data` 结构（`model.WithdrawTx`）：

```json
{
  "withdraw": {
    "to": "0xPool",
    "data": "0x...",   // withdraw(amount)
    "value": "0"
  }
}
```

> 实际收到的 USDT 数量由链上的当前 `exchangeRate` 决定，前端只需要传入想赎回的 FToken 数量。

---

### 6.6 POST `/tx/mock-usdt/mint`

- 功能：在测试网环境构建 MockUSDT `mint` 交易，用于水龙头场景。
- 前提：当前链配置中有 `mockUsdt` 地址（即测试网）。
- 请求 Body：

```json
{
  "to": "0x...",           // 必填，接收 MockUSDT 的地址
  "amount": "100000000"    // 必填，MockUSDT 数量，6 位
}
```

- 响应 `data` 结构（`model.TxCall`）：

```json
{
  "to": "0xMockUSDT",
  "data": "0x...",   // mint(to, amount)
  "value": "0"
}
```

> 注意：这笔交易必须由拥有 `mint` 权限的钱包（例如 owner）签名并发送。

---

## 7. Swagger / OpenAPI

后端同时提供 Swagger 文档接口，前端可以用来调试或导入 Postman：

- `GET /swagger`  
  - 返回 Swagger UI 页面（HTML）。
- `GET /swagger/openapi.json`  
  - 返回 OpenAPI JSON 规范，可导入到 Postman / Hoppscotch 等工具中。

