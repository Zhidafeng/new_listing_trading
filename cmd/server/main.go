package main

import (
	"flag"

	"new_listing_trade/internal/api"
	"new_listing_trade/internal/config"
	"new_listing_trade/internal/logger"
	"new_listing_trade/internal/service"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	port := flag.String("port", "8081", "HTTP服务端口")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfigOrCreateDefault(*configPath)
	if err != nil {
		logger.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	if err := logger.Init(cfg.Log.Level, cfg.Log.File, cfg.Log.MaxSize, cfg.Log.MaxAge, cfg.Log.Compress); err != nil {
		logger.Fatalf("初始化日志失败: %v", err)
	}

	logger.Info("配置加载成功")
	logger.Infof("日志级别: %s, 日志文件: %s", cfg.Log.Level, cfg.Log.File)

	// 创建币对监控服务
	monitor := service.NewSymbolMonitor()

	// 启动监控服务（会立即拉取一次数据）
	if err := monitor.Start(); err != nil {
		logger.Fatalf("启动币对监控服务失败: %v", err)
	}

	logger.Info("币对监控服务启动成功")
	logger.Infof("当前币对数量: %d", monitor.GetSymbolCount())
	logger.Infof("最后更新时间: %s", monitor.GetLastUpdateTime().Format("2006-01-02 15:04:05"))

	// 创建交易服务（如果配置了API密钥）
	var tradingService *service.TradingService
	if cfg.Binance.APIKey != "" && cfg.Binance.SecretKey != "" {
		tradingService, err = service.NewTradingService(cfg)
		if err != nil {
			logger.Warnf("警告: 交易服务初始化失败: %v", err)
			logger.Warn("交易功能将不可用，但监控功能仍可正常使用")
		} else {
			logger.Info("交易服务启动成功")
		}
	} else {
		logger.Warn("警告: 未配置API密钥，交易功能不可用")
		logger.Warn("如需使用交易功能，请在配置文件中设置 binance.api_key 和 binance.secret_key")
	}

	// 创建HTTP服务器
	httpServer := api.NewServer(*port, monitor, tradingService)

	// 启动HTTP服务器（阻塞运行）
	logger.Info("服务运行中...")
	logger.Info("币对监控服务：每2分钟检查一次新币对")
	if tradingService != nil {
		logger.Info("交易服务：已启用")
	}

	if err := httpServer.Start(); err != nil {
		logger.Fatalf("HTTP服务器启动失败: %v", err)
	}
}
