package routes

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rogerwesterbo/svennescamping-backend/internal/handlers/adminhandler"
	"github.com/rogerwesterbo/svennescamping-backend/internal/handlers/healthhandler"
	"github.com/rogerwesterbo/svennescamping-backend/internal/handlers/transactionshandler"
	"github.com/rogerwesterbo/svennescamping-backend/internal/handlers/userhandler"
	"github.com/rogerwesterbo/svennescamping-backend/internal/middlewares"
	"github.com/rogerwesterbo/svennescamping-backend/internal/services"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
	"go.uber.org/zap"
)

// corsPreflightHandler handles CORS preflight OPTIONS requests
// CORS headers are set by the CORSMiddleware, this handler just returns 200 OK
func corsPreflightHandler(w http.ResponseWriter, r *http.Request) {
	// No additional headers needed - CORSMiddleware handles all CORS headers
	w.WriteHeader(http.StatusOK)
}

// addCORSPreflightHandlers adds OPTIONS method handlers to support CORS preflight requests
// This is necessary because browsers send OPTIONS requests before actual requests with certain headers
func addCORSPreflightHandlers(router *mux.Router) {
	router.Methods("OPTIONS").HandlerFunc(corsPreflightHandler)
}

// SetupRoutes configures all the routes for the application
func SetupRoutes(router *mux.Router, logger *zap.Logger) {
	router.Use(middlewares.CORSMiddleware)
	router.Use(middlewares.LoggingMiddleware)
	router.Use(middlewares.ContentTypeMiddleware)

	// Handle OPTIONS requests globally for CORS preflight
	addCORSPreflightHandlers(router)

	// Root endpoint - simple response for basic connectivity check
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"svennescamping-backend"}`))
	}).Methods("GET")

	// ACME challenge endpoint for Let's Encrypt cert-manager
	// This allows cert-manager to place challenge files that can be served
	router.PathPrefix("/.well-known/acme-challenge/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the URL path
		token := r.URL.Path[len("/.well-known/acme-challenge/"):]
		if token == "" {
			http.NotFound(w, r)
			return
		}

		// For now, return 404 - cert-manager should handle this via ingress
		// But having this handler prevents 405 Method Not Allowed errors
		http.NotFound(w, r)
	}).Methods("GET")

	// Health check endpoint (unprotected)
	router.HandleFunc("/health", healthhandler.HealthHandler(logger)).Methods("GET")

	// v1 API routes (protected with auth middleware)
	v1 := router.PathPrefix("/v1").Subrouter()

	// Handle OPTIONS requests for v1 routes as well
	addCORSPreflightHandlers(v1)

	v1.Use(middlewares.AuthMiddleware)

	// Basic access check - user must have any access (not no_access role)
	v1.Use(middlewares.RequireAccess())

	// User endpoint - accessible to all authenticated users with access
	v1.HandleFunc("/user", userhandler.UserHandler(logger)).Methods("GET")

	// Transaction endpoints - require user role or higher
	transactionsRouter := v1.PathPrefix("/transactions").Subrouter()
	transactionsRouter.Use(middlewares.RequireMinimumRole(entities.RoleUser))
	transactionsRouter.HandleFunc("", transactionshandler.TransactionsHandler(services.GlobalTransactionService)).Methods("GET")
	transactionsRouter.HandleFunc("/by-id", transactionshandler.TransactionByIDHandler(services.GlobalTransactionService)).Methods("GET")
	transactionsRouter.HandleFunc("/refresh-cache", transactionshandler.RefreshCacheHandler(services.GlobalTransactionService)).Methods("POST")

	// Admin endpoints - require admin role
	adminRouter := v1.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middlewares.RequireRole(entities.RoleUser))
	adminRouter.HandleFunc("/users", adminhandler.ListUsersHandler(logger)).Methods("GET")
	adminRouter.HandleFunc("/assign-role", adminhandler.AssignRoleHandler(logger)).Methods("POST")
	adminRouter.HandleFunc("/background-fetcher-status", adminhandler.BackgroundFetcherStatusHandler(logger)).Methods("GET")

	// Catch-all handler for unmatched routes - must be last
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If it's an ACME challenge path, return 404 (cert-manager should handle via ingress)
		if strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
			http.NotFound(w, r)
			return
		}
		// For other unmatched paths, return 404
		http.NotFound(w, r)
	})
}
