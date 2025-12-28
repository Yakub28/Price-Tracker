package models

import "time"

type Stock struct {
	Symbol        string
	CurrentPrice  float64
	PreviousPrice float64
	ChangePercent float64
	LastUpdated   time.Time
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
