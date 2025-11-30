package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/polonkoevv/linkchecker/internal/api/http/handlers/links"
	"github.com/polonkoevv/linkchecker/internal/api/http/server"
	"github.com/polonkoevv/linkchecker/internal/config"
	"github.com/polonkoevv/linkchecker/internal/service/link"
	"github.com/polonkoevv/linkchecker/internal/storage/inmemory"
)

// App wires together configuration, storage, services and HTTP server.
type App struct {
	cfg     *config.Config
	storage *inmemory.Storage
	server  *http.Server
}

// New constructs the application with all required dependencies.
func New(cfg *config.Config) (*App, error) {
	stg := inmemory.New()
	if err := stg.LoadFromFile(cfg.Storage.FileStoragePath); err != nil {
		return nil, fmt.Errorf("load storage from file: %w", err)
	}
	slog.Info("in-memory storage initialized", slog.String("file", cfg.Storage.FileStoragePath))

	srv := link.New(stg, cfg.Server.MaxWorkersNum)

	handler := links.New(srv, cfg.Server.RequestTimeout)
	mux := server.ConfigRoutes(handler)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	httpServer := server.NewServer(
		addr,
		mux,
		cfg.Server.ReadHeaderTimeout,
		cfg.Server.ReadTimeout,
		cfg.Server.WriteTimeout,
		cfg.Server.IdleTimeout,
	)

	return &App{
		cfg:     cfg,
		storage: stg,
		server:  httpServer,
	}, nil
}

// Run starts the HTTP server and handles graceful shutdown and persistence.
func (a *App) Run(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%s", a.cfg.Server.Host, a.cfg.Server.Port)

	// start HTTP server in background
	go func() {
		slog.Info("starting http server", slog.String("addr", addr))
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server error", slog.Any("error", err))
		}
	}()

	// wait for cancellation (signal from main)
	<-ctx.Done()
	slog.Info("shutdown signal received")

	// give server some time to finish active requests
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", slog.Any("error", err))
	} else {
		slog.Info("server shutdown gracefully")
	}

	// persist storage after server has stopped
	if err := a.storage.SaveToFile(a.cfg.Storage.FileStoragePath); err != nil {
		slog.Error("failed to save storage to file", slog.Any("error", err))
		return err
	}

	slog.Info("storage saved to file", slog.String("file", a.cfg.Storage.FileStoragePath))
	return nil
}
