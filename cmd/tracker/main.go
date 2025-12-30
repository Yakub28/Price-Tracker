package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"stock-tracker/internal/api/websocket"
	"stock-tracker/internal/metrics"
	"stock-tracker/internal/repository"
	"stock-tracker/internal/tracker"
	"stock-tracker/pkg/config"
	"stock-tracker/pkg/logger"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.Debug)
	logger.Info().Bool("debug", cfg.Debug).Str("version", "2.0.0").Msg("Starting Stock Price Tracker")

	// Initialize database
	repo, err := repository.NewPostgresRepository(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer repo.Close()

	// Initialize metrics
	m := metrics.New()

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Start Prometheus metrics server
	metricsAddr := fmt.Sprintf(":%d", cfg.MetricsPort)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Info().Str("address", metricsAddr).Msg("Starting Prometheus metrics server")
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start metrics server")
		}
	}()

	logger.Info().
		Dur("update_interval", cfg.UpdateInterval).
		Float64("alert_threshold", cfg.AlertThreshold).
		Int("metrics_port", cfg.MetricsPort).
		Str("database_url", maskDatabaseURL(cfg.DatabaseURL)).
		Msg("Configuration loaded")

	stockTracker := tracker.New(cfg.APIKey, cfg.UpdateInterval, cfg.AlertThreshold, m, repo, wsHub)
	defer stockTracker.Close()

	for _, symbol := range cfg.DefaultSymbols {
		stockTracker.AddStock(symbol)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
		stockTracker.Close()
		repo.Close()
		logger.Info().Msg("Graceful shutdown complete")
		os.Exit(0)
	}()

	logger.Info().Msg("Stock tracker running. Press Ctrl+C to stop")
	logger.Info().Str("metrics_url", fmt.Sprintf("http://localhost%s/metrics", metricsAddr)).Msg("Prometheus metrics available")

	stockTracker.Run()
}

func maskDatabaseURL(url string) string {
	// Simple masking for logging
	if len(url) > 20 {
		return url[:20] + "..."
	}
	return url
}
