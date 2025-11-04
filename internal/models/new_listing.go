package models

import "time"

// NewListingSymbol 新上币对信息（带下单状态）
type NewListingSymbol struct {
	Symbol      string     // 币对名称
	OnboardDate int64      // 上线时间（毫秒时间戳）
	Status      string     // 交易状态
	FoundTime   time.Time  // 发现时间
	IsOrdered   bool       // 是否已下单
	OrderTime   *time.Time // 下单时间（如果已下单）
}
