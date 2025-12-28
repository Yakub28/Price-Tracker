package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	APIKey         string
	UpdateInterval time.Duration
	AlertThreshold float64
	DefaultSymbols []string
}

func Load() (*Config, error) {
	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ALPHA_VANTAGE_API_KEY environment variable not set")
	}

	return &Config{
		APIKey:         apiKey,
		UpdateInterval: 5 * time.Minute,
		AlertThreshold: 5.0,
		DefaultSymbols: []string{"AAPL", "GOOGL", "MSFT", "TSLA"},
	}, nil
}
