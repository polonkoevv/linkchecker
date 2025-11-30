package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	storageFile := "storage.json"
	stg := inmemory.New()
	if err := stg.LoadFromFile(storageFile); err != nil {
		slog.Error("failed to load storage from file", slog.Any("error", err))
	}
	slog.Info("in-memory storage initialized", slog.String("file", storageFile))

	pdfGen := pdfgenerator.GoFPDFGenerator{}
	srv := link.New(stg, 5*time.Second, &pdfGen, 8)
	handler := links.New(srv)
	mux := server.ConfigRoutes(handler)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	httpServer := server.NewServer(addr, mux)

	// контекст с сигналами
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// запускаем сервер
	go func() {
		slog.Info("starting http server", slog.String("addr", addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server error", slog.Any("error", err))
		}
	}()

	// ждём сигнала
	<-ctx.Done()
	slog.Info("shutdown signal received")

	// 1) сохраняем стор в файл
	if err := stg.SaveToFile(storageFile); err != nil {
		slog.Error("failed to save storage to file", slog.Any("error", err))
	} else {
		slog.Info("storage saved to file", slog.String("file", storageFile))
	}

	// 2) даём серверу завершить активные запросы
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", slog.Any("error", err))
	} else {
		slog.Info("server shutdown gracefully")
	}
}
