# CLOB V2 迁移状态

最后更新：2026-05-02

本文档用于记录当前 SDK 对 Polymarket CLOB V2 的兼容状态，并作为后续
V2 适配工作的实现清单。当前文档基于第一阶段交易链路迁移后的代码状态整理。

## 当前状态

当前默认创建订单链路已经切换到 CLOB V2：

- `Client.CreateOrder`
- `Client.CreateOrderFromSignable`
- `Client.PostOrder`
- `Client.PostOrders`
- `clob.SignOrder`
- `clob.NewOrderBuilder(...).BuildSignableWithContext`
- `clob.NewOrderBuilder(...).BuildMarketWithContext`

V1 签名逻辑仍然保留，用于兼容旧链路。需要显式指定才会走 V1：

```go
client = client.WithOrderVersion(1)
```

或者：

```go
order.Version = 1
```

也就是说，当前改造不是删除 V1，而是将默认订单创建链路切到 V2，同时保留
V1 作为兼容选项。

## 已完成

### V2 订单签名

- 已增加 V2 EIP-712 domain version `2`。
- 已增加 V2 订单 typed-data 字段：
  - `timestamp`
  - `metadata`
  - `builder`
- V2 签名消息中不再包含以下 V1 字段：
  - `taker`
  - `expiration`
  - `nonce`
  - `feeRateBps`
- V1 签名仍保留在显式版本选择之后。

### V2 订单提交 Payload

V2 订单提交 payload 当前包含：

- `salt`
- `maker`
- `signer`
- `taker`
- `tokenId`
- `makerAmount`
- `takerAmount`
- `side`
- `signatureType`
- `timestamp`
- `expiration`
- `metadata`
- `builder`
- `signature`

注意：`expiration` 仍然会出现在提交给 API 的 payload 中，但不会进入 V2
EIP-712 签名消息。

### 交易所合约地址

- 已增加 Polygon 和 Amoy 的 V1/V2 普通市场、negative-risk 市场合约地址。
- 已增加合约地址覆盖配置：
  - `clob.Client.WithExchangeAddresses`
  - `polymarket.WithCLOBExchangeAddresses`

### Builder Code

- 已支持在 V2 订单中写入 `builder` bytes32 code：
  - `clob.Client.WithBuilderCode`
  - `polymarket.WithCLOBBuilderCode`
  - `clob.OrderBuilder.BuilderCode`
- 已有的 builder auth headers 仍保留，用于需要鉴权的 builder 相关接口。

### Demo

已新增 [examples/v2_trading](../examples/v2_trading)，覆盖：

- 限价单
- 市价单
- 取消订单

## 待补充工作

### P0：版本发现与版本不匹配重试

状态：未实现

官方 V2 client 会查询 API 当前支持的 order version，并在遇到版本不匹配时
刷新版本后重试。当前 Go SDK 默认走 V2，也允许显式切回 V1，但还没有主动
向 CLOB API 查询当前版本。

需要补充：

- 增加 `GET /version`。
- 在 CLOB client 中缓存返回的版本。
- 增加强制刷新版本的方法。
- 遇到 order-version mismatch 响应时，刷新版本并自动重试一次。

影响：

- 当前 SDK 可以支持现阶段 V2-first 的交易链路。
- 如果后续 Polymarket 调整活跃订单版本，或者 API 返回版本不匹配错误，
  SDK 还不能自动恢复。

### P0：V2 Market Info 缓存

状态：未实现

官方 V2 client 使用市场信息接口统一解析 token 到 condition 的关系，并缓存
fee、tick size、negative-risk 等信息。

需要补充：

- 增加 `GET /clob-markets/{conditionID}`。
- 增加 `GET /markets-by-token/{tokenID}`。
- 增加 compact market details 的响应模型，包括：
  - condition ID
  - token IDs
  - min tick size
  - negative-risk 标记
  - fee details（`fd.r`、`fd.e`）
  - 如果接口仍返回 legacy base fees，则兼容 `mbf`、`tbf`
- 缓存 token 到 condition 的映射。
- 构建 V2 订单时优先使用该缓存。

影响：

- 当前 order builder 仍分别调用 `/tick-size`、`/fee-rate`、`/neg-risk`。
- V2 fee exponent 和 platform fee details 还没有进入市价单金额计算逻辑。

### P0：Builder Fee 支持

状态：未实现

当前 SDK 可以在 V2 订单中写入 `builder` bytes32 code，但还没有查询 builder
fee，也没有在市价买单金额计算中纳入 builder fee。

需要补充：

- 增加 `GET /fees/builder-fees/{builderCode}`。
- 增加 builder fee 响应模型。
- 按 builder code 缓存 fee rate。
- BUY 市价单金额调整时计入 builder taker fee。

影响：

- 当前已经可以创建带 builder 归因的 V2 订单。
- 带 builder code 的 BUY 市价单还没有完整的 fee-aware sizing。

### P1：POLY_1271 签名

