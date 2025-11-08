package service

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	"new_listing_trade/internal/api/binance"
	"new_listing_trade/internal/config"
	"new_listing_trade/internal/logger"
	"new_listing_trade/internal/models"
)

// TradingService 交易服务
type TradingService struct {
	client *binance.Client
	config *config.Config
	mu     sync.RWMutex
}

// NewTradingService 创建交易服务
func NewTradingService(cfg *config.Config) (*TradingService, error) {
	if cfg.Binance.APIKey == "" || cfg.Binance.SecretKey == "" {
		return nil, fmt.Errorf("币安API密钥未配置")
	}

	// 根据配置创建客户端
	apiType := cfg.Binance.APIType
	if apiType == "" {
		apiType = "fapi" // 默认使用fapi
	}
	client := binance.NewClientWithConfig(cfg.Binance.APIKey, cfg.Binance.SecretKey, apiType, cfg.Binance.BaseURL)

	// 记录使用的API类型和baseURL
	if apiType == "papi" {
		logger.Infof("使用统一账户接口 (papi): %s", binance.BinancePortfolioBaseURL)
	} else {
		logger.Infof("使用U本位合约接口 (fapi): %s", binance.BinanceFuturesBaseURL)
	}

	return &TradingService{
		client: client,
		config: cfg,
	}, nil
}

// CreateMarketSellOrder 创建市价卖单（做空，按USDT金额）
func (ts *TradingService) CreateMarketSellOrder(symbol string, notionalUSDT string) (*models.OrderResponse, error) {
	if notionalUSDT == "" {
		notionalUSDT = ts.config.Trading.DefaultNotional
	}

	// 检查是否使用统一账户接口（papi），统一账户接口不支持notional参数，需要使用quantity
	apiType := ts.config.Binance.APIType
	if apiType == "" {
		apiType = "fapi"
	}

	req := &models.OrderRequest{
		Symbol:       symbol,
		Side:         "SELL", // 做空
		Type:         "MARKET",
		PositionSide: ts.config.Trading.PositionSide,
	}

	// 如果是统一账户接口（papi），需要将notional转换为quantity
	if apiType == "papi" {
		// 获取交易所信息，用于获取交易对精度规则
		exchangeInfo, err := ts.client.GetExchangeInfo()
		if err != nil {
			return nil, fmt.Errorf("获取交易所信息失败: %w", err)
		}

		// 获取交易对信息
		symbolInfo, err := binance.GetSymbolInfo(exchangeInfo, symbol)
		if err != nil {
			return nil, fmt.Errorf("获取交易对信息失败: %w", err)
		}

		// 获取当前价格
		tickerPrice, err := ts.client.GetTickerPrice(symbol)
		if err != nil {
			return nil, fmt.Errorf("获取当前价格失败，无法计算quantity: %w", err)
		}

		// 计算quantity
		notionalFloat, err := strconv.ParseFloat(notionalUSDT, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的USDT金额: %s", notionalUSDT)
		}

		priceFloat, err := strconv.ParseFloat(tickerPrice.Price, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的价格: %s", tickerPrice.Price)
		}

		if priceFloat <= 0 {
			return nil, fmt.Errorf("价格无效: %f", priceFloat)
		}

		// quantity = notional / price
		quantityFloat := notionalFloat / priceFloat

		// 验证并调整quantity精度
		quantityStr, err := binance.ValidateAndAdjustQuantity(quantityFloat, symbolInfo)
		if err != nil {
			return nil, fmt.Errorf("调整quantity精度失败: %w", err)
		}

		req.Quantity = quantityStr

		logger.Infof("创建市价卖单（做空，统一账户）: %s, USDT金额: %s, 当前价格: %s, 计算数量: %s",
			symbol, notionalUSDT, tickerPrice.Price, req.Quantity)
	} else {
		// fapi接口可以使用notional
		req.Notional = notionalUSDT
		logger.Infof("创建市价卖单（做空）: %s, USDT金额: %s", symbol, notionalUSDT)
	}

	return ts.client.CreateOrder(req)
}

