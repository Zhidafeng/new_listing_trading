# API使用说明

## 启动服务

```bash
# 使用默认配置和端口
go run cmd/server/main.go

# 指定配置文件和端口
go run cmd/server/main.go -config config.yaml -port 8080
```

## API接口

### 1. 模拟新币上线并触发交易

**接口**: `POST /api/simulate/new-listing`

**请求体**:
```json
{
  "symbol": "BTCUSDT",
  "notional_usdt": "100"
}
```

**参数说明**:
- `symbol` (必需): 币对名称，例如 "BTCUSDT"
- `notional_usdt` (可选): USDT金额，留空使用配置文件中的默认值

**响应示例**:
```json
{
  "success": true,
  "message": "新币上线模拟成功，已创建订单",
  "symbol": "BTCUSDT",
  "order_set": {
    "symbol": "BTCUSDT",
    "sell_order": {
      "orderId": 123456,
      "symbol": "BTCUSDT",
      "status": "FILLED",
      "avgPrice": "50000.00",
      "executedQty": "0.002"
    },
    "stop_loss_order": {
      "orderId": 123457,
      "symbol": "BTCUSDT",
      "type": "STOP_MARKET",
      "stopPrice": "51000.00"
    },
    "take_profit_order": {
      "orderId": 123458,
      "symbol": "BTCUSDT",
      "type": "TAKE_PROFIT_MARKET",
      "stopPrice": "47500.00"
    }
  }
}
```

**使用示例**:
```bash
curl -X POST http://localhost:8080/api/simulate/new-listing \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "notional_usdt": "100"
  }'
```

### 2. 获取服务状态

**接口**: `GET /api/status`

**响应示例**:
```json
{
  "symbol_count": 573,
  "new_listing_count": 5,
  "last_update_time": "2025-11-04 16:31:56",
  "trading_enabled": true
}
```

**使用示例**:
```bash
curl http://localhost:8080/api/status
```

### 3. 获取新币对列表

**接口**: `GET /api/new-listings`

**响应示例**:
```json
{
  "BTCUSDT": {
    "symbol": "BTCUSDT",
    "onboard_date": 1569398400000,
    "status": "TRADING",
    "found_time": "2025-11-04T16:31:56Z",
    "is_ordered": true,
    "order_time": "2025-11-04T16:32:00Z"
  }
}
```

**使用示例**:
```bash
curl http://localhost:8080/api/new-listings
```

### 4. 获取所有币对

**接口**: `GET /api/symbols`

**响应示例**:
```json
{
  "BTCUSDT": {
    "symbol": "BTCUSDT",
    "onboard_date": 1569398400000,
    "status": "TRADING"
  }
}
```

**使用示例**:
```bash
curl http://localhost:8080/api/symbols
```

## 完整交易流程

当调用模拟新币上线接口时，系统会：

1. **添加新币对**到监控服务的 `newListings` 列表
2. **检查是否已下单**，如果已下单则返回错误
3. **创建市价卖单（做空）**（按USDT金额）
4. **获取成交价格**作为开仓价格
5. **创建止损订单**（基于配置的止损百分比，价格上涨触发）
6. **创建止盈订单**（基于配置的止盈百分比，价格下跌触发）
7. **标记币对为已下单**
8. **返回订单信息**

## 注意事项

1. 如果未配置API密钥，交易功能不可用，但监控功能仍可正常使用
2. 如果币对已经下单过，再次调用会返回错误
3. 止损和止盈订单使用 `closePosition=true`，会自动平掉整个持仓
4. 所有订单金额单位为USDT

