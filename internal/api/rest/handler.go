package rest

import (
	"encoding/json"
	"net/http"
	"stock-tracker/internal/metrics"
	"stock-tracker/internal/repository"
	"stock-tracker/pkg/logger"
	"strconv"
	"time"

	ws "stock-tracker/internal/api/websocket"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Handler struct {
	repo     repository.StockRepository
	wsHub    *ws.Hub
	metrics  *metrics.Metrics
	upgrader websocket.Upgrader
}

func NewHandler(repo repository.StockRepository, wsHub *ws.Hub, m *metrics.Metrics) *Handler {
	return &Handler{
		repo:    repo,
		wsHub:   wsHub,
		metrics: m,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, ErrorResponse{Error: http.StatusText(status), Message: message})
}

// GetAllStocks returns all tracked stocks
func (h *Handler) GetAllStocks(w http.ResponseWriter, r *http.Request) {
	stocks, err := h.repo.GetAllStocks(r.Context())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get all stocks")
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve stocks")
		return
	}

	h.respondJSON(w, http.StatusOK, stocks)
}

// GetStock returns a specific stock by symbol
func (h *Handler) GetStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	stock, err := h.repo.GetStock(r.Context(), symbol)
	if err != nil {
		logger.Error().Err(err).Str("symbol", symbol).Msg("Failed to get stock")
		h.respondError(w, http.StatusNotFound, "Stock not found")
		return
	}

	h.respondJSON(w, http.StatusOK, stock)
}

// GetPriceHistory returns historical prices for a stock
func (h *Handler) GetPriceHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	// Parse query parameters
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	from := time.Now().Add(-24 * time.Hour)
	if f := r.URL.Query().Get("from"); f != "" {
		if parsed, err := time.Parse(time.RFC3339, f); err == nil {
			from = parsed
		}
	}

	to := time.Now()
	if t := r.URL.Query().Get("to"); t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			to = parsed
		}
	}

	prices, err := h.repo.GetPriceHistory(r.Context(), symbol, from, to, limit)
	if err != nil {
		logger.Error().Err(err).Str("symbol", symbol).Msg("Failed to get price history")
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve price history")
		return
	}

	h.respondJSON(w, http.StatusOK, prices)
}

// GetAlerts returns alerts for a stock
func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	alerts, err := h.repo.GetAlerts(r.Context(), symbol, limit)
	if err != nil {
		logger.Error().Err(err).Str("symbol", symbol).Msg("Failed to get alerts")
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve alerts")
		return
	}

	h.respondJSON(w, http.StatusOK, alerts)
}

// GetRecentAlerts returns recent alerts across all stocks
func (h *Handler) GetRecentAlerts(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	alerts, err := h.repo.GetRecentAlerts(r.Context(), limit)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get recent alerts")
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve alerts")
		return
	}

	h.respondJSON(w, http.StatusOK, alerts)
}

// HandleWebSocket upgrades HTTP connection to WebSocket
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to upgrade to WebSocket")
		return
	}

	client := ws.NewClient(h.wsHub, conn)

	h.wsHub.RegisterClient(client)

	go client.WritePump()
	go client.ReadPump()
}

// HealthCheck endpoint
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}
