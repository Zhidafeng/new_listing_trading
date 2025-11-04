package binance

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"new_listing_trade/internal/models"
)

// GetSymbolInfo 从exchangeInfo中获取指定交易对的信息
func GetSymbolInfo(exchangeInfo *models.ExchangeInfo, symbol string) (*models.Symbol, error) {
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == symbol {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("交易对 %s 不存在", symbol)
}

// GetStepSize 获取交易对的stepSize（数量步进值）
func GetStepSize(symbolInfo *models.Symbol) (float64, error) {
	for _, filter := range symbolInfo.Filters {
		if filter.FilterType == "LOT_SIZE" && filter.StepSize != "" {
			stepSize, err := strconv.ParseFloat(filter.StepSize, 64)
			if err != nil {
				return 0, fmt.Errorf("解析stepSize失败: %w", err)
			}
			return stepSize, nil
		}
	}
	// 如果没有找到LOT_SIZE过滤器，返回默认值0.00000001（8位小数）
	return 0.00000001, nil
}

// GetMinQty 获取交易对的最小交易量
func GetMinQty(symbolInfo *models.Symbol) (float64, error) {
	for _, filter := range symbolInfo.Filters {
		if filter.FilterType == "LOT_SIZE" && filter.MinQty != "" {
			minQty, err := strconv.ParseFloat(filter.MinQty, 64)
			if err != nil {
				return 0, fmt.Errorf("解析minQty失败: %w", err)
			}
			return minQty, nil
		}
	}
	return 0, nil
}

// AdjustQuantity 根据stepSize调整quantity精度
func AdjustQuantity(quantity float64, stepSize float64) float64 {
	if stepSize <= 0 {
		return quantity
	}
	// 向下取整到stepSize的倍数
	adjusted := math.Floor(quantity/stepSize) * stepSize
	return adjusted
}

// FormatQuantity 格式化quantity字符串，根据stepSize确定精度
func FormatQuantity(quantity float64, stepSize float64, stepSizeStr string) string {
	if stepSize <= 0 {
		return strconv.FormatFloat(quantity, 'f', 8, 64)
	}

	// 调整quantity到stepSize的倍数
	adjusted := AdjustQuantity(quantity, stepSize)

	// 计算需要的小数位数
	precision := getPrecision(stepSizeStr)

	// 格式化quantity
	return strconv.FormatFloat(adjusted, 'f', precision, 64)
}

// getPrecision 从stepSize字符串中获取精度（小数位数）
func getPrecision(stepSizeStr string) int {
	// 处理科学计数法格式（如 "1e-8"）
	if strings.Contains(strings.ToLower(stepSizeStr), "e") {
		parts := strings.Split(strings.ToLower(stepSizeStr), "e")
		if len(parts) == 2 {
			expStr := strings.TrimSpace(parts[1])
			if strings.HasPrefix(expStr, "-") {
				exp, err := strconv.Atoi(expStr[1:])
				if err == nil {
					return exp
				}
			}
		}
	}

	// 处理普通小数格式
	if !strings.Contains(stepSizeStr, ".") {
		return 0
	}

	parts := strings.Split(stepSizeStr, ".")
	if len(parts) != 2 {
		return 8
	}

	decimalPart := parts[1]
	// 移除尾部的0
	decimalPart = strings.TrimRight(decimalPart, "0")

	if len(decimalPart) == 0 {
		return 0
	}

	return len(decimalPart)
}

// ValidateAndAdjustQuantity 验证并调整quantity，确保符合交易对规则
func ValidateAndAdjustQuantity(quantity float64, symbolInfo *models.Symbol) (string, error) {
	minQty, err := GetMinQty(symbolInfo)
	if err != nil {
		return "", err
	}

	stepSize, err := GetStepSize(symbolInfo)
	if err != nil {
		return "", err
	}

	// 获取stepSize字符串（用于精度计算）
	stepSizeStr := ""
	for _, filter := range symbolInfo.Filters {
		if filter.FilterType == "LOT_SIZE" && filter.StepSize != "" {
			stepSizeStr = filter.StepSize
			break
		}
	}
	if stepSizeStr == "" {
		stepSizeStr = "0.00000001" // 默认值
	}

	// 确保quantity >= minQty
	if quantity < minQty {
		return "", fmt.Errorf("交易数量 %.8f 小于最小交易量 %.8f", quantity, minQty)
	}

	// 调整quantity到stepSize的倍数
	adjusted := AdjustQuantity(quantity, stepSize)

	// 再次检查是否满足minQty
	if adjusted < minQty {
		// 向上调整到下一个stepSize倍数
		adjusted = math.Ceil(minQty/stepSize) * stepSize
	}

	// 格式化quantity
	quantityStr := FormatQuantity(adjusted, stepSize, stepSizeStr)

	return quantityStr, nil
}

// GetTickSize 获取交易对的tickSize（价格步进值）
func GetTickSize(symbolInfo *models.Symbol) (float64, error) {
	for _, filter := range symbolInfo.Filters {
		if filter.FilterType == "PRICE_FILTER" && filter.TickSize != "" {
			tickSize, err := strconv.ParseFloat(filter.TickSize, 64)
			if err != nil {
				return 0, fmt.Errorf("解析tickSize失败: %w", err)
			}
			return tickSize, nil
		}
	}
	// 如果没有找到PRICE_FILTER过滤器，返回默认值0.00000001（8位小数）
	return 0.00000001, nil
}

// GetMinPrice 获取交易对的最小价格
func GetMinPrice(symbolInfo *models.Symbol) (float64, error) {
	for _, filter := range symbolInfo.Filters {
		if filter.FilterType == "PRICE_FILTER" && filter.MinPrice != "" {
			minPrice, err := strconv.ParseFloat(filter.MinPrice, 64)
			if err != nil {
				return 0, fmt.Errorf("解析minPrice失败: %w", err)
			}
			return minPrice, nil
		}
	}
	return 0, nil
}

// AdjustPrice 根据tickSize调整价格精度
func AdjustPrice(price float64, tickSize float64) float64 {
	if tickSize <= 0 {
		return price
	}
	// 向下取整到tickSize的倍数
	adjusted := math.Floor(price/tickSize) * tickSize
	return adjusted
}

// FormatPrice 格式化价格字符串，根据tickSize确定精度
func FormatPrice(price float64, tickSize float64, tickSizeStr string) string {
	if tickSize <= 0 {
		return strconv.FormatFloat(price, 'f', 8, 64)
	}

	// 调整价格到tickSize的倍数
	adjusted := AdjustPrice(price, tickSize)

	// 计算需要的小数位数
	precision := getPrecision(tickSizeStr)

	// 格式化价格
	return strconv.FormatFloat(adjusted, 'f', precision, 64)
}

// ValidateAndAdjustPrice 验证并调整价格，确保符合交易对规则
func ValidateAndAdjustPrice(price float64, symbolInfo *models.Symbol) (string, error) {
	minPrice, err := GetMinPrice(symbolInfo)
	if err != nil {
		return "", err
	}

	tickSize, err := GetTickSize(symbolInfo)
	if err != nil {
		return "", err
	}

	// 获取tickSize字符串（用于精度计算）
	tickSizeStr := ""
	for _, filter := range symbolInfo.Filters {
		if filter.FilterType == "PRICE_FILTER" && filter.TickSize != "" {
			tickSizeStr = filter.TickSize
			break
		}
	}
	if tickSizeStr == "" {
		tickSizeStr = "0.00000001" // 默认值
	}

	// 确保price >= minPrice
	if minPrice > 0 && price < minPrice {
		return "", fmt.Errorf("价格 %.8f 小于最小价格 %.8f", price, minPrice)
	}

	// 调整价格到tickSize的倍数
	adjusted := AdjustPrice(price, tickSize)

	// 再次检查是否满足minPrice
	if minPrice > 0 && adjusted < minPrice {
		// 向上调整到下一个tickSize倍数
		adjusted = math.Ceil(minPrice/tickSize) * tickSize
	}

	// 格式化价格
	priceStr := FormatPrice(adjusted, tickSize, tickSizeStr)

	return priceStr, nil
}
