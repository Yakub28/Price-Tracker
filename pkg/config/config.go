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
	MetricsPort    int
	APIPort        int
	DatabaseURL    string
	Debug          bool
}

func Load() (*Config, error) {
	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ALPHA_VANTAGE_API_KEY environment variable not set")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://stockuser:stockpass@localhost:5432/stocktracker?sslmode=disable"
	}

	debug := os.Getenv("DEBUG") == "true"

	return &Config{
		APIKey:         apiKey,
		UpdateInterval: 5 * time.Minute,
		AlertThreshold: 5.0,
		DefaultSymbols: []string{"AAPL", "GOOGL", "MSFT", "TSLA"},
		MetricsPort:    9091,
		APIPort:        8080,
		DatabaseURL:    databaseURL,
		Debug:          debug,
	}, nil
}