// convertConditionalOrderToOrderResponse 将条件单响应转换为订单响应（用于兼容性）
func convertConditionalOrderToOrderResponse(condResp *models.ConditionalOrderResponse) *models.OrderResponse {
	return &models.OrderResponse{
		OrderID:       condResp.StrategyID,
		Symbol:        condResp.Symbol,
		Status:        condResp.StrategyStatus,
		ClientOrderID: condResp.NewClientStrategyId,
		Price:         condResp.Price,
		OrigQty:       condResp.OrigQty,
		TimeInForce:   condResp.TimeInForce,
		Type:          condResp.StrategyType,
		ReduceOnly:    condResp.ReduceOnly,
		Side:          condResp.Side,
		PositionSide:  condResp.PositionSide,
		StopPrice:     condResp.StopPrice,
		WorkingType:   condResp.WorkingType,
		PriceProtect:  condResp.PriceProtect,
		UpdateTime:    condResp.UpdateTime,
		Time:          condResp.BookTime,
	}
}

// CreateStopLossOrder 创建止损订单（做空时，价格上涨触发止损）
// quantity: 平仓数量（仅papi需要，可选参数）
func (ts *TradingService) CreateStopLossOrder(symbol string, entryPrice float64, quantity ...string) (*models.OrderResponse, error) {
	if !ts.config.Trading.StopLoss.Enabled {
		return nil, fmt.Errorf("止损功能未启用")
	}

	// 计算止损价格（做空时，止损价高于开仓价）
	stopLossPercent := ts.config.Trading.StopLoss.Percent / 100.0
	stopLossPrice := entryPrice * (1 + stopLossPercent) // 做空时价格上涨触发止损

	// 获取交易所信息，用于获取交易对精度规则
	exchangeInfo, err := ts.client.GetExchangeInfo()
	if err != nil {
		return nil, fmt.Errorf("获取交易所信息失败: %w", err)
	}

	// 获取交易对信息
	symbolInfo, err := binance.GetSymbolInfo(exchangeInfo, symbol)
	if err != nil {
		return nil, fmt.Errorf("获取交易对信息失败: %w", err)
	}

	// 调整stopPrice精度
	stopPriceStr, err := binance.ValidateAndAdjustPrice(stopLossPrice, symbolInfo)
	if err != nil {
		return nil, fmt.Errorf("调整止损价格精度失败: %w", err)
	}

	// 检查API类型，统一账户接口使用条件单接口
	apiType := ts.config.Binance.APIType
	if apiType == "" {
		apiType = "fapi"
	}

	// 统一账户接口使用条件单接口
	if apiType == "papi" {
		// 统一账户条件单接口需要quantity参数
		quantityStr := ""
		if len(quantity) > 0 && quantity[0] != "" {
			quantityStr = quantity[0]
		} else {
			return nil, fmt.Errorf("统一账户条件单需要quantity参数，请提供平仓数量")
		}

		// 使用条件单接口，STOP_MARKET类型需要stopPrice和quantity
		condReq := &models.ConditionalOrderRequest{
			Symbol:       symbol,
			Side:         "BUY", // 止损是买入平仓（做空的反向操作）
			StrategyType: "STOP_MARKET",
			StopPrice:    stopPriceStr,
			Quantity:     quantityStr,
			ReduceOnly:   "true", // 只减仓，确保平仓
			PositionSide: ts.config.Trading.PositionSide,
			WorkingType:  ts.config.Trading.StopLoss.WorkingType,
			PriceProtect: "TRUE",
		}

		logger.Infof("创建止损条件单（做空，统一账户）: %s, 开仓价格: %.8f, 止损价格: %s, 数量: %s, 止损百分比: %.2f%%, 策略类型: STOP_MARKET",
			symbol, entryPrice, stopPriceStr, quantityStr, ts.config.Trading.StopLoss.Percent)

		condResp, err := ts.client.CreateConditionalOrder(condReq)
		if err != nil {
			return nil, err
		}

		// 转换为OrderResponse以保持兼容性
		return convertConditionalOrderToOrderResponse(condResp), nil
	} else {
		// fapi接口使用普通订单接口
		req := &models.OrderRequest{
			Symbol:        symbol,
			Side:          "BUY", // 止损是买入平仓（做空的反向操作）
			StopPrice:     stopPriceStr,
			ClosePosition: "true", // 平仓，自动平掉整个持仓（不需要quantity）
			PositionSide:  ts.config.Trading.PositionSide,
			WorkingType:   ts.config.Trading.StopLoss.WorkingType,
			PriceProtect:  "true",
			Type:          "STOP_MARKET",
		}

		logger.Infof("创建止损订单（做空，fapi）: %s, 开仓价格: %.8f, 止损价格: %s, 止损百分比: %.2f%%, 订单类型: STOP_MARKET",
			symbol, entryPrice, stopPriceStr, ts.config.Trading.StopLoss.Percent)
		return ts.client.CreateOrder(req)
	}
}

