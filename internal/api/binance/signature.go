package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
	"time"
)

// SignRequest 对请求参数进行HMAC SHA256签名
func SignRequest(params url.Values, secretKey string) string {
	// 将参数排序并拼接成字符串
	queryString := params.Encode()

	// 使用HMAC SHA256签名
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(queryString))
	signature := hex.EncodeToString(mac.Sum(nil))

	return signature
}

// SignQueryString 对查询字符串进行HMAC SHA256签名
func SignQueryString(queryString, secretKey string) string {
	// 使用HMAC SHA256签名
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(queryString))
	signature := hex.EncodeToString(mac.Sum(nil))

	return signature
}

// BuildQueryString 构建查询字符串（已排序）
func BuildQueryString(params map[string]string) string {
	// 提取所有键并排序
	keys := make([]string, 0, len(params))
	for k := range params {
		if params[k] != "" && params[k] != "0" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 构建查询字符串
	values := make([]string, 0, len(keys))
	for _, k := range keys {
		values = append(values, k+"="+url.QueryEscape(params[k]))
	}

	return strings.Join(values, "&")
}

// GetTimestamp 获取当前时间戳（毫秒）
func GetTimestamp() int64 {
	return time.Now().UnixMilli()
}
