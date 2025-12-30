package tracker

import (
	"context"
	"fmt"
	"stock-tracker/internal/alerts"
	"stock-tracker/internal/api"
	"stock-tracker/internal/api/websocket"
	"stock-tracker/internal/metrics"
	"stock-tracker/internal/models"
	"stock-tracker/internal/repository"
	"stock-tracker/pkg/logger"
	"sync"
	"time"
)

type StockTracker struct {
	stocks   map[string]*models.Stock
	mu       sync.RWMutex
	client   *api.AlphaVantageClient
	monitor  *alerts.AlertMonitor
	metrics  *metrics.Metrics
	repo     repository.StockRepository
	wsHub    *websocket.Hub
	interval time.Duration
}

func New(apiKey string, interval time.Duration, alertThreshold float64, m *metrics.Metrics, repo repository.StockRepository, wsHub *websocket.Hub) *StockTracker {
	client := api.NewClient(apiKey, m, repo)

	return &StockTracker{
		stocks:   make(map[string]*models.Stock),
		client:   client,
		monitor:  alerts.NewMonitor(alertThreshold, m, repo, wsHub),
		metrics:  m,
		repo:     repo,
		wsHub:    wsHub,
		interval: interval,
	}
}

func (st *StockTracker) AddStock(symbol string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if _, exists := st.stocks[symbol]; !exists {
		stock := models.NewStock(symbol)
		st.stocks[symbol] = stock
		st.metrics.TrackedStocksCount.Inc()

		// Create stock in database
		ctx := context.Background()
		if err := st.repo.CreateStock(ctx, stock); err != nil {
			logger.Error().Err(err).Str("symbol", symbol).Msg("Failed to create stock in database")
		}

		logger.Info().Str("symbol", symbol).Msg("Added stock to tracking list")
	}
}

func (st *StockTracker) RemoveStock(symbol string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if _, exists := st.stocks[symbol]; exists {
		delete(st.stocks, symbol)
		st.metrics.TrackedStocksCount.Dec()
		logger.Info().Str("symbol", symbol).Msg("Removed stock from tracking list")
	}
}

func (st *StockTracker) UpdateStock(symbol string) error {
	start := time.Now()
	ctx := context.Background()

	logger.Debug().Str("symbol", symbol).Msg("Starting stock update")

	newData, err := st.client.GetQuote(ctx, symbol)
	duration := time.Since(start).Seconds()

	st.metrics.StockUpdateDuration.WithLabelValues(symbol).Observe(duration)

	if err != nil {
		st.metrics.StockUpdatesTotal.WithLabelValues(symbol, "error").Inc()
		logger.Error().Err(err).Str("symbol", symbol).Float64("duration_seconds", duration).Msg("Failed to update stock")
		return fmt.Errorf("failed to update %s: %w", symbol, err)
	}

	st.mu.Lock()
	if stock, exists := st.stocks[symbol]; exists {
		stock.UpdatePrice(newData.CurrentPrice, newData.ChangePercent)
		stock.ID = newData.ID

		st.metrics.CurrentStockPrice.WithLabelValues(symbol).Set(stock.CurrentPrice)
		st.metrics.StockPriceChange.WithLabelValues(symbol).Set(stock.ChangePercent)

		st.mu.Unlock()

		// Broadcast update via WebSocket
		st.wsHub.BroadcastStockUpdate(stock)

		st.monitor.CheckStock(stock)
		st.metrics.StockUpdatesTotal.WithLabelValues(symbol, "success").Inc()

		logger.Info().Str("symbol", symbol).Float64("price", stock.CurrentPrice).Float64("change_percent", stock.ChangePercent).Float64("duration_seconds", duration).Msg("Successfully updated stock")
	} else {
		st.mu.Unlock()
	}

	return nil
}

func (st *StockTracker) UpdateAll() {
	logger.Info().Msg("Starting update cycle for all stocks")

	st.mu.RLock()
	symbols := make([]string, 0, len(st.stocks))
	for symbol := range st.stocks {
		symbols = append(symbols, symbol)
	}
	st.mu.RUnlock()

	successCount := 0
	for _, symbol := range symbols {
		if err := st.UpdateStock(symbol); err != nil {
			logger.Error().Err(err).Str("symbol", symbol).Msg("Error updating stock in batch")
		} else {
			successCount++
		}
		time.Sleep(12 * time.Second)
	}

	st.metrics.UpdateCyclesTotal.Inc()
	logger.Info().Int("total", len(symbols)).Int("success", successCount).Int("failed", len(symbols)-successCount).Msg("Completed update cycle")
}

func (st *StockTracker) Display() {
	st.mu.RLock()
	defer st.mu.RUnlock()

	fmt.Println("\n" + "===========================================")
	fmt.Printf("Stock Tracker Update - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("===========================================")

	for _, stock := range st.stocks {
		changeSymbol := "→"
		if stock.ChangePercent > 0 {
			changeSymbol = "↑"
		} else if stock.ChangePercent < 0 {
			changeSymbol = "↓"
		}

		fmt.Printf("%-6s: $%-8.2f %s %.2f%% (Last: %s)\n",
			stock.Symbol, stock.CurrentPrice, changeSymbol,
			stock.ChangePercent, stock.LastUpdated.Format("15:04:05"),
		)
	}
	fmt.Println("===========================================\n")
}

func (st *StockTracker) Run() {
	st.monitor.Start()
	ticker := time.NewTicker(st.interval)
	defer ticker.Stop()

	logger.Info().Msg("Performing initial stock update")
	st.UpdateAll()
	st.Display()

	for range ticker.C {
		st.UpdateAll()
		st.Display()
	}
}

func (st *StockTracker) Close() {
	logger.Info().Msg("Closing stock tracker")
	st.monitor.Close()
}
