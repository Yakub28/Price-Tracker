package main

import (
	"log"
	"os"
	"os/signal"
	"stock-tracker/internal/tracker"
	"stock-tracker/pkg/config"
	"syscall"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("Starting Stock Price Tracker...")
	log.Printf("Update interval: %v", cfg.UpdateInterval)
	log.Printf("Alert threshold: %.2f%%", cfg.AlertThreshold)

	stockTracker := tracker.New(cfg.APIKey, cfg.UpdateInterval, cfg.AlertThreshold)
	defer stockTracker.Close()

	for _, symbol := range cfg.DefaultSymbols {
		stockTracker.AddStock(symbol)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nShutting down gracefully...")
		stockTracker.Close()
		os.Exit(0)
	}()

	log.Println("Press Ctrl+C to stop")
	stockTracker.Run()
}
