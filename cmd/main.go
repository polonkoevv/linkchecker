package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/polonkoevv/linkchecker/internal/api/http/handlers/links"
	"github.com/polonkoevv/linkchecker/internal/api/http/server"
	"github.com/polonkoevv/linkchecker/internal/config"
	"github.com/polonkoevv/linkchecker/internal/logger"
	"github.com/polonkoevv/linkchecker/internal/pdfgenerator"
	"github.com/polonkoevv/linkchecker/internal/service/link"
	"github.com/polonkoevv/linkchecker/internal/storage/inmemory"
)

func main() {

	cfg := config.MustLoad()

	appLogger, closeLogFile, err := logger.SetupLogger(cfg.Logger.LogPath, cfg.Logger.LevelInfo)
	if err != nil {
		slog.Error("Error while setting up logger", slog.Any("error", err))
		os.Exit(1)
	}
	slog.SetDefault(appLogger)
	defer closeLogFile()

	stg := inmemory.New()

	pdfGen := pdfgenerator.GoFPDFGenerator{}

	srv := link.New(stg, 5*time.Second, &pdfGen)

	handler := links.New(srv)

	mux := server.ConfigRoutes(handler)

	server := server.NewServer(fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port), mux)

	server.ListenAndServe()
}
