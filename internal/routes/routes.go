package routes

import (
	"net/http"

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
}
