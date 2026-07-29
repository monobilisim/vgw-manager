// Package api provides the HTTP API server for vgw-manager.
package api

import (
	"net/http"
	"time"
)

const apiPrefix = "/v1"

// NewServer creates an *http.Server with all routes registered and
// sensible timeouts. The token is required for all routes except /healthz.
func NewServer(version string) *http.Server {
	mux := http.NewServeMux()

	// Health check — no auth required.
	mux.HandleFunc("GET /healthz", handleHealth)

	// Bucket routes.
	mux.HandleFunc("GET "+apiPrefix+"/buckets", handleListBuckets)
	mux.HandleFunc("POST "+apiPrefix+"/buckets", mutating(handleCreateBucket))
	mux.HandleFunc("DELETE "+apiPrefix+"/buckets/{name}", mutating(handleDeleteBucket))
	mux.HandleFunc("POST "+apiPrefix+"/buckets/{name}/public", mutating(handleMakePublic))
	mux.HandleFunc("POST "+apiPrefix+"/buckets/{name}/private", mutating(handleMakePrivate))

	// User routes.
	mux.HandleFunc("GET "+apiPrefix+"/users", handleListUsers)
	mux.HandleFunc("GET "+apiPrefix+"/users/{access}", handleGetUser)
	mux.HandleFunc("POST "+apiPrefix+"/users", mutating(handleCreateUser))
	mux.HandleFunc("DELETE "+apiPrefix+"/users/{access}", mutating(handleDeleteUser))

	// Provision route.
	mux.HandleFunc("POST "+apiPrefix+"/provision", mutating(handleProvision))

	// Catch-all for unknown routes → 404.
	mux.HandleFunc("/", handleNotFound)

	handler := chain(
		mux,
		recoveryMiddleware,
		loggingMiddleware,
		authMiddleware,
	)

	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

// handleHealth returns a simple health check response. No auth required.
func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleNotFound returns 404 for any unmatched route.
func handleNotFound(w http.ResponseWriter, r *http.Request) {
	// Go 1.22 ServeMux already returns 405 for wrong-method matches on
	// registered patterns; this catches truly unknown paths.
	writeError(w, http.StatusNotFound, errNotFound)
}

