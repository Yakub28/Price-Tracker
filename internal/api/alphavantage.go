package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"stock-tracker/internal/models"
)

const baseURL = "https://www.alphavantage.co/query"

type AlphaVantageClient struct {
	apiKey     string
	httpClient *http.Client
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

func NewClient(apiKey string) *AlphaVantageClient {
	return &AlphaVantageClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *AlphaVantageClient) GetQuote(symbol string) (*models.Stock, error) {
	url := fmt.Sprintf("%s?function=GLOBAL_QUOTE&symbol=%s&apikey=%s",
		baseURL, symbol, c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var data globalQuoteResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if data.GlobalQuote.Symbol == "" {
		return nil, fmt.Errorf("invalid symbol or API limit reached")
	}

	return c.parseResponse(&data, symbol)
}

func (c *AlphaVantageClient) parseResponse(data *globalQuoteResponse, symbol string) (*models.Stock, error) {
	var price float64
	if _, err := fmt.Sscanf(data.GlobalQuote.Price, "%f", &price); err != nil {
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
