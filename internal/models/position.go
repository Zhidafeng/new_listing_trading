package models

// PositionRisk 持仓风险信息（币安API返回格式）
type PositionRisk struct {
	Symbol           string `json:"symbol"`           // 交易对
	PositionAmt      string `json:"positionAmt"`      // 持仓数量（正数为多，负数为空）
	EntryPrice       string `json:"entryPrice"`       // 开仓价格
	MarkPrice        string `json:"markPrice"`        // 标记价格
	UnRealizedProfit string `json:"unRealizedProfit"` // 未实现盈亏
	LiquidationPrice string `json:"liquidationPrice"` // 强平价格
	Leverage         string `json:"leverage"`         // 杠杆倍数
	MarginType       string `json:"marginType"`       // 保证金类型
	IsolatedMargin   string `json:"isolatedMargin"`   // 逐仓保证金
	PositionSide     string `json:"positionSide"`     // 持仓方向：BOTH/LONG/SHORT
	Notional         string `json:"notional"`         // 持仓名义价值
	IsolatedWallet   string `json:"isolatedWallet"`   // 逐仓钱包余额
	UpdateTime       int64  `json:"updateTime"`       // 更新时间
}

// Position 持仓信息（内部使用）
type Position struct {
	Symbol           string  `json:"symbol"`            // 交易对
	PositionAmt      float64 `json:"position_amt"`      // 持仓数量（正数为多，负数为空）
	EntryPrice       float64 `json:"entry_price"`       // 开仓价格
	MarkPrice        float64 `json:"mark_price"`        // 标记价格
	UnRealizedProfit float64 `json:"unrealized_profit"` // 未实现盈亏（USDT）
	LiquidationPrice float64 `json:"liquidation_price"` // 强平价格
	Leverage         float64 `json:"leverage"`          // 杠杆倍数
	MarginType       string  `json:"margin_type"`       // 保证金类型
	PositionSide     string  `json:"position_side"`     // 持仓方向：BOTH/LONG/SHORT
	Notional         float64 `json:"notional"`          // 持仓名义价值（USDT）
	UpdateTime       int64   `json:"update_time"`       // 更新时间
	ProfitPercent    float64 `json:"profit_percent"`    // 收益率百分比
}

// NegativePositionResponse 负收益仓位响应
type NegativePositionResponse struct {
	TotalCount int        `json:"total_count"` // 负收益仓位总数
	Positions  []Position `json:"positions"`   // 负收益仓位列表（按亏损从大到小排序）
}
