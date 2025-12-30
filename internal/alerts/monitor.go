package alerts

import (
	"context"
	"fmt"
	"stock-tracker/internal/api/websocket"
	"stock-tracker/internal/metrics"
	"stock-tracker/internal/models"
	"stock-tracker/internal/repository"
	"stock-tracker/pkg/logger"
	"time"
)

type AlertMonitor struct {
	threshold float64
	alertChan chan string
	metrics   *metrics.Metrics
	repo      repository.StockRepository
	wsHub     *websocket.Hub
}

func NewMonitor(threshold float64, m *metrics.Metrics, repo repository.StockRepository, wsHub *websocket.Hub) *AlertMonitor {
	return &AlertMonitor{
		threshold: threshold,
		alertChan: make(chan string, 100),
		metrics:   m,
		repo:      repo,
		wsHub:     wsHub,
	}
}

func (m *AlertMonitor) Start() {
	go func() {
		for alert := range m.alertChan {
			logger.Warn().Str("alert", alert).Msg("Price alert triggered")
		}
	}()
}

func (m *AlertMonitor) CheckStock(stock *models.Stock) {
	if stock.PreviousPrice == 0 {
		return
	}

	pctChange := stock.CalculatePriceChange()
	if pctChange > m.threshold || pctChange < -m.threshold {
		alertType := "price_increase"
		if pctChange < 0 {
			alertType = "price_decrease"
		}

		m.metrics.AlertsTriggered.WithLabelValues(stock.Symbol, alertType).Inc()

		message := fmt.Sprintf("%s changed by %.2f%% (from $%.2f to $%.2f)",
			stock.Symbol, pctChange, stock.PreviousPrice, stock.CurrentPrice)

		// Save alert to database
		ctx := context.Background()
		alert := &models.Alert{
			StockID:     stock.ID,
			Symbol:      stock.Symbol,
			AlertType:   alertType,
			Threshold:   m.threshold,
			Message:     message,
			TriggeredAt: time.Now(),
		}

		if err := m.repo.SaveAlert(ctx, alert); err != nil {
			logger.Error().Err(err).Str("symbol", stock.Symbol).Msg("Failed to save alert to database")
		}

		// Broadcast alert via WebSocket
		m.wsHub.BroadcastAlert(alert)

		select {
		case m.alertChan <- message:
		default:
			logger.Warn().Msg("Alert channel full, dropping alert")
		}
	}
}

func (m *AlertMonitor) Close() {
	close(m.alertChan)
}