状态：只有常量，未实现签名逻辑

当前已经增加 `auth.SignaturePoly1271 = 3`，但 V2 签名链路遇到该类型会返回
明确的 unsupported 错误。

需要补充：

- 实现 EIP-1271 / smart-contract wallet 签名支持。
- 实现官方 client 使用的 V2 `TypedDataSign` wrapper signature 格式。
- 增加基于官方 SDK 输出或已知向量的测试。

影响：

- EOA、Proxy、Gnosis Safe 风格签名仍可用。
- 使用 POLY_1271 的智能合约钱包或 vault 流程暂不可用。

### P1：Builder Trades V2 数据结构

状态：已有接口，但仍偏旧数据结构

当前 SDK 已有 `BuilderTrades`，但请求和响应类型仍更接近旧版 trade query
shape。

需要补充：

- builder trade 查询参数增加 `builder_code`。
- 增加 V2 builder trade 响应模型，字段可能包括：
  - builder
  - size USDC
  - fee / fee USDC
  - maker
  - trade type
  - created/updated timestamps
- 决定是在现有 `BuilderTrades` 上演进，还是新增 `BuilderTradesV2`。

影响：

- 如果 API 仍兼容旧格式，当前 builder trade 查询可能还能工作。
- 但 V2 builder-specific analytics 还没有完整表达。

### P1：Pre-Migration Orders

状态：未实现

官方 V2 client 暴露了 pre-migration order 查询能力，用于迁移阶段或历史兼容
场景。

需要补充：

- 增加 `GET /data/pre-migration-orders`。
- 增加请求和响应类型。
- 如接口分页，增加分页 helper。

影响：

- 常规 open order 查询已经可用。
- 迁移期历史订单支持还不完整。

### P1：Market Live Activity Endpoint

状态：仍使用旧接口

当前 SDK 方法：

- `GET /v1/market-trades-events/{id}`

官方 V2 client 路径：

- `GET /markets/live-activity/{conditionID}`

需要补充：

- 增加 V2 路径。
- 确认响应 schema 差异。
- 决定更新 `MarketTradesEvents`，还是新增一个命名更清晰的 V2 方法。

影响：

- 旧市场成交事件接口可能仍可用。
- V2 live activity endpoint 还没有暴露。

### P1：Rewards User Markets Endpoint

状态：仍使用旧接口

当前 SDK 方法：

- `GET /rewards/user/by-market`

官方 V2 client 路径：

- `GET /rewards/user/markets`

需要补充：

- 确认 `/rewards/user/by-market` 是否仍被线上 API 支持。
- 增加或迁移到 `/rewards/user/markets`。
- 确认响应 schema 是否兼容。

影响：

- 如果旧接口仍兼容，当前 rewards by market 方法可能继续可用。
- 但 V2 路径还没有完整对齐。

### P2：RFQ V2 订单 Payload Helper

状态：helper 仍偏 V1 风格

当前 RFQ accept / approve helper 构建的 payload 仍包含 V1 风格字段：

- `nonce`
- `feeRateBps`
- `taker`

需要补充：

- 确认 RFQ V2 accept / approve 的 payload schema。
- 如果 RFQ 期望 V2 order，补充 `timestamp`、`metadata`、`builder`。
- 保留 V1 helper 行为，并通过版本选择区分。

影响：

- 标准 CLOB V2 订单已经支持。
- RFQ 流程在生产使用前仍可能需要 V2 专项适配。

### P2：Fee-Aware Market Order Sizing

状态：基础市价单可用，但未完整处理费用

当前 market order builder 可以创建并提交 V2 市价单，但还没有完整复刻官方
SDK 对 BUY 市价单的 fee-aware amount adjustment。

需要补充：

- 在 builder 中接受可选用户 USDC balance，或新增独立 helper。
- 使用 V2 market info 中的 platform fee rate 和 exponent。
- 如果存在 builder code，计入 builder taker fee。
- 调整 BUY 市价单金额，避免费用计入后出现余额不足。

影响：

- 使用显式金额的市价单当前可以工作。
- 在该逻辑完成前，用户需要自行预留费用空间。

## 建议实现顺序

1. 增加 `/version` 和版本不匹配自动重试。
2. 增加 `/markets-by-token`、`/clob-markets` 的模型和缓存。
3. 增加 builder fee 查询，并完善 BUY 市价单 fee-aware sizing。
4. 补齐 builder trades、live activity、rewards、pre-migration orders 等 V2
   endpoint parity。
5. 实现 POLY_1271 签名支持。
6. 确认 RFQ V2 payload schema 后更新 RFQ helpers。

## 当前验证

当前 V2 交易链路迁移已覆盖以下测试：

- V2 typed-data shape
- V1 typed-data 兼容性
- V2 order payload shape
- client / order builder 的 V2 builder 默认值
- signature type enum values

完整测试命令：

```bash
GOTOOLCHAIN=auto go test ./...
```
