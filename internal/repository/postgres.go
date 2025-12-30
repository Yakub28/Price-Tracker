package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"stock-tracker/internal/models"
	"stock-tracker/pkg/logger"
	"time"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(databaseURL string) (*PostgresRepository, error) {
	ctx := context.Background()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().Msg("Successfully connected to PostgreSQL database")

	return &PostgresRepository{pool: pool}, nil
}

func (r *PostgresRepository) Close() {
	r.pool.Close()
}

func (r *PostgresRepository) CreateStock(ctx context.Context, stock *models.Stock) error {
	query := `
		INSERT INTO stocks (symbol, name, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (symbol) DO UPDATE SET updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, stock.Symbol, stock.Name).
		Scan(&stock.ID, &stock.CreatedAt, &stock.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create stock: %w", err)
	}

	logger.Debug().
		Str("symbol", stock.Symbol).
		Int("id", stock.ID).
		Msg("Stock created in database")

	return nil
}

func (r *PostgresRepository) GetStock(ctx context.Context, symbol string) (*models.Stock, error) {
	query := `
		SELECT id, symbol, name, created_at, updated_at
		FROM stocks
		WHERE symbol = $1
	`

	stock := &models.Stock{}
	err := r.pool.QueryRow(ctx, query, symbol).Scan(
		&stock.ID, &stock.Symbol, &stock.Name,
		&stock.CreatedAt, &stock.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	return stock, nil
}

func (r *PostgresRepository) GetAllStocks(ctx context.Context) ([]*models.Stock, error) {
	query := `
		SELECT s.id, s.symbol, s.name, s.created_at, s.updated_at,
		       sp.price, sp.change_percent, sp.timestamp
		FROM stocks s
		LEFT JOIN LATERAL (
			SELECT price, change_percent, timestamp
			FROM stock_prices
			WHERE stock_id = s.id
			ORDER BY timestamp DESC
			LIMIT 1
		) sp ON true
		ORDER BY s.symbol
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all stocks: %w", err)
	}
	defer rows.Close()

	var stocks []*models.Stock
	for rows.Next() {
		stock := &models.Stock{}
		var price, changePercent *float64
		var timestamp *time.Time

		err := rows.Scan(
			&stock.ID, &stock.Symbol, &stock.Name,
			&stock.CreatedAt, &stock.UpdatedAt,
			&price, &changePercent, &timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock: %w", err)
		}

		if price != nil {
			stock.CurrentPrice = *price
		}
		if changePercent != nil {
			stock.ChangePercent = *changePercent
		}
		if timestamp != nil {
			stock.LastUpdated = *timestamp
		}

		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func (r *PostgresRepository) UpdateStock(ctx context.Context, stock *models.Stock) error {
	query := `
		UPDATE stocks
		SET name = $1, updated_at = NOW()
		WHERE symbol = $2
		RETURNING updated_at
	`

	err := r.pool.QueryRow(ctx, query, stock.Name, stock.Symbol).
		Scan(&stock.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	return nil
}

func (r *PostgresRepository) DeleteStock(ctx context.Context, symbol string) error {
	query := `DELETE FROM stocks WHERE symbol = $1`

	_, err := r.pool.Exec(ctx, query, symbol)
	if err != nil {
		return fmt.Errorf("failed to delete stock: %w", err)
	}

	logger.Info().Str("symbol", symbol).Msg("Stock deleted from database")
	return nil
}

func (r *PostgresRepository) SavePrice(ctx context.Context, price *models.StockPrice) error {
	query := `
		INSERT INTO stock_prices (stock_id, price, change_percent, volume, timestamp)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		price.StockID, price.Price, price.ChangePercent,
		price.Volume, price.Timestamp,
	).Scan(&price.ID)

	if err != nil {
		return fmt.Errorf("failed to save price: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetPriceHistory(ctx context.Context, symbol string, from, to time.Time, limit int) ([]*models.StockPrice, error) {
	query := `
		SELECT sp.id, sp.stock_id, s.symbol, sp.price, sp.change_percent, sp.volume, sp.timestamp
		FROM stock_prices sp
		JOIN stocks s ON s.id = sp.stock_id
		WHERE s.symbol = $1 AND sp.timestamp BETWEEN $2 AND $3
		ORDER BY sp.timestamp DESC
		LIMIT $4
	`

	rows, err := r.pool.Query(ctx, query, symbol, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}
	defer rows.Close()

	var prices []*models.StockPrice
	for rows.Next() {
		price := &models.StockPrice{}
		err := rows.Scan(
			&price.ID, &price.StockID, &price.Symbol,
			&price.Price, &price.ChangePercent, &price.Volume,
			&price.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan price: %w", err)
		}
		prices = append(prices, price)
	}

	return prices, nil
}

func (r *PostgresRepository) GetLatestPrice(ctx context.Context, symbol string) (*models.StockPrice, error) {
	query := `
		SELECT sp.id, sp.stock_id, s.symbol, sp.price, sp.change_percent, sp.volume, sp.timestamp
		FROM stock_prices sp
		JOIN stocks s ON s.id = sp.stock_id
		WHERE s.symbol = $1
		ORDER BY sp.timestamp DESC
		LIMIT 1
	`

	price := &models.StockPrice{}
	err := r.pool.QueryRow(ctx, query, symbol).Scan(
		&price.ID, &price.StockID, &price.Symbol,
		&price.Price, &price.ChangePercent, &price.Volume,
		&price.Timestamp,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get latest price: %w", err)
	}

	return price, nil
}

func (r *PostgresRepository) SaveAlert(ctx context.Context, alert *models.Alert) error {
	query := `
		INSERT INTO alerts (stock_id, alert_type, threshold, message, triggered_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		alert.StockID, alert.AlertType, alert.Threshold,
		alert.Message, alert.TriggeredAt,
	).Scan(&alert.ID)

	if err != nil {
		return fmt.Errorf("failed to save alert: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetAlerts(ctx context.Context, symbol string, limit int) ([]*models.Alert, error) {
	query := `
		SELECT a.id, a.stock_id, s.symbol, a.alert_type, a.threshold, a.message, a.triggered_at
		FROM alerts a
		JOIN stocks s ON s.id = a.stock_id
		WHERE s.symbol = $1
		ORDER BY a.triggered_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		alert := &models.Alert{}
		err := rows.Scan(
			&alert.ID, &alert.StockID, &alert.Symbol,
			&alert.AlertType, &alert.Threshold, &alert.Message,
			&alert.TriggeredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (r *PostgresRepository) GetRecentAlerts(ctx context.Context, limit int) ([]*models.Alert, error) {
	query := `
		SELECT a.id, a.stock_id, s.symbol, a.alert_type, a.threshold, a.message, a.triggered_at
		FROM alerts a
		JOIN stocks s ON s.id = a.stock_id
		ORDER BY a.triggered_at DESC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		alert := &models.Alert{}
		err := rows.Scan(
			&alert.ID, &alert.StockID, &alert.Symbol,
			&alert.AlertType, &alert.Threshold, &alert.Message,
			&alert.TriggeredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}
