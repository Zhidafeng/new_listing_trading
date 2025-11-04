package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// LoadConfigOrCreateDefault 加载配置文件，如果不存在则创建默认配置
func LoadConfigOrCreateDefault(path string) (*Config, error) {
	config, err := LoadConfig(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 创建默认配置
			defaultConfig := GetDefaultConfig()
			if err := SaveConfig(path, defaultConfig); err != nil {
				return nil, fmt.Errorf("创建默认配置文件失败: %w", err)
			}
			return defaultConfig, nil
		}
		return nil, err
	}
	return config, nil
}

// SaveConfig 保存配置文件
func SaveConfig(path string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *Config {
	return &Config{
		Binance: BinanceConfig{
			APIKey:    "",
			SecretKey: "",
			APIType:   "fapi", // 默认使用fapi
			BaseURL:   "",
		},
		Trading: TradingConfig{
			DefaultNotional: "10", // 默认10 USDT
			PositionSide:    "BOTH",
			StopLoss: StopLossConfig{
				Enabled:     true,
				Percent:     2.0, // 2%止损
				WorkingType: "MARK_PRICE",
			},
			TakeProfit: TakeProfitConfig{
				Enabled:     true,
				Percent:     5.0, // 5%止盈
				WorkingType: "MARK_PRICE",
			},
		},
		Log: LogConfig{
			Level:    "info",         // 默认info级别
			File:     "logs/app.log", // 默认日志文件路径
			MaxSize:  100,            // 100MB
			MaxAge:   7,              // 保留7天
			Compress: true,           // 压缩旧日志
		},
	}
}