// CreateTakeProfitOrder 创建止盈订单（做空时，价格下跌触发止盈）
// quantity: 平仓数量（仅papi需要，可选参数）
func (ts *TradingService) CreateTakeProfitOrder(symbol string, entryPrice float64, quantity ...string) (*models.OrderResponse, error) {
	if !ts.config.Trading.TakeProfit.Enabled {
		return nil, fmt.Errorf("止盈功能未启用")
	}

	// 计算止盈价格（做空时，止盈价低于开仓价）
	takeProfitPercent := ts.config.Trading.TakeProfit.Percent / 100.0
	takeProfitPrice := entryPrice * (1 - takeProfitPercent) // 做空时价格下跌触发止盈

	// 获取交易所信息，用于获取交易对精度规则
	exchangeInfo, err := ts.client.GetExchangeInfo()
	if err != nil {
		return nil, fmt.Errorf("获取交易所信息失败: %w", err)
	}

	// 获取交易对信息
	symbolInfo, err := binance.GetSymbolInfo(exchangeInfo, symbol)
	if err != nil {
		return nil, fmt.Errorf("获取交易对信息失败: %w", err)
	}

	// 调整stopPrice精度
	stopPriceStr, err := binance.ValidateAndAdjustPrice(takeProfitPrice, symbolInfo)
	if err != nil {
		return nil, fmt.Errorf("调整止盈价格精度失败: %w", err)
	}

	// 检查API类型，统一账户接口使用条件单接口
	apiType := ts.config.Binance.APIType
	if apiType == "" {
		apiType = "fapi"
	}

	// 统一账户接口使用条件单接口
	if apiType == "papi" {
		// 统一账户条件单接口需要quantity参数
		quantityStr := ""
		if len(quantity) > 0 && quantity[0] != "" {
			quantityStr = quantity[0]
		} else {
			return nil, fmt.Errorf("统一账户条件单需要quantity参数，请提供平仓数量")
		}

		// 使用条件单接口，TAKE_PROFIT_MARKET类型需要stopPrice和quantity
		condReq := &models.ConditionalOrderRequest{
			Symbol:       symbol,
			Side:         "BUY", // 止盈是买入平仓（做空的反向操作）
			StrategyType: "TAKE_PROFIT_MARKET",
			StopPrice:    stopPriceStr,
			Quantity:     quantityStr,
			ReduceOnly:   "true", // 只减仓，确保平仓
			PositionSide: ts.config.Trading.PositionSide,
			WorkingType:  ts.config.Trading.TakeProfit.WorkingType,
			PriceProtect: "TRUE",
		}

		logger.Infof("创建止盈条件单（做空，统一账户）: %s, 开仓价格: %.8f, 止盈价格: %s, 数量: %s, 止盈百分比: %.2f%%, 策略类型: TAKE_PROFIT_MARKET",
			symbol, entryPrice, stopPriceStr, quantityStr, ts.config.Trading.TakeProfit.Percent)

		condResp, err := ts.client.CreateConditionalOrder(condReq)
		if err != nil {
			return nil, err
		}

		// 转换为OrderResponse以保持兼容性
		return convertConditionalOrderToOrderResponse(condResp), nil
	} else {
		// fapi接口使用普通订单接口
		req := &models.OrderRequest{
			Symbol:        symbol,
			Side:          "BUY", // 止盈是买入平仓（做空的反向操作）
			StopPrice:     stopPriceStr,
			ClosePosition: "true", // 平仓，自动平掉整个持仓（不需要quantity）
			PositionSide:  ts.config.Trading.PositionSide,
			WorkingType:   ts.config.Trading.TakeProfit.WorkingType,
			PriceProtect:  "true",
			Type:          "TAKE_PROFIT_MARKET",
		}

		logger.Infof("创建止盈订单（做空，fapi）: %s, 开仓价格: %.8f, 止盈价格: %s, 止盈百分比: %.2f%%, 订单类型: TAKE_PROFIT_MARKET",
			symbol, entryPrice, stopPriceStr, ts.config.Trading.TakeProfit.Percent)
		return ts.client.CreateOrder(req)
	}
}

