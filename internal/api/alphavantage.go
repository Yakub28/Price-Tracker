package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"stock-tracker/internal/metrics"
	"stock-tracker/internal/models"
	"stock-tracker/internal/repository"
	"stock-tracker/pkg/logger"
	"time"
)

const (
	baseURL      = "https://www.alphavantage.co/query"
	providerName = "alphavantage"
)

type AlphaVantageClient struct {
	apiKey     string
	httpClient *http.Client
	metrics    *metrics.Metrics
	repo       repository.StockRepository
}

type globalQuoteResponse struct {
	GlobalQuote struct {
		Symbol           string `json:"01. symbol"`
		Price            string `json:"05. price"`
		Change           string `json:"09. change"`
		ChangePercent    string `json:"10. change percent"`
		LatestTradingDay string `json:"07. latest trading day"`
	} `json:"Global Quote"`
}

func NewClient(apiKey string, m *metrics.Metrics, repo repository.StockRepository) *AlphaVantageClient {
	return &AlphaVantageClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		metrics: m,
		repo:    repo,
	}
}

func (c *AlphaVantageClient) GetQuote(ctx context.Context, symbol string) (*models.Stock, error) {
	start := time.Now()

	logger.Debug().Str("symbol", symbol).Str("provider", providerName).Msg("Fetching quote from API")

	url := fmt.Sprintf("%s?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", baseURL, symbol, c.apiKey)

	resp, err := c.httpClient.Get(url)
	duration := time.Since(start).Seconds()

	c.metrics.APICallDuration.WithLabelValues(providerName, symbol).Observe(duration)

	if err != nil {
		c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "error").Inc()
		c.metrics.APICallErrors.WithLabelValues(providerName, symbol, "network_error").Inc()
		logger.Error().Err(err).Str("symbol", symbol).Float64("duration_seconds", duration).Msg("Failed to fetch data from API")
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "error").Inc()
		c.metrics.APICallErrors.WithLabelValues(providerName, symbol, "http_error").Inc()
		logger.Error().Int("status_code", resp.StatusCode).Str("symbol", symbol).Msg("API returned non-200 status code")
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "error").Inc()
		c.metrics.APICallErrors.WithLabelValues(providerName, symbol, "read_error").Inc()
		logger.Error().Err(err).Str("symbol", symbol).Msg("Failed to read response body")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var data globalQuoteResponse
	if err := json.Unmarshal(body, &data); err != nil {
		c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "error").Inc()
		c.metrics.APICallErrors.WithLabelValues(providerName, symbol, "parse_error").Inc()
		logger.Error().Err(err).Str("symbol", symbol).Msg("Failed to parse JSON response")
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if data.GlobalQuote.Symbol == "" {
		c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "error").Inc()
		c.metrics.APICallErrors.WithLabelValues(providerName, symbol, "invalid_symbol").Inc()
		logger.Warn().Str("symbol", symbol).Msg("Invalid symbol or API limit reached")
		return nil, fmt.Errorf("invalid symbol or API limit reached")
	}

	stock, err := c.parseResponse(ctx, &data, symbol)
	if err != nil {
		c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "error").Inc()
		c.metrics.APICallErrors.WithLabelValues(providerName, symbol, "parse_error").Inc()
		return nil, err
	}

	c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "success").Inc()
	logger.Info().Str("symbol", symbol).Float64("price", stock.CurrentPrice).Float64("change_percent", stock.ChangePercent).Float64("duration_seconds", duration).Msg("Successfully fetched stock quote")

	return stock, nil
}

func (c *AlphaVantageClient) parseResponse(ctx context.Context, data *globalQuoteResponse, symbol string) (*models.Stock, error) {
	var price float64
	if _, err := fmt.Sscanf(data.GlobalQuote.Price, "%f", &price); err != nil {
		logger.Error().Err(err).Str("symbol", symbol).Str("price_string", data.GlobalQuote.Price).Msg("Failed to parse price")
		return nil, fmt.Errorf("failed to parse price: %w", err)
	}

	var changePercent float64
	changePercentStr := data.GlobalQuote.ChangePercent
	if len(changePercentStr) > 0 && changePercentStr[len(changePercentStr)-1] == '%' {
		fmt.Sscanf(changePercentStr[:len(changePercentStr)-1], "%f", &changePercent)
	}

	stock := models.NewStock(symbol)
	stock.UpdatePrice(price, changePercent)

	// Get or create stock in database
	dbStock, err := c.repo.GetStock(ctx, symbol)
	if err != nil {
		// Stock doesn't exist, create it
		if err := c.repo.CreateStock(ctx, stock); err != nil {
			logger.Error().Err(err).Str("symbol", symbol).Msg("Failed to create stock in database")
		} else {
			dbStock = stock
		}
	}

	if dbStock != nil {
		stock.ID = dbStock.ID

		// Save price history
		priceRecord := &models.StockPrice{
			StockID:       stock.ID,
			Symbol:        symbol,
			Price:         price,
			ChangePercent: changePercent,
			Timestamp:     time.Now(),
		}

		if err := c.repo.SavePrice(ctx, priceRecord); err != nil {
			logger.Error().Err(err).Str("symbol", symbol).Msg("Failed to save price to database")
		}
	}

	return stock, nil
}
