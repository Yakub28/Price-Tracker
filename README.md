# Stock Price Tracker

A production-grade automated stock price tracker written in Go.

## Features

- Real-time stock price monitoring
- Configurable update intervals
- Price change alerts
- Thread-safe concurrent operations
- Clean separation of concerns
- Graceful shutdown handling

## Project Structure

```
stock-tracker/
├── cmd/
│   └── tracker/          # Application entry point
│       └── main.go
├── internal/             # Private application code
│   ├── models/          # Data models
│   │   └── stock.go
│   ├── api/             # External API clients
│   │   └── alphavantage.go
│   ├── tracker/         # Core tracking logic
│   │   └── tracker.go
│   └── alerts/          # Alert monitoring
│       └── monitor.go
├── pkg/                 # Public reusable packages
│   └── config/
│       └── config.go
├── go.mod
└── README.md
```

### Prometheus Metrics

**API Metrics:**
- `stock_tracker_api_calls_total` - Total API calls (by provider, symbol, status)
- `stock_tracker_api_call_duration_seconds` - API call latency histogram
- `stock_tracker_api_errors_total` - API errors (by type)

**Stock Update Metrics:**
- `stock_tracker_update_duration_seconds` - Update operation duration
- `stock_tracker_updates_total` - Total updates (by status)
- `stock_tracker_current_price` - Current stock prices (gauge)
- `stock_tracker_price_change_percent` - Price change percentage (gauge)

**Alert Metrics:**
- `stock_tracker_alerts_triggered_total` - Total alerts fired

**System Metrics:**
- `stock_tracker_tracked_stocks_count` - Number of tracked stocks
- `stock_tracker_update_cycles_total` - Completed update cycles

View metrics:

bashcurl http://localhost:9090/metrics
## Setup

1. Get a free API key from Alpha Vantage:
   https://www.alphavantage.co/support/#api-key

2. Set the environment variable:
   ```bash
   export ALPHA_VANTAGE_API_KEY="your_api_key_here"
   ```

3. Initialize the module:
   ```bash
   go mod init github.com/yourusername/stock-tracker
   go mod tidy
   ```

4. Run the application:
   ```bash
   go run cmd/tracker/main.go
   ```

## Configuration

Edit `pkg/config/config.go` to customize:
- Update interval (default: 5 minutes)
- Alert threshold (default: 5%)
- Default stock symbols

## Building

```bash
go build -o stock-tracker cmd/tracker/main.go
./stock-tracker
```

## Running Tests

```bash
go test ./...
```