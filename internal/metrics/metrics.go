package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	// API call metrics
	APICallsTotal   *prometheus.CounterVec
	APICallDuration *prometheus.HistogramVec
	APICallErrors   *prometheus.CounterVec

	// Stock update metrics
	StockUpdateDuration *prometheus.HistogramVec
	StockUpdatesTotal   *prometheus.CounterVec

	// Current stock prices (gauges)
	CurrentStockPrice *prometheus.GaugeVec
	StockPriceChange  *prometheus.GaugeVec

	// Alert metrics
	AlertsTriggered *prometheus.CounterVec

	// System metrics
	TrackedStocksCount prometheus.Gauge
	UpdateCyclesTotal  prometheus.Counter
}

func New() *Metrics {
	return &Metrics{
		APICallsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "stock_tracker_api_calls_total",
				Help: "Total number of API calls made",
			},
			[]string{"provider", "symbol", "status"},
		),

		APICallDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "stock_tracker_api_call_duration_seconds",
				Help:    "Duration of API calls in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"provider", "symbol"},
		),

		APICallErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "stock_tracker_api_errors_total",
				Help: "Total number of API call errors",
			},
			[]string{"provider", "symbol", "error_type"},
		),

		StockUpdateDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "stock_tracker_update_duration_seconds",
				Help:    "Duration of stock update operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"symbol"},
		),

		StockUpdatesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "stock_tracker_updates_total",
				Help: "Total number of stock updates",
			},
			[]string{"symbol", "status"},
		),

		CurrentStockPrice: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "stock_tracker_current_price",
				Help: "Current stock price in USD",
			},
			[]string{"symbol"},
		),

		StockPriceChange: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "stock_tracker_price_change_percent",
				Help: "Stock price change percentage",
			},
			[]string{"symbol"},
		),

		AlertsTriggered: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "stock_tracker_alerts_triggered_total",
				Help: "Total number of alerts triggered",
			},
			[]string{"symbol", "alert_type"},
		),

		TrackedStocksCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "stock_tracker_tracked_stocks_count",
				Help: "Number of stocks being tracked",
			},
		),

		UpdateCyclesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "stock_tracker_update_cycles_total",
				Help: "Total number of update cycles completed",
			},
		),
	}
}
