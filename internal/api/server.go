package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"new_listing_trade/internal/logger"
	"new_listing_trade/internal/models"
	"new_listing_trade/internal/service"
)

// Server HTTP服务器
type Server struct {
	symbolMonitor  *service.SymbolMonitor
	tradingService *service.TradingService
	port           string
	engine         *gin.Engine
}

// NewServer 创建新的HTTP服务器
func NewServer(port string, symbolMonitor *service.SymbolMonitor, tradingService *service.TradingService) *Server {
	// 设置为release模式（生产环境）或debug模式（开发环境）
	gin.SetMode(gin.ReleaseMode)

	engine := gin.Default()

	// 添加CORS中间件（如果需要）
	engine.Use(corsMiddleware())

	server := &Server{
		symbolMonitor:  symbolMonitor,
		tradingService: tradingService,
		port:           port,
		engine:         engine,
	}

	// 注册路由
	server.registerRoutes()

	return server
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// registerRoutes 注册路由
func (s *Server) registerRoutes() {
	api := s.engine.Group("/api")
	{
		api.POST("/simulate/new-listing", s.handleSimulateNewListing)
		api.GET("/status", s.handleStatus)
		api.GET("/new-listings", s.handleGetNewListings)
		api.GET("/symbols", s.handleGetSymbols)
	}

	// 健康检查
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	logger.Infof("HTTP服务器启动在端口 %s", s.port)
	logger.Info("API接口:")
	logger.Info("  POST /api/simulate/new-listing - 模拟新币上线并触发交易")
	logger.Info("  GET  /api/status - 获取服务状态")
	logger.Info("  GET  /api/new-listings - 获取新币对列表")
	logger.Info("  GET  /api/symbols - 获取所有币对")
	logger.Info("  GET  /health - 健康检查")

	return s.engine.Run(":" + s.port)
}

// SimulateNewListingRequest 模拟新币上线请求
type SimulateNewListingRequest struct {
	Symbol       string `json:"symbol" binding:"required"` // 币对名称，例如 "BTCUSDT"
	NotionalUSDT string `json:"notional_usdt,omitempty"`   // USDT金额，留空使用配置默认值
}

// SimulateNewListingResponse 模拟新币上线响应
type SimulateNewListingResponse struct {
	Success  bool              `json:"success"`
	Message  string            `json:"message"`
	Symbol   string            `json:"symbol,omitempty"`
	OrderSet *OrderSetResponse `json:"order_set,omitempty"`
}

// OrderSetResponse 订单集合响应
type OrderSetResponse struct {
	Symbol          string                `json:"symbol"`
	SellOrder       *models.OrderResponse `json:"sell_order,omitempty"` // 做空卖单
	StopLossOrder   *models.OrderResponse `json:"stop_loss_order,omitempty"`
	TakeProfitOrder *models.OrderResponse `json:"take_profit_order,omitempty"`
	StopLossError   string                `json:"stop_loss_error,omitempty"`
	TakeProfitError string                `json:"take_profit_error,omitempty"`
}

// handleSimulateNewListing 处理模拟新币上线请求
func (s *Server) handleSimulateNewListing(c *gin.Context) {
	var req SimulateNewListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查交易服务是否可用
	if s.tradingService == nil {
		c.JSON(http.StatusServiceUnavailable, SimulateNewListingResponse{
			Success: false,
			Message: "交易服务未初始化，请检查配置文件中的API密钥设置",
		})
		return
	}

	// 模拟新币上线：添加到监控服务的新币对列表
	onboardDate := time.Now().UnixMilli()
	added := s.symbolMonitor.AddNewListing(req.Symbol, onboardDate)
	if !added {
		logger.Infof("币对 %s 已存在，继续执行交易流程", req.Symbol)
	}

	// 检查是否已经下单过
	if listing, exists := s.symbolMonitor.GetNewListing(req.Symbol); exists && listing.IsOrdered {
		c.JSON(http.StatusBadRequest, SimulateNewListingResponse{
			Success: false,
			Message: "币对 " + req.Symbol + " 已经下单过了",
			Symbol:  req.Symbol,
		})
		return
	}

	// 执行交易流程
	logger.Infof("模拟新币上线: %s, 开始执行交易流程...", req.Symbol)
	orderSet, err := s.tradingService.CreateOrdersWithStopLossAndTakeProfit(req.Symbol, req.NotionalUSDT)

	if err != nil {
		logger.Errorf("交易流程执行失败: %v", err)
		c.JSON(http.StatusInternalServerError, SimulateNewListingResponse{
			Success: false,
			Message: "交易流程执行失败: " + err.Error(),
			Symbol:  req.Symbol,
		})
		return
	}

	// 标记为已下单
	if s.symbolMonitor != nil {
		s.symbolMonitor.MarkAsOrdered(req.Symbol)
	}

	// 构建响应
	orderSetResp := &OrderSetResponse{
		Symbol: orderSet.Symbol,
	}

	if orderSet.SellOrder != nil {
		orderSetResp.SellOrder = orderSet.SellOrder
	}
	if orderSet.StopLossOrder != nil {
		orderSetResp.StopLossOrder = orderSet.StopLossOrder
	}
	if orderSet.TakeProfitOrder != nil {
		orderSetResp.TakeProfitOrder = orderSet.TakeProfitOrder
	}
	if orderSet.StopLossError != nil {
		orderSetResp.StopLossError = orderSet.StopLossError.Error()
	}
	if orderSet.TakeProfitError != nil {
		orderSetResp.TakeProfitError = orderSet.TakeProfitError.Error()
	}

	c.JSON(http.StatusOK, SimulateNewListingResponse{
		Success:  true,
		Message:  "新币上线模拟成功，已创建订单",
		Symbol:   req.Symbol,
		OrderSet: orderSetResp,
	})
}

// handleStatus 处理状态查询请求
func (s *Server) handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"symbol_count":      s.symbolMonitor.GetSymbolCount(),
		"new_listing_count": s.symbolMonitor.GetNewListingCount(),
		"last_update_time":  s.symbolMonitor.GetLastUpdateTime().Format("2006-01-02 15:04:05"),
		"trading_enabled":   s.tradingService != nil,
	})
}

// handleGetNewListings 获取新币对列表
func (s *Server) handleGetNewListings(c *gin.Context) {
	listings := s.symbolMonitor.GetNewListings()
	c.JSON(http.StatusOK, listings)
}

// handleGetSymbols 获取所有币对
func (s *Server) handleGetSymbols(c *gin.Context) {
	symbols := s.symbolMonitor.GetSymbols()
	c.JSON(http.StatusOK, symbols)
}
