package helpers

import (
	"net/http"

	"github.com/gorilla/mux"
)

// CORSPreflightHandler handles CORS preflight OPTIONS requests
func CORSPreflightHandler(w http.ResponseWriter, r *http.Request) {
	// CORS headers are already set by CORSMiddleware
	// Just return 200 OK to indicate the preflight is successful
	w.WriteHeader(http.StatusOK)
}

// AddCORSPreflightHandlers adds OPTIONS handlers to a router for CORS preflight support
func AddCORSPreflightHandlers(router *mux.Router) {
	router.Methods("OPTIONS").HandlerFunc(CORSPreflightHandler)
}

// AddCORSPreflightToSubrouter adds OPTIONS handlers to a subrouter for CORS preflight support
func AddCORSPreflightToSubrouter(subrouter *mux.Router) {
	subrouter.Methods("OPTIONS").HandlerFunc(CORSPreflightHandler)
}
