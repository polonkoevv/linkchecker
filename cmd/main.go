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
		slog.Error("error while setting up logger", slog.Any("error", err))
		os.Exit(1)
	}
	slog.SetDefault(appLogger)
	defer closeLogFile()

	slog.Info("configuration loaded",
		slog.String("host", cfg.Server.Host),
		slog.String("port", cfg.Server.Port),
	)

	stg := inmemory.New()
	slog.Info("in-memory storage initialized")

	pdfGen := pdfgenerator.GoFPDFGenerator{}

	srv := link.New(stg, 5*time.Second, &pdfGen)
	slog.Info("link service initialized")

	handler := links.New(srv)
	mux := server.ConfigRoutes(handler)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	httpServer := server.NewServer(addr, mux)

	slog.Info("starting http server", slog.String("addr", addr))
	if err := httpServer.ListenAndServe(); err != nil {
		slog.Error("http server stopped with error", slog.Any("error", err))
		os.Exit(1)
	}
}
