package main

import (
	"log/slog"
	"os"

	"github.com/polonkoevv/linkchecker/internal/config"
	"github.com/polonkoevv/linkchecker/internal/logger"
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

	slog.Info("Loaded config")
}
