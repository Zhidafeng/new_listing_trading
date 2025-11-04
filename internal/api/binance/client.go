package binance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"new_listing_trade/internal/models"
)

const (
	// BinanceFuturesBaseURL 币安期货API基础URL（U本位合约）
	BinanceFuturesBaseURL = "https://fapi.binance.com"
	// BinancePortfolioBaseURL 币安统一账户API基础URL
	BinancePortfolioBaseURL = "https://papi.binance.com"

	// FAPI端点（U本位合约）
	FAPIExchangeInfoEndpoint = "/fapi/v1/exchangeInfo"
	FAPIOrderEndpoint        = "/fapi/v1/order"
	FAPITickerPriceEndpoint  = "/fapi/v1/ticker/price"

	// PAPI端点（统一账户U本位合约）
	PAPIExchangeInfoEndpoint     = "/papi/v1/um/exchangeInfo"
	PAPIOrderEndpoint            = "/papi/v1/um/order"
	PAPITickerPriceEndpoint      = "/papi/v1/um/ticker/price"
	PAPIConditionalOrderEndpoint = "/papi/v1/um/conditional/order" // 统一账户条件单接口
)

// APIError 自定义API错误
type APIError struct {
	Code            int    `json:"code"`
	Msg             string `json:"msg"`
	StatusCode      int    // HTTP状态码
	IsRetryable     bool   // 是否可重试
	IsUnknownStatus bool   // 是否为未知执行状态
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API错误: code=%d, msg=%s, status=%d", e.Code, e.Msg, e.StatusCode)
}

// Client 币安期货API客户端
type Client struct {
	baseURL    string
	apiKey     string
	secretKey  string
	apiType    string // "fapi" 或 "papi"
	httpClient *http.Client
}

