# V2 Trading Demo

This example uses the current SDK CLOB V2 order flow to place a limit order,
place a market order, and cancel an order.

Set authentication first:

```bash
export POLYMARKET_PK="0x..."
export POLYMARKET_API_KEY="..."
export POLYMARKET_API_SECRET="..."
export POLYMARKET_API_PASSPHRASE="..."
export POLYMARKET_TOKEN_ID="..."
```

Optional wallet settings:

```bash
export POLYMARKET_CHAIN_ID="137"          # 137 Polygon, 80002 Amoy
export POLYMARKET_SIGNATURE_TYPE="0"      # 0 EOA, 1 Proxy, 2 Safe
export POLYMARKET_FUNDER="0x..."          # required for proxy/safe funder override
export POLYMARKET_BUILDER_CODE="0x..."    # optional V2 builder bytes32
```

Place a GTC limit order:

```bash
export POLYMARKET_DEMO_MODE="limit"
export POLYMARKET_LIMIT_SIDE="BUY"
export POLYMARKET_LIMIT_PRICE="0.50"
export POLYMARKET_LIMIT_SIZE="5"
go run ./examples/v2_trading
```

Place a FAK market order:

```bash
export POLYMARKET_DEMO_MODE="market"
export POLYMARKET_MARKET_SIDE="BUY"
export POLYMARKET_MARKET_AMOUNT_USDC="5"
go run ./examples/v2_trading
```

For a SELL market order, use shares:

```bash
export POLYMARKET_MARKET_SIDE="SELL"
export POLYMARKET_MARKET_AMOUNT_SHARES="5"
go run ./examples/v2_trading
```

Cancel an order:

```bash
export POLYMARKET_DEMO_MODE="cancel"
export POLYMARKET_CANCEL_ORDER_ID="..."
go run ./examples/v2_trading
```

`POLYMARKET_DEMO_MODE=all` runs limit, market, then cancel. If
`POLYMARKET_CANCEL_ORDER_ID` is not set, it cancels the limit order it just
created.
