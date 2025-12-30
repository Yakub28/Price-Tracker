package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func SetupRoutes(handler *Handler) http.Handler {
	r := mux.NewRouter()

	// API v1 routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(JSONContentType)
	api.Use(LoggingMiddleware)

	// Stock endpoints
	api.HandleFunc("/stocks", handler.GetAllStocks).Methods("GET")
	api.HandleFunc("/stocks/{symbol}", handler.GetStock).Methods("GET")
	api.HandleFunc("/stocks/{symbol}/history", handler.GetPriceHistory).Methods("GET")
	api.HandleFunc("/stocks/{symbol}/alerts", handler.GetAlerts).Methods("GET")

	// Alert endpoints
	api.HandleFunc("/alerts", handler.GetRecentAlerts).Methods("GET")

	// Health check
	api.HandleFunc("/health", handler.HealthCheck).Methods("GET")

	// WebSocket endpoint
	r.HandleFunc("/ws", handler.HandleWebSocket)

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	return c.Handler(r)
}