// NewClient 创建新的币安期货API客户端（公开接口，默认fapi）
func NewClient() *Client {
	return &Client{
		baseURL: BinanceFuturesBaseURL,
		apiType: "fapi",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientWithAuth 创建带认证的币安期货API客户端
func NewClientWithAuth(apiKey, secretKey string) *Client {
	return &Client{
		baseURL:   BinanceFuturesBaseURL,
		apiKey:    apiKey,
		secretKey: secretKey,
		apiType:   "fapi",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientWithConfig 根据配置创建客户端
func NewClientWithConfig(apiKey, secretKey, apiType, baseURL string) *Client {
	client := &Client{
		apiKey:    apiKey,
		secretKey: secretKey,
		apiType:   apiType,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// 设置baseURL
	if baseURL != "" {
		client.baseURL = baseURL
	} else {
		// 根据apiType设置默认baseURL
		if apiType == "papi" {
			client.baseURL = BinancePortfolioBaseURL
		} else {
			client.baseURL = BinanceFuturesBaseURL
		}
	}

	return client
}

// getEndpoint 根据API类型获取端点路径
// 注意：exchangeInfo和tickerPrice接口始终使用fapi，不在此函数中处理
func (c *Client) getEndpoint(endpoint string) string {
	if c.apiType == "papi" {
		// 统一账户接口路径映射
		switch endpoint {
		case "exchangeInfo":
			// exchangeInfo不使用此映射，直接使用fapi
			return PAPIExchangeInfoEndpoint
		case "order":
			return PAPIOrderEndpoint
		case "tickerPrice":
			// tickerPrice不使用此映射，直接使用fapi
			return PAPITickerPriceEndpoint
		}
	} else {
		// fapi接口路径
		switch endpoint {
		case "exchangeInfo":
			// exchangeInfo不使用此映射，直接使用fapi
			return FAPIExchangeInfoEndpoint
		case "order":
			return FAPIOrderEndpoint
		case "tickerPrice":
			// tickerPrice不使用此映射，直接使用fapi
			return FAPITickerPriceEndpoint
		}
	}
	return endpoint
}

// handleHTTPError 处理HTTP错误响应
func (c *Client) handleHTTPError(statusCode int, body []byte) error {
	bodyStr := string(body)

	// 解析错误响应
	var errResp models.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil {
		apiErr := &APIError{
			Code:       errResp.Code,
			Msg:        errResp.Msg,
			StatusCode: statusCode,
		}

		// 处理HTTP 503错误的不同类型
		if statusCode == http.StatusServiceUnavailable {
			if strings.Contains(errResp.Msg, "Unknown error") {
				apiErr.IsUnknownStatus = true
				apiErr.IsRetryable = false // 需要先查询状态，不能直接重试
			} else if strings.Contains(errResp.Msg, "Service Unavailable") {
				apiErr.IsRetryable = true
			} else if strings.Contains(errResp.Msg, "Request throttled") || errResp.Code == -1008 {
				apiErr.IsRetryable = true
			}
		} else if statusCode == http.StatusTooManyRequests {
			// HTTP 429 限流
			apiErr.IsRetryable = true
		}

		return apiErr
	}

	return fmt.Errorf("API返回错误状态码 %d: %s", statusCode, bodyStr)
}

// NewClientWithBaseURL 使用自定义基础URL创建客户端
func NewClientWithBaseURL(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		apiType: "fapi",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetExchangeInfo 获取交易所信息（始终使用fapi接口，无论papi还是fapi）
func (c *Client) GetExchangeInfo() (*models.ExchangeInfo, error) {
	// exchangeInfo接口统一使用fapi，不需要区分papi/fapi
	url := fmt.Sprintf("%s%s", BinanceFuturesBaseURL, FAPIExchangeInfoEndpoint)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp.StatusCode, body)
	}

	var exchangeInfo models.ExchangeInfo
	if err := json.Unmarshal(body, &exchangeInfo); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &exchangeInfo, nil
}

// CreateOrder 创建订单（下单，带重试机制）
func (c *Client) CreateOrder(req *models.OrderRequest) (*models.OrderResponse, error) {
	if c.apiKey == "" || c.secretKey == "" {
		return nil, fmt.Errorf("API密钥和密钥未设置，请使用NewClientWithAuth创建客户端")
	}

	var result *models.OrderResponse

	// 使用重试机制
	err := RetryWithBackoff(func() error {
		// 每次重试时更新时间戳
		reqCopy := *req
		reqCopy.Timestamp = time.Now().UnixMilli()

		resp, err := c.createOrderInternal(&reqCopy)
		if err != nil {
			return err
		}
		result = resp
		return nil
	}, DefaultRetryConfig())

	if err != nil {
		return nil, err
	}

	return result, nil
}

// createOrderInternal 创建订单的内部实现
func (c *Client) createOrderInternal(req *models.OrderRequest) (*models.OrderResponse, error) {
	// 设置时间戳，使用当前时间
	req.Timestamp = time.Now().UnixMilli()

	// 如果没有设置recvWindow，设置默认值为10000ms（10秒），避免时间戳超出窗口
	if req.RecvWindow == 0 {
		req.RecvWindow = 10000 // 10秒窗口
	}

	// 构建参数字典
	params := make(map[string]string)
	params["symbol"] = req.Symbol
	params["side"] = req.Side
	params["type"] = req.Type

	if req.TimeInForce != "" {
		params["timeInForce"] = req.TimeInForce
	}
	if req.Quantity != "" {
		params["quantity"] = req.Quantity
	}
	if req.Notional != "" {
		params["notional"] = req.Notional
	}
	if req.Price != "" {
		params["price"] = req.Price
	}
	if req.StopPrice != "" {
		params["stopPrice"] = req.StopPrice
	}
	if req.ReduceOnly != "" {
		params["reduceOnly"] = req.ReduceOnly
	}
	if req.ClosePosition != "" {
		params["closePosition"] = req.ClosePosition
	}
	if req.PositionSide != "" {
		params["positionSide"] = req.PositionSide
	}
	if req.CallbackRate != "" {
		params["callbackRate"] = req.CallbackRate
	}
	if req.WorkingType != "" {
		params["workingType"] = req.WorkingType
	}
	if req.PriceProtect != "" {
		params["priceProtect"] = req.PriceProtect
	}
	if req.NewOrderRespType != "" {
		params["newOrderRespType"] = req.NewOrderRespType
	}
	params["recvWindow"] = strconv.FormatInt(req.RecvWindow, 10)
	params["timestamp"] = strconv.FormatInt(req.Timestamp, 10)

	// 构建查询字符串
	queryString := BuildQueryString(params)

	// 签名（对查询字符串进行HMAC SHA256签名）
	signature := signQueryString(queryString, c.secretKey)
	// 重新构建包含签名的查询字符串
	queryStringWithSig := queryString + "&signature=" + signature

	// 构建URL
	endpoint := c.getEndpoint("order")
	requestURL := fmt.Sprintf("%s%s?%s", c.baseURL, endpoint, queryStringWithSig)

	// 创建POST请求
	httpReq, err := http.NewRequest(http.MethodPost, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("X-MBX-APIKEY", c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查错误响应
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp.StatusCode, body)
	}

	// 解析响应
	var orderResp models.OrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &orderResp, nil
}

// QueryOrder 查询订单
func (c *Client) QueryOrder(params *models.OrderQueryParams) (*models.OrderResponse, error) {
	if c.apiKey == "" || c.secretKey == "" {
		return nil, fmt.Errorf("API密钥和密钥未设置，请使用NewClientWithAuth创建客户端")
	}

	// 设置时间戳（如果未设置）
	if params.Timestamp == 0 {
		params.Timestamp = time.Now().UnixMilli()
	}

	// 如果没有设置recvWindow，设置默认值为10000ms（10秒）
	if params.RecvWindow == 0 {
		params.RecvWindow = 10000
	}

	// 构建参数字典
	queryParams := make(map[string]string)
	queryParams["symbol"] = params.Symbol

	if params.OrderID > 0 {
		queryParams["orderId"] = strconv.FormatInt(params.OrderID, 10)
	}
	if params.OrigClientOrderID != "" {
		queryParams["origClientOrderId"] = params.OrigClientOrderID
	}
	queryParams["recvWindow"] = strconv.FormatInt(params.RecvWindow, 10)
	queryParams["timestamp"] = strconv.FormatInt(params.Timestamp, 10)

	// 构建查询字符串
	queryString := BuildQueryString(queryParams)

	// 签名（对查询字符串进行HMAC SHA256签名）
	signature := signQueryString(queryString, c.secretKey)
	// 重新构建包含签名的查询字符串
	queryStringWithSig := queryString + "&signature=" + signature

	// 构建URL
	endpoint := c.getEndpoint("order")
	requestURL := fmt.Sprintf("%s%s?%s", c.baseURL, endpoint, queryStringWithSig)

	// 创建GET请求
	httpReq, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("X-MBX-APIKEY", c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查错误响应
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp.StatusCode, body)
	}

	// 解析响应
	var orderResp models.OrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &orderResp, nil
}

// CreateConditionalOrder 创建条件单（统一账户专用，带重试机制）
func (c *Client) CreateConditionalOrder(req *models.ConditionalOrderRequest) (*models.ConditionalOrderResponse, error) {
	if c.apiKey == "" || c.secretKey == "" {
		return nil, fmt.Errorf("API密钥和密钥未设置，请使用NewClientWithAuth创建客户端")
	}

	if c.apiType != "papi" {
		return nil, fmt.Errorf("条件单接口仅适用于统一账户(papi)，当前API类型: %s", c.apiType)
	}

	var result *models.ConditionalOrderResponse

	// 使用重试机制
	err := RetryWithBackoff(func() error {
		// 每次重试时更新时间戳
		reqCopy := *req
		reqCopy.Timestamp = time.Now().UnixMilli()

		resp, err := c.createConditionalOrderInternal(&reqCopy)
		if err != nil {
			return err
		}
		result = resp
		return nil
	}, DefaultRetryConfig())

	if err != nil {
		return nil, err
	}

	return result, nil
}

// createConditionalOrderInternal 创建条件单的内部实现
func (c *Client) createConditionalOrderInternal(req *models.ConditionalOrderRequest) (*models.ConditionalOrderResponse, error) {
	// 设置时间戳，使用当前时间
	req.Timestamp = time.Now().UnixMilli()

	// 如果没有设置recvWindow，设置默认值为10000ms（10秒）
	if req.RecvWindow == 0 {
		req.RecvWindow = 10000 // 10秒窗口
	}

	// 构建参数字典
	params := make(map[string]string)
	params["symbol"] = req.Symbol
	params["side"] = req.Side
	params["strategyType"] = req.StrategyType

	if req.PositionSide != "" {
		params["positionSide"] = req.PositionSide
	}
	if req.TimeInForce != "" {
		params["timeInForce"] = req.TimeInForce
	}
	if req.Quantity != "" {
		params["quantity"] = req.Quantity
	}
	if req.ReduceOnly != "" {
		params["reduceOnly"] = req.ReduceOnly
	}
	if req.Price != "" {
		params["price"] = req.Price
	}
	if req.WorkingType != "" {
		params["workingType"] = req.WorkingType
	}
	if req.PriceProtect != "" {
		params["priceProtect"] = req.PriceProtect
	}
	if req.NewClientStrategyId != "" {
		params["newClientStrategyId"] = req.NewClientStrategyId
	}
	if req.StopPrice != "" {
		params["stopPrice"] = req.StopPrice
	}
	if req.ActivationPrice != "" {
		params["activationPrice"] = req.ActivationPrice
	}
	if req.CallbackRate != "" {
		params["callbackRate"] = req.CallbackRate
	}
	if req.PriceMatch != "" {
		params["priceMatch"] = req.PriceMatch
	}
	if req.SelfTradePreventionMode != "" {
		params["selfTradePreventionMode"] = req.SelfTradePreventionMode
	}
	if req.GoodTillDate > 0 {
		params["goodTillDate"] = strconv.FormatInt(req.GoodTillDate, 10)
	}
	params["recvWindow"] = strconv.FormatInt(req.RecvWindow, 10)
	params["timestamp"] = strconv.FormatInt(req.Timestamp, 10)

	// 构建查询字符串
	queryString := BuildQueryString(params)

	// 签名（对查询字符串进行HMAC SHA256签名）
	signature := signQueryString(queryString, c.secretKey)
	// 重新构建包含签名的查询字符串
	queryStringWithSig := queryString + "&signature=" + signature

	// 构建URL（条件单接口）
	requestURL := fmt.Sprintf("%s%s?%s", c.baseURL, PAPIConditionalOrderEndpoint, queryStringWithSig)

	// 创建POST请求
	httpReq, err := http.NewRequest(http.MethodPost, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("X-MBX-APIKEY", c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查错误响应
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp.StatusCode, body)
	}

	// 解析响应
	var condOrderResp models.ConditionalOrderResponse
	if err := json.Unmarshal(body, &condOrderResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &condOrderResp, nil
}

// GetTickerPrice 获取指定交易对的当前价格（始终使用fapi接口，无论papi还是fapi）
func (c *Client) GetTickerPrice(symbol string) (*models.TickerPrice, error) {
	// tickerPrice接口统一使用fapi，不需要区分papi/fapi
	url := fmt.Sprintf("%s%s?symbol=%s", BinanceFuturesBaseURL, FAPITickerPriceEndpoint, symbol)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp.StatusCode, body)
	}

	var tickerPrice models.TickerPrice
	if err := json.Unmarshal(body, &tickerPrice); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &tickerPrice, nil
}

// signQueryString 对查询字符串进行HMAC SHA256签名
func signQueryString(queryString, secretKey string) string {
	return SignQueryString(queryString, secretKey)
}
