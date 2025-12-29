package alerts

import (
	"fmt"
	"stock-tracker/internal/metrics"
	"stock-tracker/internal/models"
	"stock-tracker/pkg/logger"
)

type AlertMonitor struct {
	threshold float64
	alertChan chan string
	metrics   *metrics.Metrics
}

func NewMonitor(threshold float64, m *metrics.Metrics) *AlertMonitor {
	return &AlertMonitor{
		threshold: threshold,
		alertChan: make(chan string, 100),
		metrics:   m,
	}
}

func (m *AlertMonitor) Start() {
	go func() {
		for alert := range m.alertChan {
			logger.Warn().
				Str("alert", alert).
				Msg("Stock price alert triggered")
		}
	}()
}

func (m *AlertMonitor) CheckStock(stock *models.Stock) {
	if stock.PreviousPrice == 0 {
		return
	}

	pctChange := stock.CalculatePriceChange()
	if pctChange > m.threshold || pctChange < -m.threshold {
		alert := fmt.Sprintf("%s changed by %.2f%% (from $%.2f to $%.2f)",
			stock.Symbol, pctChange, stock.PreviousPrice, stock.CurrentPrice)

		m.metrics.AlertsTriggered.WithLabelValues(stock.Symbol, alert).Inc()
		select {
		case m.alertChan <- alert:
		default:
			logger.Warn().Msg("Alert channel full, dropping alert")
		}
	}
}

func (m *AlertMonitor) Close() {
	close(m.alertChan)
}
