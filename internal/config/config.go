package config

// Config 应用配置
type Config struct {
	Binance BinanceConfig `yaml:"binance"`
	Trading TradingConfig `yaml:"trading"`
	Log     LogConfig     `yaml:"log"`
}

// BinanceConfig 币安API配置
type BinanceConfig struct {
	APIKey    string `yaml:"api_key"`
	SecretKey string `yaml:"secret_key"`
	APIType   string `yaml:"api_type,omitempty"` // API类型: fapi (U本位合约) 或 papi (统一账户)，默认fapi
	BaseURL   string `yaml:"base_url,omitempty"` // 可选，默认使用生产环境
}

// TradingConfig 交易配置
type TradingConfig struct {
	// 止盈止损配置
	StopLoss   StopLossConfig   `yaml:"stop_loss"`
	TakeProfit TakeProfitConfig `yaml:"take_profit"`
	// 订单配置
	DefaultNotional string `yaml:"default_notional"` // 默认下单USDT金额（例如："10"表示10 USDT）
	PositionSide    string `yaml:"position_side"`    // 持仓方向 BOTH/LONG/SHORT
}

// StopLossConfig 止损配置
type StopLossConfig struct {
	Enabled     bool    `yaml:"enabled"`      // 是否启用止损
	Percent     float64 `yaml:"percent"`      // 止损百分比（例如：2.0 表示2%）
	WorkingType string  `yaml:"working_type"` // 触发类型 MARK_PRICE/CONTRACT_PRICE
}

// TakeProfitConfig 止盈配置
type TakeProfitConfig struct {
	Enabled     bool    `yaml:"enabled"`      // 是否启用止盈
	Percent     float64 `yaml:"percent"`      // 止盈百分比（例如：5.0 表示5%）
	WorkingType string  `yaml:"working_type"` // 触发类型 MARK_PRICE/CONTRACT_PRICE
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `yaml:"level"`    // 日志级别: trace, debug, info, warn, error, fatal, panic
	File     string `yaml:"file"`     // 日志文件路径，留空则只输出到控制台
	MaxSize  int    `yaml:"max_size"` // 单个日志文件最大大小(MB)，0表示不限制
	MaxAge   int    `yaml:"max_age"`  // 日志文件保留天数，0表示不删除
	Compress bool   `yaml:"compress"` // 是否压缩旧日志文件
}
