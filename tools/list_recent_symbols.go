package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"new_listing_trade/internal/models"
)

const (
	BinanceFuturesBaseURL = "https://fapi.binance.com"
	FAPIExchangeInfoEndpoint = "/fapi/v1/exchangeInfo"
)

// SymbolInfo 币对信息（用于排序）
type SymbolInfo struct {
	Symbol      string
	OnboardDate int64
	Status      string
}

func main() {
	// 获取交易所信息
	url := fmt.Sprintf("%s%s", BinanceFuturesBaseURL, FAPIExchangeInfoEndpoint)
	
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("获取数据失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API返回错误: %d - %s\n", resp.StatusCode, string(body))
		return
	}

	var exchangeInfo models.ExchangeInfo
	if err := json.Unmarshal(body, &exchangeInfo); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		return
	}

	// 收集所有TRADING状态的币对
	var symbols []SymbolInfo
	for _, symbol := range exchangeInfo.Symbols {
		if symbol.Status == "TRADING" && symbol.OnboardDate > 0 {
			symbols = append(symbols, SymbolInfo{
				Symbol:      symbol.Symbol,
				OnboardDate: symbol.OnboardDate,
				Status:      symbol.Status,
			})
		}
	}

	// 按上线时间倒序排序（最新的在前）
	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].OnboardDate > symbols[j].OnboardDate
	})

	// 取前100个
	count := 100
	if len(symbols) < count {
		count = len(symbols)
	}

	// 输出币对列表
	for i := 0; i < count; i++ {
		fmt.Println(symbols[i].Symbol)
	}
}

