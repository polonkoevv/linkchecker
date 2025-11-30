package server

import (
	"net/http"
	"time"

	"github.com/polonkoevv/linkchecker/internal/api/http/handlers/links"
	"github.com/polonkoevv/linkchecker/internal/api/http/middleware"
)

// ConfigRoutes registers HTTP routes for link operations with middleware and returns a mux.
func ConfigRoutes(linksHandler *links.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	// Middleware chain for POST requests (validation + logging)
	postMiddleware := middleware.Chain(
		middleware.Logging,
		middleware.ValidateBodySize,
		middleware.ValidateJSONContentType,
		middleware.ValidateJSONStructure,
	)

	// Middleware chain for GET requests (only logging)
	getMiddleware := middleware.Chain(
		middleware.Logging,
	)

	mux.HandleFunc("POST /links", postMiddleware(linksHandler.Check))
	mux.HandleFunc("GET /links", getMiddleware(linksHandler.GetAll))
	mux.HandleFunc("POST /report", postMiddleware(linksHandler.GenerateReport))

	return mux
}

// NewServer constructs an http.Server with the provided address, handler and timeouts.
func NewServer(addr string, mux *http.ServeMux, readHeaderTimeout, readTimeout, writeTimeout, idleTimeout time.Duration) *http.Server {

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}
