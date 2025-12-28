package alerts

import (
	"fmt"
	"log"
	"stock-tracker/internal/models"
)

type AlertMonitor struct {
	threshold float64
	alertChan chan string
}

func NewMonitor(threshold float64) *AlertMonitor {
	return &AlertMonitor{
		threshold: threshold,
		alertChan: make(chan string, 100),
	}
}

func (m *AlertMonitor) Start() {
	go func() {
		for alert := range m.alertChan {
			log.Printf("🚨 ALERT: %s\n", alert)
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

		select {
		case m.alertChan <- alert:
		default:
			log.Println("Alert channel full, dropping alert")
		}
	}
}

func (m *AlertMonitor) Close() {
	close(m.alertChan)
}
