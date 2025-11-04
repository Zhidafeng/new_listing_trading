package binance

import (
	"time"

	"new_listing_trade/internal/logger"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries   int           // 最大重试次数
	InitialDelay time.Duration // 初始延迟时间
	MaxDelay     time.Duration // 最大延迟时间
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     2 * time.Second,
	}
}

// RetryWithBackoff 带退避的重试机制
func RetryWithBackoff(fn func() error, config RetryConfig) error {
	var lastErr error
	delay := config.InitialDelay

	for i := 0; i < config.MaxRetries; i++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// 检查是否是可重试的错误
		apiErr, ok := err.(*APIError)
		if !ok {
			// 不是APIError，可能是网络错误，可以重试
			if i < config.MaxRetries-1 {
				logger.Warnf("请求失败，%v后重试: %v", delay, err)
				time.Sleep(delay)
				delay *= 2
				if delay > config.MaxDelay {
					delay = config.MaxDelay
				}
				continue
			}
			break
		}

		// 如果是未知执行状态错误，不能重试
		if apiErr.IsUnknownStatus {
			logger.Warnf("收到未知执行状态错误，需要先查询订单状态: %v", err)
			return err
		}

		// 如果是可重试的错误
		if apiErr.IsRetryable {
			if i < config.MaxRetries-1 {
				logger.Warnf("API错误（可重试），%v后重试: %v", delay, err)
				time.Sleep(delay)
				delay *= 2
				if delay > config.MaxDelay {
					delay = config.MaxDelay
				}
				continue
			}
		} else {
			// 不可重试的错误，直接返回
			return err
		}
	}

	return lastErr
}
