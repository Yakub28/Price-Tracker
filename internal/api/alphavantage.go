package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"stock-tracker/internal/metrics"
	"stock-tracker/internal/models"
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

func NewClient(apiKey string, m *metrics.Metrics) *AlphaVantageClient {
	return &AlphaVantageClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		metrics: m,
	}
}

func (c *AlphaVantageClient) GetQuote(symbol string) (*models.Stock, error) {
	start := time.Now()

	logger.Debug().
		Str("symbol", symbol).
		Str("provider", providerName).
		Msg("Fetching quote from API")

	url := fmt.Sprintf("%s?function=GLOBAL_QUOTE&symbol=%s&apikey=%s",
		baseURL, symbol, c.apiKey)

	resp, err := c.httpClient.Get(url)
	duration := time.Since(start).Seconds()

	// Record API call duration
	c.metrics.APICallDuration.WithLabelValues(providerName, symbol).Observe(duration)

	if err != nil {
		c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "error").Inc()
		c.metrics.APICallErrors.WithLabelValues(providerName, symbol, "network_error").Inc()

		logger.Error().
			Err(err).
			Str("symbol", symbol).
			Float64("duration_seconds", duration).
			Msg("Failed to fetch data from API")

		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "error").Inc()
		c.metrics.APICallErrors.WithLabelValues(providerName, symbol, "http_error").Inc()

		logger.Error().
			Int("status_code", resp.StatusCode).
			Str("symbol", symbol).
			Msg("API returned non-200 status code")

		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "error").Inc()
		c.metrics.APICallErrors.WithLabelValues(providerName, symbol, "read_error").Inc()

		logger.Error().
			Err(err).
			Str("symbol", symbol).
			Msg("Failed to read response body")

		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var data globalQuoteResponse
	if err := json.Unmarshal(body, &data); err != nil {
		c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "error").Inc()
		c.metrics.APICallErrors.WithLabelValues(providerName, symbol, "parse_error").Inc()

		logger.Error().
			Err(err).
			Str("symbol", symbol).
			Str("response", string(body)).
			Msg("Failed to parse JSON response")

		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if data.GlobalQuote.Symbol == "" {
		c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "error").Inc()
		c.metrics.APICallErrors.WithLabelValues(providerName, symbol, "invalid_symbol").Inc()

		logger.Warn().
			Str("symbol", symbol).
			Msg("Invalid symbol or API limit reached")

		return nil, fmt.Errorf("invalid symbol or API limit reached")
	}

	stock, err := c.parseResponse(&data, symbol)
	if err != nil {
		c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "error").Inc()
		c.metrics.APICallErrors.WithLabelValues(providerName, symbol, "parse_error").Inc()
		return nil, err
	}

	// Success metrics
	c.metrics.APICallsTotal.WithLabelValues(providerName, symbol, "success").Inc()

	logger.Info().
		Str("symbol", symbol).
		Float64("price", stock.CurrentPrice).
		Float64("change_percent", stock.ChangePercent).
		Float64("duration_seconds", duration).
		Msg("Successfully fetched stock quote")

	return stock, nil
}

func (c *AlphaVantageClient) parseResponse(data *globalQuoteResponse, symbol string) (*models.Stock, error) {
	var price float64
	if _, err := fmt.Sscanf(data.GlobalQuote.Price, "%f", &price); err != nil {
		logger.Error().
			Err(err).
			Str("symbol", symbol).
			Str("price_string", data.GlobalQuote.Price).
			Msg("Failed to parse price")
		return nil, fmt.Errorf("failed to parse price: %w", err)
	}

	var changePercent float64
	changePercentStr := data.GlobalQuote.ChangePercent
	if len(changePercentStr) > 0 && changePercentStr[len(changePercentStr)-1] == '%' {
		fmt.Sscanf(changePercentStr[:len(changePercentStr)-1], "%f", &changePercent)
	}

	stock := models.NewStock(symbol)
	stock.UpdatePrice(price, changePercent)

	return stock, nil
}
