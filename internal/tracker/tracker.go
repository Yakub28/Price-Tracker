package tracker

import (
	"fmt"
	"log"
	"stock-tracker/internal/alerts"
	"stock-tracker/internal/api"
	"stock-tracker/internal/models"
	"sync"
	"time"
)

type StockTracker struct {
	stocks   map[string]*models.Stock
	mu       sync.RWMutex
	client   *api.AlphaVantageClient
	monitor  *alerts.AlertMonitor
	interval time.Duration
}

func New(apiKey string, interval time.Duration, alertThreshold float64) *StockTracker {
	return &StockTracker{
		stocks:   make(map[string]*models.Stock),
		client:   api.NewClient(apiKey),
		monitor:  alerts.NewMonitor(alertThreshold),
		interval: interval,
	}
}

func (st *StockTracker) AddStock(symbol string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if _, exists := st.stocks[symbol]; !exists {
		st.stocks[symbol] = models.NewStock(symbol)
		log.Printf("Added %s to tracking list", symbol)
	}
}

func (st *StockTracker) RemoveStock(symbol string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	delete(st.stocks, symbol)
	log.Printf("Removed %s from tracking list", symbol)
}

func (st *StockTracker) UpdateStock(symbol string) error {
	newData, err := st.client.GetQuote(symbol)
	if err != nil {
		return fmt.Errorf("failed to update %s: %w", symbol, err)
	}

	st.mu.Lock()
	if stock, exists := st.stocks[symbol]; exists {
		stock.UpdatePrice(newData.CurrentPrice, newData.ChangePercent)
		st.mu.Unlock()

		st.monitor.CheckStock(stock)
	} else {
		st.mu.Unlock()
	}

	return nil
}

func (st *StockTracker) UpdateAll() {
	st.mu.RLock()
	symbols := make([]string, 0, len(st.stocks))
	for symbol := range st.stocks {
		symbols = append(symbols, symbol)
	}
	st.mu.RUnlock()

	for _, symbol := range symbols {
		if err := st.UpdateStock(symbol); err != nil {
			log.Printf("Error updating %s: %v", symbol, err)
		}
		// Rate limiting: 5 calls per minute for free tier
		time.Sleep(12 * time.Second)
	}
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
			stock.Symbol,
			stock.CurrentPrice,
			changeSymbol,
			stock.ChangePercent,
			stock.LastUpdated.Format("15:04:05"),
		)
	}
	fmt.Println("===========================================\n")
}

func (st *StockTracker) Run() {
	st.monitor.Start()

	ticker := time.NewTicker(st.interval)
	defer ticker.Stop()

	// Initial update
	log.Println("Performing initial stock update...")
	st.UpdateAll()
	st.Display()

	for range ticker.C {
		log.Println("Updating stock prices...")
		st.UpdateAll()
		st.Display()
	}
}

func (st *StockTracker) Close() {
	st.monitor.Close()
}
