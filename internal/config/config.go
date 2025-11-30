package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config aggregates all application configuration sections.
type Config struct {
	Server  HTTPConfig
	Logger  LoggerConfig
	Storage StorageConfig
}

// StorageConfig holds configuration for persistence layer.
type StorageConfig struct {
	FileStoragePath string
}

// HTTPConfig contains HTTP server address and timeout settings.
type HTTPConfig struct {
	Host              string
	Port              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	RequestTimeout    time.Duration
	MaxWorkersNum     int
}

// LoggerConfig describes logging level and destination file.
type LoggerConfig struct {
	LevelInfo string
	LogPath   string
}

const (
	defaultConfigPath = ".env"
	Path              = "CONFIG_PATH"
)

// MustLoad loads configuration or panics if it fails.
func MustLoad() *Config {
	cfg, err := load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	return cfg
}

func load() (*Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		if loadErr := godotenv.Load(); loadErr != nil {
			return nil, fmt.Errorf("failed to load .env: %w", loadErr)
		}
	}

	configPath := getConfigPath()
	if configPath != "" {
		fileInfo, err := os.Stat(configPath)

		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file %s does not exist", configPath)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to access config file: %w", err)
		}

		if fileInfo.IsDir() {
			return nil, fmt.Errorf("config path is a directory, not a file: %s", configPath)
		}
	}

	var cfg Config

	// HTTP Server load
	cfg.Server.Host = os.Getenv("HOST")
	cfg.Server.Port = os.Getenv("PORT")

	tempTll, err := strconv.Atoi(os.Getenv("READ_HEADER_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert timeout")
	}
	cfg.Server.ReadHeaderTimeout = time.Duration(tempTll) * time.Second

	tempTll, err = strconv.Atoi(os.Getenv("READ_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert timeout")
	}
	cfg.Server.ReadTimeout = time.Duration(tempTll) * time.Second

	tempTll, err = strconv.Atoi(os.Getenv("WRITE_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert timeout")
	}
	cfg.Server.WriteTimeout = time.Duration(tempTll) * time.Second

	tempTll, err = strconv.Atoi(os.Getenv("IDLE_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert timeout")
	}
	cfg.Server.IdleTimeout = time.Duration(tempTll) * time.Second

	tempTll, err = strconv.Atoi(os.Getenv("REQUEST_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert timeout")
	}
	cfg.Server.RequestTimeout = time.Duration(tempTll) * time.Second

	maxWorkersNum, err := strconv.Atoi(os.Getenv("MAX_WORKERS_NUM"))
	if err != nil {
		maxWorkersNum = 4
	}
	cfg.Server.MaxWorkersNum = maxWorkersNum

	// Logger load
	cfg.Logger.LevelInfo = os.Getenv("LEVEL_INFO")
	cfg.Logger.LogPath = os.Getenv("LOGGING_PATH")

	// Storage load
	cfg.Storage.FileStoragePath = os.Getenv("FILE_STORAGE_PATH")

	return &cfg, nil
}

func getConfigPath() string {
	configPath := os.Getenv(Path)
	if configPath == "" {
		configPath = defaultConfigPath
	}

	return configPath
}
