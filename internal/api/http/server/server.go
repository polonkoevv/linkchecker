package server

import (
	"net/http"

	"github.com/polonkoevv/linkchecker/internal/api/http/handlers/links"
)

func ConfigRoutes(linksHandler *links.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /links", linksHandler.Check)
	mux.HandleFunc("GET /links", linksHandler.GetAll)
	mux.HandleFunc("POST /report", linksHandler.GenerateReport)

	return mux
}

func NewServer(addr string, mux *http.ServeMux) *http.Server {

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
