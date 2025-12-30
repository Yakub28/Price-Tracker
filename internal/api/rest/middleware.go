package rest

import (
	"net/http"
	"stock-tracker/pkg/logger"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Msg("HTTP request received")

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Dur("duration", duration).
			Msg("HTTP request completed")
	})
}

func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
