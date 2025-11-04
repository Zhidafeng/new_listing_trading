package models

// OrderRequest 下单请求
type OrderRequest struct {
	Symbol           string `json:"symbol"`                     // 交易对
	Side             string `json:"side"`                       // 买卖方向 BUY/SELL
	Type             string `json:"type"`                       // 订单类型 LIMIT/MARKET/STOP/TAKE_PROFIT等
	TimeInForce      string `json:"timeInForce"`                // 有效方式 GTC/IOC/FOK
	Quantity         string `json:"quantity,omitempty"`         // 数量
	Notional         string `json:"notional,omitempty"`         // USDT金额（用于市价单）
	Price            string `json:"price,omitempty"`            // 价格
	StopPrice        string `json:"stopPrice,omitempty"`        // 止损价格
	ReduceOnly       string `json:"reduceOnly,omitempty"`       // 只减仓标识
	ClosePosition    string `json:"closePosition,omitempty"`    // 平仓标识
	PositionSide     string `json:"positionSide,omitempty"`     // 持仓方向 BOTH/LONG/SHORT
	CallbackRate     string `json:"callbackRate,omitempty"`     // 触发价格百分比
	WorkingType      string `json:"workingType,omitempty"`      // 触发类型
	PriceProtect     string `json:"priceProtect,omitempty"`     // 价格保护
	NewOrderRespType string `json:"newOrderRespType,omitempty"` // 响应类型
	RecvWindow       int64  `json:"recvWindow,omitempty"`       // 接收窗口
	Timestamp        int64  `json:"timestamp"`                  // 时间戳
}

// OrderResponse 下单响应
type OrderResponse struct {
	OrderID                 int64  `json:"orderId"`
	Symbol                  string `json:"symbol"`
	Status                  string `json:"status"`
	ClientOrderID           string `json:"clientOrderId"`
	Price                   string `json:"price"`
	AvgPrice                string `json:"avgPrice"`
	OrigQty                 string `json:"origQty"`
	ExecutedQty             string `json:"executedQty"`
	CumQuote                string `json:"cumQuote"`
	TimeInForce             string `json:"timeInForce"`
	Type                    string `json:"type"`
	ReduceOnly              bool   `json:"reduceOnly"`
	ClosePosition           bool   `json:"closePosition"`
	Side                    string `json:"side"`
	PositionSide            string `json:"positionSide"`
	StopPrice               string `json:"stopPrice"`
	WorkingType             string `json:"workingType"`
	PriceProtect            bool   `json:"priceProtect"`
	OrigType                string `json:"origType"`
	PriceMatch              string `json:"priceMatch"`
	SelfTradePreventionMode string `json:"selfTradePreventionMode"`
	GoodTillDate            int64  `json:"goodTillDate"`
	Time                    int64  `json:"time"`
	UpdateTime              int64  `json:"updateTime"`
}

// OrderQueryParams 查询订单参数
type OrderQueryParams struct {
	Symbol            string `json:"symbol"`                      // 交易对
	OrderID           int64  `json:"orderId,omitempty"`           // 订单ID
	OrigClientOrderID string `json:"origClientOrderId,omitempty"` // 原始客户端订单ID
	RecvWindow        int64  `json:"recvWindow,omitempty"`        // 接收窗口
	Timestamp         int64  `json:"timestamp"`                   // 时间戳
}

// ConditionalOrderRequest 条件单请求（统一账户专用）
type ConditionalOrderRequest struct {
	Symbol                  string `json:"symbol"`                            // 交易对
	Side                    string `json:"side"`                              // 买卖方向 BUY/SELL
	PositionSide            string `json:"positionSide,omitempty"`            // 持仓方向 BOTH/LONG/SHORT
	StrategyType            string `json:"strategyType"`                      // 条件单类型 STOP/STOP_MARKET/TAKE_PROFIT/TAKE_PROFIT_MARKET
	TimeInForce             string `json:"timeInForce,omitempty"`             // 有效方式 GTC/IOC/FOK/GTD
	Quantity                string `json:"quantity,omitempty"`                // 数量（STOP/TAKE_PROFIT 必需）
	ReduceOnly              string `json:"reduceOnly,omitempty"`              // 只减仓标识 true/false
	Price                   string `json:"price,omitempty"`                   // 价格（STOP/TAKE_PROFIT 必需）
	WorkingType             string `json:"workingType,omitempty"`             // 触发类型 MARK_PRICE/CONTRACT_PRICE
	PriceProtect            string `json:"priceProtect,omitempty"`            // 价格保护 TRUE/FALSE
	NewClientStrategyId     string `json:"newClientStrategyId,omitempty"`     // 用户自定义订单号
	StopPrice               string `json:"stopPrice,omitempty"`               // 触发价格
	ActivationPrice         string `json:"activationPrice,omitempty"`         // TRAILING_STOP_MARKET 激活价格
	CallbackRate            string `json:"callbackRate,omitempty"`            // TRAILING_STOP_MARKET 回调率
	PriceMatch              string `json:"priceMatch,omitempty"`              // 价格匹配模式
	SelfTradePreventionMode string `json:"selfTradePreventionMode,omitempty"` // 自成交保护模式
	GoodTillDate            int64  `json:"goodTillDate,omitempty"`            // GTD 自动取消时间
	RecvWindow              int64  `json:"recvWindow,omitempty"`              // 接收窗口
	Timestamp               int64  `json:"timestamp"`                         // 时间戳
}

// ConditionalOrderResponse 条件单响应（统一账户专用）
type ConditionalOrderResponse struct {
	NewClientStrategyId     string `json:"newClientStrategyId"`     // 用户自定义订单号
	StrategyID              int64  `json:"strategyId"`              // 策略ID
	StrategyStatus          string `json:"strategyStatus"`          // 策略状态 NEW/TRIGGERED/CANCELLED等
	StrategyType            string `json:"strategyType"`            // 策略类型
	OrigQty                 string `json:"origQty"`                 // 原始数量
	Price                   string `json:"price"`                   // 价格
	ReduceOnly              bool   `json:"reduceOnly"`              // 只减仓
	Side                    string `json:"side"`                    // 买卖方向
	PositionSide            string `json:"positionSide"`            // 持仓方向
	StopPrice               string `json:"stopPrice"`               // 触发价格
	Symbol                  string `json:"symbol"`                  // 交易对
	TimeInForce             string `json:"timeInForce"`             // 有效方式
	ActivatePrice           string `json:"activatePrice"`           // 激活价格
	PriceRate               string `json:"priceRate"`               // 价格率
	BookTime                int64  `json:"bookTime"`                // 下单时间
	UpdateTime              int64  `json:"updateTime"`              // 更新时间
	WorkingType             string `json:"workingType"`             // 触发类型
	PriceProtect            bool   `json:"priceProtect"`            // 价格保护
	SelfTradePreventionMode string `json:"selfTradePreventionMode"` // 自成交保护模式
	GoodTillDate            int64  `json:"goodTillDate"`            // GTD 自动取消时间
	PriceMatch              string `json:"priceMatch"`              // 价格匹配模式
}

// ErrorResponse API错误响应
type ErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
