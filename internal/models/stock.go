package models

import "time"

type Stock struct {
	ID            int       `json:"id"`
	Symbol        string    `json:"symbol"`
	Name          string    `json:"name,omitempty"`
	CurrentPrice  float64   `json:"current_price"`
	PreviousPrice float64   `json:"previous_price"`
	ChangePercent float64   `json:"change_percent"`
	LastUpdated   time.Time `json:"last_updated"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type StockPrice struct {
	ID            int64     `json:"id"`
	StockID       int       `json:"stock_id"`
	Symbol        string    `json:"symbol"`
	Price         float64   `json:"price"`
	ChangePercent float64   `json:"change_percent"`
	Volume        int64     `json:"volume,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

type Alert struct {
	ID          int       `json:"id"`
	StockID     int       `json:"stock_id"`
	Symbol      string    `json:"symbol"`
	AlertType   string    `json:"alert_type"`
	Threshold   float64   `json:"threshold"`
	Message     string    `json:"message"`
	TriggeredAt time.Time `json:"triggered_at"`
}

func NewStock(symbol string) *Stock {
	return &Stock{
		Symbol: symbol,
	}
}

func (s *Stock) UpdatePrice(newPrice, changePercent float64) {
	s.PreviousPrice = s.CurrentPrice
	s.CurrentPrice = newPrice
	s.ChangePercent = changePercent
	s.LastUpdated = time.Now()
}

func (s *Stock) CalculatePriceChange() float64 {
	if s.PreviousPrice == 0 {
		return 0
	}
	return ((s.CurrentPrice - s.PreviousPrice) / s.PreviousPrice) * 100
}
