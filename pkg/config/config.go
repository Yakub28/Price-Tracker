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
	MetricsPort    int
	Debug          bool
	DefaultSymbols []string
}

func Load() (*Config, error) {
	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ALPHA_VANTAGE_API_KEY environment variable not set")
	}
	debug := os.Getenv("DEBUG") == "true"

	return &Config{
		APIKey:         apiKey,
		UpdateInterval: 5 * time.Minute,
		AlertThreshold: 5.0,
		MetricsPort:    9090,
		DefaultSymbols: []string{"AAPL", "GOOGL", "MSFT", "TSLA"},
		Debug:          debug,
	}, nil
}
