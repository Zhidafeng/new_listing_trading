# 币安统一账户接口说明

## 概述

币安统一账户（Portfolio Margin）是一个统一的账户系统，允许在一个账户下管理多种交易产品：
- U本位合约（USD-M Futures）
- 币本位合约（COIN-M Futures）
- 现货
- 期权

## 接口差异对比

### 当前实现（U本位合约 fapi）

| 项目 | U本位合约 (fapi) | 统一账户 (papi) |
|------|------------------|----------------|
| BaseURL | `https://fapi.binance.com` | `https://papi.binance.com` |
| 交易所信息 | `/fapi/v1/exchangeInfo` | `/papi/v1/um/exchangeInfo` |
| 下单接口 | `/fapi/v1/order` | `/papi/v1/um/order` |
| 查询订单 | `/fapi/v1/order` | `/papi/v1/um/order` |
| 价格查询 | `/fapi/v1/ticker/price` | `/papi/v1/um/ticker/price` |

### 统一账户接口特点

1. **统一账户管理**：所有资产在一个账户下，共享保证金
2. **接口路径**：统一账户U本位合约接口路径为 `/papi/v1/um/...`（um表示U本位）
3. **签名方式**：支持HMAC和RSA两种签名方式
4. **错误处理**：对HTTP 503错误有更详细的分类和处理要求

## 关键差异说明

### 1. 接口路径差异

统一账户接口需要在原路径前加上 `/papi/v1/`，并根据产品类型选择：
- U本位合约：`/papi/v1/um/...`
- 币本位合约：`/papi/v1/cm/...`
- 现货：`/papi/v1/sapi/...`

### 2. HTTP 503 错误处理

统一账户对503错误有特殊处理要求：

#### A. "Unknown error"（执行状态未知）
- **语义**：请求已接收但状态未知，可能已成功
- **处理**：不要直接重试，先查询订单状态确认是否已执行
- **是否计入限速**：可能计入

#### B. "Service Unavailable"（失败）
- **语义**：服务暂不可用，100%失败
- **处理**：退避重试（200ms → 400ms → 800ms，最多3-5次）
- **是否计入限速**：不计入

#### C. "Request throttled"（-1008，失败）
- **语义**：系统过载限流
- **处理**：退避重试并降低并发
- **特殊说明**：平仓订单（`closePosition=true`）不受限流影响

### 3. 签名方式

统一账户支持两种签名方式：

#### HMAC签名（当前使用）
```go
// 使用HMAC SHA256签名
mac := hmac.New(sha256.New, []byte(secretKey))
mac.Write([]byte(queryString))
signature := hex.EncodeToString(mac.Sum(nil))
```

#### RSA签名（可选）
```bash
# 使用RSA私钥签名
echo -n 'queryString' | openssl dgst -keyform PEM -sha256 -sign private_key.pem | openssl enc -base64
```

### 4. 访问限制

- **IP限制**：基于IP的访问频率限制
- **权重系统**：不同接口有不同的权重值
- **频率限制**：违反限制会收到HTTP 429，继续违反会被封IP（HTTP 418）

## 迁移到统一账户接口

如果需要迁移到统一账户接口，需要修改以下内容：

### 1. 修改BaseURL
```go
const (
    BinancePortfolioBaseURL = "https://papi.binance.com"
)
```

### 2. 修改接口端点
```go
const (
    ExchangeInfoEndpoint = "/papi/v1/um/exchangeInfo"  // um表示U本位
    OrderEndpoint = "/papi/v1/um/order"
    OrderQueryEndpoint = "/papi/v1/um/order"
    TickerPriceEndpoint = "/papi/v1/um/ticker/price"
)
```

### 3. 错误处理增强

需要特别处理HTTP 503错误的不同类型：

```go
func handle503Error(body []byte) error {
    errMsg := string(body)
    
    if strings.Contains(errMsg, "Unknown error") {
        // 执行状态未知，需要查询订单状态
        return ErrUnknownExecutionStatus
    }
    
    if strings.Contains(errMsg, "Service Unavailable") {
        // 服务不可用，可以退避重试
        return ErrServiceUnavailable
    }
    
    if strings.Contains(errMsg, "Request throttled") {
        // 系统限流，需要降低并发
        return ErrRequestThrottled
    }
    
    return fmt.Errorf("HTTP 503: %s", errMsg)
}
```

### 4. 重试逻辑

对于503错误，需要实现退避重试：

```go
func retryWithBackoff(fn func() error, maxRetries int) error {
    backoff := 200 * time.Millisecond
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if isRetryableError(err) {
            time.Sleep(backoff)
            backoff *= 2 // 指数退避
            continue
        }
        
        return err
    }
    return fmt.Errorf("重试 %d 次后仍失败", maxRetries)
}
```

## 配置建议

在配置文件中可以支持两种模式：

```yaml
binance:
  api_key: "your_api_key"
  secret_key: "your_secret_key"
  # 接口类型：fapi (U本位合约) 或 papi (统一账户)
  api_type: "fapi"  # 或 "papi"
  base_url: ""      # 可选，留空使用默认值
```

## 注意事项

1. **账户类型**：需要确保API密钥对应的账户已开通统一账户功能
2. **接口兼容性**：统一账户接口与fapi接口参数基本一致，但响应格式可能有细微差异
3. **错误处理**：必须正确处理503错误的不同类型，避免重复下单
4. **限流策略**：统一账户对限流更严格，需要合理控制请求频率
5. **平仓订单优先**：`closePosition=true`的订单不受-1008限流影响，适合用于止损止盈

## 参考文档

- [币安统一账户API文档](https://developers.binance.com/docs/zh-CN/derivatives/portfolio-margin/general-info)
- [错误代码说明](https://developers.binance.com/docs/derivatives/portfolio-margin/error-code)

