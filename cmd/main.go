package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/polonkoevv/linkchecker/internal/app"
	"github.com/polonkoevv/linkchecker/internal/config"
	"github.com/polonkoevv/linkchecker/internal/logger"
)

func main() {
	cfg := config.MustLoad()

	appLogger, closeLogFile, err := logger.SetupLogger(cfg.Logger.LogPath, cfg.Logger.LevelInfo)
	if err != nil {
		slog.Error("error while setting up logger", slog.Any("error", err))
		os.Exit(1)
	}
	slog.SetDefault(appLogger)
	defer func() {
		if err := closeLogFile(); err != nil {
			slog.Error("failed to close log file", slog.Any("error", err))
		}
	}()

	a, err := app.New(cfg)
	if err != nil {
		slog.Error("failed to initialize app", slog.Any("error", err))
		return // defer выполнится автоматически
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := a.Run(ctx); err != nil && err != http.ErrServerClosed {
		slog.Error("app stopped with error", slog.Any("error", err))
		return // defer выполнится автоматически
	}
}
