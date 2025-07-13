package middlewares

import (
	"net/http"
	"strings"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from environment variable
		originsStr := viper.GetString(consts.CORS_ORIGINS)
		origins := strings.Split(originsStr, ";")

		// Trim whitespace from origins
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}

		// Get the origin from the request
		requestOrigin := r.Header.Get("Origin")

		// Debug logging
		logger.Debug("CORS check",
			zap.String("requestOrigin", requestOrigin),
			zap.Strings("allowedOrigins", origins),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		// Check if the request origin is allowed
		allowedOrigin := ""
		for _, origin := range origins {
			if origin == requestOrigin || origin == "*" {
				allowedOrigin = origin
				break
			}
		}

		// Always set CORS headers regardless of origin for better compatibility
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "300") // Cache preflight response for 5 minutes

		// Set CORS headers - always set the origin header for valid origins
		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			logger.Debug("CORS allowed", zap.String("origin", allowedOrigin))
		} else if requestOrigin != "" {
			// Origin not allowed - log for debugging but still allow for development
			logger.Warn("CORS rejected - origin not in allowed list",
				zap.String("origin", requestOrigin),
				zap.Strings("allowedOrigins", origins),
			)
			// For development, you might want to allow all origins temporarily
			// Uncomment the next line if you want to allow all origins for debugging
			// w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// No origin header (e.g., same-origin requests, Postman, curl)
			logger.Debug("No origin header in request")
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
