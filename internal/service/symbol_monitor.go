package service

import (
	"sync"
	"time"

	"new_listing_trade/internal/api/binance"
	"new_listing_trade/internal/logger"
	"new_listing_trade/internal/models"
)

// SymbolMonitor 币对监控服务
type SymbolMonitor struct {
	client         *binance.Client
	symbols        map[string]*models.Symbol           // 内存中存储的币对，key为symbol名称
	newListings    map[string]*models.NewListingSymbol // 新上币对列表，key为symbol名称
	mu             sync.RWMutex                        // 读写锁保护symbols和newListings
	lastUpdateTime time.Time                           // 最后更新时间
	onNewSymbols   func([]*models.Symbol)              // 发现新币对时的回调函数
	isInitialized  bool                                // 是否已完成初始化
}

// NewSymbolMonitor 创建新的币对监控服务
func NewSymbolMonitor() *SymbolMonitor {
	return &SymbolMonitor{
		client:      binance.NewClient(),
		symbols:     make(map[string]*models.Symbol),
		newListings: make(map[string]*models.NewListingSymbol),
	}
}

// SetOnNewSymbolsCallback 设置发现新币对时的回调函数
func (sm *SymbolMonitor) SetOnNewSymbolsCallback(callback func([]*models.Symbol)) {
	sm.onNewSymbols = callback
}

// Start 启动监控服务
func (sm *SymbolMonitor) Start() error {
	logger.Info("启动币对监控服务...")

	// 启动时立即拉取一次（初始化）
	if err := sm.fetchAndUpdate(true); err != nil {
		return err
	}

	// 启动定时任务，每2分钟拉取一次
	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for range ticker.C {
			if err := sm.fetchAndUpdate(false); err != nil {
				logger.Errorf("定时拉取币对数据失败: %v", err)
			}
		}
	}()

	logger.Info("币对监控服务启动成功")
	return nil
}

// fetchAndUpdate 拉取并更新币对数据
func (sm *SymbolMonitor) fetchAndUpdate(isInitial bool) error {
	if isInitial {
		logger.Info("开始初始化币对数据...")
	} else {
		logger.Info("开始拉取币对数据...")
	}

	exchangeInfo, err := sm.client.GetExchangeInfo()
	if err != nil {
		return err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 找出新币对
	var newSymbols []*models.Symbol
	var newListings []*models.NewListingSymbol
	foundTime := time.Now()
	currentTimeMillis := foundTime.UnixMilli() // 当前时间（毫秒）

	for _, symbol := range exchangeInfo.Symbols {
		// 只处理状态为TRADING的币对
		if symbol.Status != "TRADING" || symbol.Status != "PENDING_TRADING" {
			continue
		}

		// 检查是否为新币对
		isNewSymbol := false

		if isInitial {
			// 初始化时：只有onboardDate大于当前时间的币对才当成新币对
			if symbol.OnboardDate > currentTimeMillis {
				isNewSymbol = true
				logger.Infof("检测到即将上线的新币对: %s, 上线时间: %s (当前时间: %s)",
					symbol.Symbol,
					time.Unix(symbol.OnboardDate/1000, 0).Format("2006-01-02 15:04:05"),
					foundTime.Format("2006-01-02 15:04:05"))
			}
		} else {
			// 后续更新时：如果内存中不存在，则为新币对
			if _, exists := sm.symbols[symbol.Symbol]; !exists {
				isNewSymbol = true
			}
		}

		if isNewSymbol {
			newSymbols = append(newSymbols, &symbol)

			// 如果不在新币对列表中，则添加到新币对列表
			if _, exists := sm.newListings[symbol.Symbol]; !exists {
				newListing := &models.NewListingSymbol{
					Symbol:      symbol.Symbol,
					OnboardDate: symbol.OnboardDate,
					Status:      symbol.Status,
					FoundTime:   foundTime,
					IsOrdered:   false,
					OrderTime:   nil,
				}
				sm.newListings[symbol.Symbol] = newListing
				newListings = append(newListings, newListing)
			}
		}

		// 更新或添加币对信息
		sm.symbols[symbol.Symbol] = &symbol
	}

	sm.lastUpdateTime = time.Now()

	// 如果不是初始化，且有新币对，才打印新币对信息
	if !isInitial && len(newSymbols) > 0 {
		logger.Infof("发现新币对，数量: %d", len(newSymbols))
		for _, symbol := range newSymbols {
			logger.Infof("  新币对: %s, 上线时间: %s",
				symbol.Symbol,
				time.Unix(symbol.OnboardDate/1000, 0).Format("2006-01-02 15:04:05"))
		}
	}

	// 如果有新币对，触发回调
	if len(newSymbols) > 0 && sm.onNewSymbols != nil {
		sm.onNewSymbols(newSymbols)
	}

	if isInitial {
		logger.Infof("币对数据初始化完成，当前币对数量: %d, 新币对数量: %d", len(sm.symbols), len(newListings))
		sm.isInitialized = true
	} else {
		logger.Infof("币对数据更新完成，当前币对数量: %d, 新币对数量: %d", len(sm.symbols), len(newSymbols))
	}

	return nil
}

// GetSymbols 获取所有币对
func (sm *SymbolMonitor) GetSymbols() map[string]*models.Symbol {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// 返回副本，避免外部修改
	result := make(map[string]*models.Symbol)
	for k, v := range sm.symbols {
		result[k] = v
	}
	return result
}

// GetSymbol 获取指定币对信息
func (sm *SymbolMonitor) GetSymbol(symbol string) (*models.Symbol, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	s, exists := sm.symbols[symbol]
	return s, exists
}

// GetLastUpdateTime 获取最后更新时间
func (sm *SymbolMonitor) GetLastUpdateTime() time.Time {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.lastUpdateTime
}

// GetSymbolCount 获取币对数量
func (sm *SymbolMonitor) GetSymbolCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.symbols)
}