// CreateOrdersWithStopLossAndTakeProfit 创建卖单（做空）并同时设置止损和止盈（按USDT金额）
func (ts *TradingService) CreateOrdersWithStopLossAndTakeProfit(symbol string, notionalUSDT string) (*OrderSet, error) {
	orderSet := &OrderSet{
		Symbol: symbol,
	}

	// 创建卖单（做空，按USDT金额）
	sellOrder, err := ts.CreateMarketSellOrder(symbol, notionalUSDT)
	if err != nil {
		return nil, fmt.Errorf("创建卖单失败: %w", err)
	}
	orderSet.SellOrder = sellOrder

	// 获取实际成交价格作为开仓价格
	entryPrice := 0.0
	if sellOrder.AvgPrice != "" {
		if avgPrice, err := strconv.ParseFloat(sellOrder.AvgPrice, 64); err == nil && avgPrice > 0 {
			entryPrice = avgPrice
		}
	}

	// 如果无法从订单响应获取价格，尝试获取当前价格
	if entryPrice == 0 {
		tickerPrice, err := ts.client.GetTickerPrice(symbol)
		if err != nil {
			logger.Warnf("获取当前价格失败: %v，将使用订单响应中的价格", err)
		} else if tickerPrice.Price != "" {
			if price, err := strconv.ParseFloat(tickerPrice.Price, 64); err == nil && price > 0 {
				entryPrice = price
			}
		}
	}

	if entryPrice == 0 {
		return nil, fmt.Errorf("无法获取开仓价格")
	}

	// 获取成交数量（用于统一账户条件单）
	executedQty := ""
	qtyFloat := 0.0

	// 优先使用 ExecutedQty
	if sellOrder.ExecutedQty != "" {
		if parsed, err := strconv.ParseFloat(sellOrder.ExecutedQty, 64); err == nil && parsed > 0 {
			qtyFloat = parsed
		}
	}

	// 如果 ExecutedQty 无效，尝试使用 OrigQty
	if qtyFloat == 0 && sellOrder.OrigQty != "" {
		if parsed, err := strconv.ParseFloat(sellOrder.OrigQty, 64); err == nil && parsed > 0 {
			qtyFloat = parsed
		}
	}

	// 如果数量还是0，尝试从 CumQuote（成交金额）和价格反推数量
	if qtyFloat == 0 && sellOrder.CumQuote != "" && entryPrice > 0 {
		if cumQuote, err := strconv.ParseFloat(sellOrder.CumQuote, 64); err == nil && cumQuote > 0 {
			qtyFloat = cumQuote / entryPrice
			logger.Infof("从成交金额反推数量: CumQuote=%s, 价格=%.8f, 数量=%.8f", sellOrder.CumQuote, entryPrice, qtyFloat)
		}
	}

	// 如果还是0，尝试从notionalUSDT和价格计算
	if qtyFloat == 0 && entryPrice > 0 {
		if notional, err := strconv.ParseFloat(notionalUSDT, 64); err == nil && notional > 0 {
			qtyFloat = notional / entryPrice
			logger.Infof("从USDT金额反推数量: notional=%s, 价格=%.8f, 数量=%.8f", notionalUSDT, entryPrice, qtyFloat)
		}
	}

	// 如果使用统一账户接口，需要验证并调整quantity精度
	apiType := ts.config.Binance.APIType
	if apiType == "" {
		apiType = "fapi"
	}

	if apiType == "papi" {
		if qtyFloat <= 0 {
			return nil, fmt.Errorf("无法获取有效的成交数量，ExecutedQty=%s, OrigQty=%s, CumQuote=%s",
				sellOrder.ExecutedQty, sellOrder.OrigQty, sellOrder.CumQuote)
		}

		// 确保数量为正数（使用绝对值）
		if qtyFloat < 0 {
			qtyFloat = -qtyFloat
		}

		// 获取交易所信息，用于获取交易对精度规则
		exchangeInfo, err := ts.client.GetExchangeInfo()
		if err != nil {
			return nil, fmt.Errorf("获取交易所信息失败: %w", err)
		}

		// 获取交易对信息
		symbolInfo, err := binance.GetSymbolInfo(exchangeInfo, symbol)
		if err != nil {
			return nil, fmt.Errorf("获取交易对信息失败: %w", err)
		}

		// 调整quantity精度
		adjustedQty, err := binance.ValidateAndAdjustQuantity(qtyFloat, symbolInfo)
		if err != nil {
			return nil, fmt.Errorf("调整数量精度失败: %w (原始数量: %.8f)", err, qtyFloat)
		}

		executedQty = adjustedQty
		logger.Infof("调整数量精度: %.8f -> %s", qtyFloat, executedQty)
	}

	// 创建止损订单
	if ts.config.Trading.StopLoss.Enabled {
		var stopLossOrder *models.OrderResponse
		var err error
		if apiType == "papi" && executedQty != "" {
			stopLossOrder, err = ts.CreateStopLossOrder(symbol, entryPrice, executedQty)
		} else {
			stopLossOrder, err = ts.CreateStopLossOrder(symbol, entryPrice)
		}
		if err != nil {
			logger.Errorf("创建止损订单失败: %v", err)
			orderSet.StopLossError = err
		} else {
			orderSet.StopLossOrder = stopLossOrder
		}
	}

	// 创建止盈订单
	if ts.config.Trading.TakeProfit.Enabled {
		var takeProfitOrder *models.OrderResponse
		var err error
		if apiType == "papi" && executedQty != "" {
			takeProfitOrder, err = ts.CreateTakeProfitOrder(symbol, entryPrice, executedQty)
		} else {
			takeProfitOrder, err = ts.CreateTakeProfitOrder(symbol, entryPrice)
		}
		if err != nil {
			logger.Errorf("创建止盈订单失败: %v", err)
			orderSet.TakeProfitError = err
		} else {
			orderSet.TakeProfitOrder = takeProfitOrder
		}
	}

	return orderSet, nil
}

