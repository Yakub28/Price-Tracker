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