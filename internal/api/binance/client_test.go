package binance

import (
	"testing"
)

func TestGetExchangeInfo(t *testing.T) {
	client := NewClient()

	exchangeInfo, err := client.GetExchangeInfo()
	if err != nil {
		t.Fatalf("获取交易所信息失败: %v", err)
	}

	if exchangeInfo == nil {
		t.Fatal("交易所信息为空")
	}

	if len(exchangeInfo.Symbols) == 0 {
		t.Fatal("交易对列表为空")
	}

	t.Logf("成功获取交易所信息:")
	t.Logf("交易对数量: %d", len(exchangeInfo.Symbols))
}
