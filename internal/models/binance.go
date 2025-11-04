package models

// ExchangeInfo 币安期货交易所信息
type ExchangeInfo struct {
	Symbols []Symbol `json:"symbols"`
}

// Symbol 交易对信息
type Symbol struct {
	Symbol      string   `json:"symbol"`
	OnboardDate int64    `json:"onboardDate"`
	Status      string   `json:"status"`
	Filters     []Filter `json:"filters,omitempty"` // 交易对过滤器
}

// Filter 交易对过滤器
type Filter struct {
	FilterType string `json:"filterType"`         // LOT_SIZE, PRICE_FILTER等
	MinQty     string `json:"minQty,omitempty"`   // 最小交易量
	MaxQty     string `json:"maxQty,omitempty"`   // 最大交易量
	StepSize   string `json:"stepSize,omitempty"` // 数量步进值
	MinPrice   string `json:"minPrice,omitempty"` // 最小价格
	MaxPrice   string `json:"maxPrice,omitempty"` // 最大价格
	TickSize   string `json:"tickSize,omitempty"` // 价格步进值
}
