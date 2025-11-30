package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// SetupLogger configures slog logger writing to file and stdout based on level.
func SetupLogger(logFile, logLevel string) (*slog.Logger, func() error, error) {
	if logFile != "" {
		logDir := filepath.Dir(logFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, nil, err
		}
	}

	var fileWriter io.Writer
	var closeFile func() error = func() error { return nil }

	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, nil, err
		}
		fileWriter = file
		closeFile = file.Close
	} else {
		fileWriter = io.Discard
	}

	var writers []io.Writer
	if fileWriter != io.Discard {
		writers = append(writers, fileWriter)
	}
	writers = append(writers, os.Stdout)

	multiWriter := io.MultiWriter(writers...)

	// Настраиваем уровень логирования
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(multiWriter, opts)
	logger := slog.New(handler)

	return logger, closeFile, nil
}