// GetNewListings 获取所有新上币对
func (sm *SymbolMonitor) GetNewListings() map[string]*models.NewListingSymbol {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// 返回副本，避免外部修改
	result := make(map[string]*models.NewListingSymbol)
	for k, v := range sm.newListings {
		// 创建副本
		listingCopy := *v
		result[k] = &listingCopy
	}
	return result
}

// GetNewListing 获取指定新币对信息
func (sm *SymbolMonitor) GetNewListing(symbol string) (*models.NewListingSymbol, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	listing, exists := sm.newListings[symbol]
	if !exists {
		return nil, false
	}

	// 返回副本
	listingCopy := *listing
	return &listingCopy, true
}

// MarkAsOrdered 标记新币对为已下单
func (sm *SymbolMonitor) MarkAsOrdered(symbol string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	listing, exists := sm.newListings[symbol]
	if !exists {
		return false
	}

	if listing.IsOrdered {
		return false // 已经下单过了
	}

	now := time.Now()
	listing.IsOrdered = true
	listing.OrderTime = &now
	return true
}

// GetUnorderedListings 获取未下单的新币对列表
func (sm *SymbolMonitor) GetUnorderedListings() []*models.NewListingSymbol {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var result []*models.NewListingSymbol
	for _, listing := range sm.newListings {
		if !listing.IsOrdered {
			listingCopy := *listing
			result = append(result, &listingCopy)
		}
	}
	return result
}

// GetNewListingCount 获取新币对数量
func (sm *SymbolMonitor) GetNewListingCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.newListings)
}

// AddNewListing 手动添加新币对（用于模拟）
func (sm *SymbolMonitor) AddNewListing(symbol string, onboardDate int64) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 检查是否已存在
	if _, exists := sm.newListings[symbol]; exists {
		return false // 已存在
	}

	// 添加到symbols map（如果不存在）
	if _, exists := sm.symbols[symbol]; !exists {
		sm.symbols[symbol] = &models.Symbol{
			Symbol:      symbol,
			OnboardDate: onboardDate,
			Status:      "TRADING",
		}
	}

	// 添加到新币对列表
	newListing := &models.NewListingSymbol{
		Symbol:      symbol,
		OnboardDate: onboardDate,
		Status:      "TRADING",
		FoundTime:   time.Now(),
		IsOrdered:   false,
		OrderTime:   nil,
	}
	sm.newListings[symbol] = newListing

	logger.Infof("手动添加新币对: %s", symbol)
	return true
}
