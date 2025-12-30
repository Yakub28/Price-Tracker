package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"stock-tracker/internal/api/rest"
	"stock-tracker/internal/api/websocket"
	"stock-tracker/internal/metrics"
	"stock-tracker/internal/repository"
	"stock-tracker/pkg/config"
	"stock-tracker/pkg/logger"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.Debug)
	logger.Info().Bool("debug", cfg.Debug).Str("version", "2.0.0").Msg("Starting Stock Tracker API Server")

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
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsAddr := fmt.Sprintf(":%d", cfg.MetricsPort)
		logger.Info().Str("address", metricsAddr).Msg("Starting Prometheus metrics server")
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start metrics server")
		}
	}()

	// Setup REST API routes
	handler := rest.NewHandler(repo, wsHub, m)
	router := rest.SetupRoutes(handler)

	// Create HTTP server
	apiAddr := fmt.Sprintf(":%d", cfg.APIPort)
	server := &http.Server{
		Addr:         apiAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info().Str("address", apiAddr).Msg("Starting API server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start API server")
		}
	}()

	logger.Info().Msgf("API server running on http://localhost%s", apiAddr)
	logger.Info().Msg("Available endpoints:")
	logger.Info().Msg("  GET  /api/v1/stocks")
	logger.Info().Msg("  GET  /api/v1/stocks/{symbol}")
	logger.Info().Msg("  GET  /api/v1/stocks/{symbol}/history")
	logger.Info().Msg("  GET  /api/v1/stocks/{symbol}/alerts")
	logger.Info().Msg("  GET  /api/v1/alerts")
	logger.Info().Msg("  GET  /api/v1/health")
	logger.Info().Msg("  WS   /ws")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	logger.Info().Str("signal", sig.String()).Msg("Received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server shutdown failed")
	}

	repo.Close()
	logger.Info().Msg("Graceful shutdown complete")
}