// QueryOrder 查询订单
func (ts *TradingService) QueryOrder(symbol string, orderID int64) (*models.OrderResponse, error) {
	params := &models.OrderQueryParams{
		Symbol:  symbol,
		OrderID: orderID,
	}
	return ts.client.QueryOrder(params)
}

// GetNegativePositions 查询收益为负的仓位并排序
func (ts *TradingService) GetNegativePositions() (*models.NegativePositionResponse, error) {
	// 查询所有持仓（symbol为空表示查询所有）
	positionRisks, err := ts.client.GetPositionRisk("")
	if err != nil {
		return nil, fmt.Errorf("查询持仓失败: %w", err)
	}

	// 转换为Position模型并筛选负收益仓位
	negativePositions := make([]models.Position, 0)
	for _, pr := range positionRisks {
		// 解析持仓数量
		positionAmt, err := strconv.ParseFloat(pr.PositionAmt, 64)
		if err != nil || positionAmt == 0 {
			// 跳过持仓数量为0的仓位
			continue
		}

		// 解析未实现盈亏
		unRealizedProfit, err := strconv.ParseFloat(pr.UnRealizedProfit, 64)
		if err != nil {
			logger.Warnf("解析未实现盈亏失败: %s, symbol: %s", pr.UnRealizedProfit, pr.Symbol)
			continue
		}

		// 只保留负收益的仓位
		if unRealizedProfit >= 0 {
			continue
		}

		// 解析其他字段
		entryPrice, _ := strconv.ParseFloat(pr.EntryPrice, 64)
		markPrice, _ := strconv.ParseFloat(pr.MarkPrice, 64)
		liquidationPrice, _ := strconv.ParseFloat(pr.LiquidationPrice, 64)
		leverage, _ := strconv.ParseFloat(pr.Leverage, 64)
		notional, _ := strconv.ParseFloat(pr.Notional, 64)

		// 计算收益率百分比
		profitPercent := 0.0
		if entryPrice > 0 && notional > 0 {
			// 收益率 = 未实现盈亏 / 持仓名义价值 * 100
			profitPercent = (unRealizedProfit / notional) * 100
		}

		position := models.Position{
			Symbol:           pr.Symbol,
			PositionAmt:      positionAmt,
			EntryPrice:       entryPrice,
			MarkPrice:        markPrice,
			UnRealizedProfit: unRealizedProfit,
			LiquidationPrice: liquidationPrice,
			Leverage:         leverage,
			MarginType:       pr.MarginType,
			PositionSide:     pr.PositionSide,
			Notional:         notional,
			UpdateTime:       pr.UpdateTime,
			ProfitPercent:    profitPercent,
		}

		negativePositions = append(negativePositions, position)
	}

	// 按亏损金额从大到小排序（未实现盈亏越小，亏损越大）
	sort.Slice(negativePositions, func(i, j int) bool {
		return negativePositions[i].UnRealizedProfit < negativePositions[j].UnRealizedProfit
	})

	logger.Infof("查询到 %d 个负收益仓位", len(negativePositions))

	return &models.NegativePositionResponse{
		TotalCount: len(negativePositions),
		Positions:  negativePositions,
	}, nil
}

// OrderSet 订单集合（卖单+止损+止盈）
type OrderSet struct {
	Symbol          string
	SellOrder       *models.OrderResponse // 做空卖单
	StopLossOrder   *models.OrderResponse
	TakeProfitOrder *models.OrderResponse
	StopLossError   error
	TakeProfitError error
}
