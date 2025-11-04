package models

// TickerPrice 当前价格信息
type TickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}
