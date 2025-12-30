package repository

import (
	"context"
	"stock-tracker/internal/models"
	"time"
)

type StockRepository interface {
	// Stock operations
	CreateStock(ctx context.Context, stock *models.Stock) error
	GetStock(ctx context.Context, symbol string) (*models.Stock, error)
	GetAllStocks(ctx context.Context) ([]*models.Stock, error)
	UpdateStock(ctx context.Context, stock *models.Stock) error
	DeleteStock(ctx context.Context, symbol string) error

	// Price history operations
	SavePrice(ctx context.Context, price *models.StockPrice) error
	GetPriceHistory(ctx context.Context, symbol string, from, to time.Time, limit int) ([]*models.StockPrice, error)
	GetLatestPrice(ctx context.Context, symbol string) (*models.StockPrice, error)

	// Alert operations
	SaveAlert(ctx context.Context, alert *models.Alert) error
	GetAlerts(ctx context.Context, symbol string, limit int) ([]*models.Alert, error)
	GetRecentAlerts(ctx context.Context, limit int) ([]*models.Alert, error)
}
