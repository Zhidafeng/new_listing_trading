# 止盈止损功能说明

## 功能概述

本系统实现了币安合约交易的止盈止损功能，支持通过配置文件灵活设置止盈止损阈值。

## 配置文件

配置文件路径：`config.yaml`

### 配置示例

```yaml
# 币安API配置
binance:
  api_key: "your_api_key"
  secret_key: "your_secret_key"
  base_url: ""  # 可选，留空使用生产环境

# 交易配置
trading:
  # 默认下单USDT金额（例如："10"表示10 USDT）
  default_notional: "10"
  # 持仓方向：BOTH/LONG/SHORT
  position_side: "BOTH"
  
  # 止损配置
  stop_loss:
    enabled: true      # 是否启用止损
    percent: 2.0       # 止损百分比（例如：2.0 表示2%，做空时价格上涨2%触发）
    working_type: "MARK_PRICE"  # 触发类型：MARK_PRICE/CONTRACT_PRICE
  
  # 止盈配置
  take_profit:
    enabled: true      # 是否启用止盈
    percent: 5.0       # 止盈百分比（例如：5.0 表示5%，做空时价格下跌5%触发）
    working_type: "MARK_PRICE"  # 触发类型：MARK_PRICE/CONTRACT_PRICE
```

## 使用方法

### 1. 加载配置

```go
import "new_listing_trade/internal/config"

cfg, err := config.LoadConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}
```

### 2. 创建交易服务

```go
import "new_listing_trade/internal/service"

tradingService, err := service.NewTradingService(cfg)
if err != nil {
    log.Fatal(err)
}
```

### 3. 创建带止盈止损的订单

```go
// 创建卖单（做空）并同时设置止损和止盈
symbol := "BTCUSDT"
notionalUSDT := "10"  // 可选，留空使用配置文件中的默认值

orderSet, err := tradingService.CreateOrdersWithStopLossAndTakeProfit(
    symbol, 
    notionalUSDT,
)
if err != nil {
    log.Fatal(err)
}

log.Printf("卖单ID: %d", orderSet.SellOrder.OrderID)
if orderSet.StopLossOrder != nil {
    log.Printf("止损订单ID: %d", orderSet.StopLossOrder.OrderID)
}
if orderSet.TakeProfitOrder != nil {
    log.Printf("止盈订单ID: %d", orderSet.TakeProfitOrder.OrderID)
}
```

### 4. 单独创建止损或止盈订单

```go
// 创建止损订单（需要先有开仓价格）
entryPrice := 50000.0  // 开仓价格
stopLossOrder, err := tradingService.CreateStopLossOrder("BTCUSDT", entryPrice)
if err != nil {
    log.Fatal(err)
}

// 创建止盈订单（需要先有开仓价格）
takeProfitOrder, err := tradingService.CreateTakeProfitOrder("BTCUSDT", entryPrice)
if err != nil {
    log.Fatal(err)
}
```

## 止盈止损计算逻辑（做空）

### 止损计算（做空）
- 止损价格 = 开仓价格 × (1 + 止损百分比/100)
- 例如：开仓价格50000，止损2%，止损价格 = 50000 × (1 + 0.02) = 51000
- 做空时价格上涨触发止损

### 止盈计算（做空）
- 止盈价格 = 开仓价格 × (1 - 止盈百分比/100)
- 例如：开仓价格50000，止盈5%，止盈价格 = 50000 × (1 - 0.05) = 47500
- 做空时价格下跌触发止盈

## 订单类型说明

- **MARKET（SELL）**: 市价卖单，用于做空开仓
- **STOP_MARKET**: 止损市价单，做空时当价格上涨到止损价格时，以市价买入平仓
- **TAKE_PROFIT_MARKET**: 止盈市价单，做空时当价格下跌到止盈价格时，以市价买入平仓

## 注意事项

1. **所有订单都是做空（SELL）方向**
2. 止盈止损订单需要在持仓建立后创建
3. 订单类型使用 `MARK_PRICE` 或 `CONTRACT_PRICE` 作为触发价格类型
4. `ClosePosition: "true"` 表示平仓订单，会自动平掉整个持仓（不需要指定数量）
5. `PriceProtect: "true"` 启用价格保护，避免滑点过大
6. 做空时：
   - 止损：价格上涨到止损价格触发（买入平仓）
   - 止盈：价格下跌到止盈价格触发（买入平仓）

