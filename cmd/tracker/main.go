package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"stock-tracker/internal/metrics"
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

	// Initialize structured logging
	logger.Init(cfg.Debug)

	logger.Info().
		Bool("debug", cfg.Debug).
		Str("version", "1.0.0").
		Msg("Starting Stock Price Tracker")

	// Initialize metrics
	m := metrics.New()

	// Start Prometheus metrics server
	metricsAddr := fmt.Sprintf(":%d", cfg.MetricsPort)
	go func() {
		http.Handle("/metrics", promhttp.Handler())

		logger.Info().
			Str("address", metricsAddr).
			Msg("Starting Prometheus metrics server")

		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			logger.Fatal().
				Err(err).
				Msg("Failed to start metrics server")
		}
	}()

	logger.Info().
		Dur("update_interval", cfg.UpdateInterval).
		Float64("alert_threshold", cfg.AlertThreshold).
		Int("metrics_port", cfg.MetricsPort).
		Msg("Configuration loaded")

	stockTracker := tracker.New(cfg.APIKey, cfg.UpdateInterval, cfg.AlertThreshold, m)
	defer stockTracker.Close()

	for _, symbol := range cfg.DefaultSymbols {
		stockTracker.AddStock(symbol)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info().
			Str("signal", sig.String()).
			Msg("Received shutdown signal")

		stockTracker.Close()
		logger.Info().Msg("Graceful shutdown complete")
		os.Exit(0)
	}()

	logger.Info().Msg("Stock tracker running. Press Ctrl+C to stop")
	logger.Info().
		Str("metrics_url", fmt.Sprintf("http://localhost%s/metrics", metricsAddr)).
		Msg("Prometheus metrics available")

	stockTracker.Run()
}
