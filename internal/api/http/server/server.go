package server

import (
	"net/http"
	"time"

	"github.com/polonkoevv/linkchecker/internal/api/http/handlers/links"
)

// ConfigRoutes registers HTTP routes for link operations and returns a mux.
func ConfigRoutes(linksHandler *links.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /links", linksHandler.Check)
	mux.HandleFunc("GET /links", linksHandler.GetAll)
	mux.HandleFunc("POST /report", linksHandler.GenerateReport)

	return mux
}

// NewServer constructs an http.Server with the provided address, handler and timeouts.
func NewServer(addr string, mux *http.ServeMux, readHeaderTimeout, readTimeout, writeTimeout, IdleTimeout time.Duration) *http.Server {

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       IdleTimeout,
	}
}
