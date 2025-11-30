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

// Default values
const (
	defaultHost              = "localhost"
	defaultPort              = "8080"
	defaultReadHeaderTimeout = 5   // seconds
	defaultReadTimeout       = 10  // seconds
	defaultWriteTimeout      = 10  // seconds
	defaultIdleTimeout       = 120 // seconds
	defaultRequestTimeout    = 30  // seconds
	defaultMaxWorkersNum     = 4
	defaultLogLevel          = "info"
	defaultLogPath           = "logs/app.log"
	defaultFileStoragePath   = "storage/links.json"
)

// MustLoad loads configuration or panics if it fails.
func MustLoad() *Config {
	cfg, err := load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	return cfg
}

// getEnvString returns environment variable value or default if empty.
func getEnvString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt returns environment variable value as int or default if empty/invalid.
func getEnvInt(key string, defaultValue int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("failed to convert %s to int: %w", key, err)
	}
	if intValue <= 0 {
		return 0, fmt.Errorf("%s must be positive, got: %d", key, intValue)
	}
	return intValue, nil
}

// validateRequired checks that required string values are not empty.
func validateRequired(key, value string) error {
	if value == "" {
		return fmt.Errorf("required environment variable %s is not set", key)
	}
	return nil
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

	// HTTP Server load with validation
	cfg.Server.Host = getEnvString("HOST", defaultHost)
	if err := validateRequired("HOST", cfg.Server.Host); err != nil {
		return nil, err
	}

	cfg.Server.Port = getEnvString("PORT", defaultPort)
	if err := validateRequired("PORT", cfg.Server.Port); err != nil {
		return nil, err
	}

	readHeaderTimeout, err := getEnvInt("READ_HEADER_TIMEOUT", defaultReadHeaderTimeout)
	if err != nil {
		return nil, fmt.Errorf("READ_HEADER_TIMEOUT: %w", err)
	}
	cfg.Server.ReadHeaderTimeout = time.Duration(readHeaderTimeout) * time.Second

	readTimeout, err := getEnvInt("READ_TIMEOUT", defaultReadTimeout)
	if err != nil {
		return nil, fmt.Errorf("READ_TIMEOUT: %w", err)
	}
	cfg.Server.ReadTimeout = time.Duration(readTimeout) * time.Second

	writeTimeout, err := getEnvInt("WRITE_TIMEOUT", defaultWriteTimeout)
	if err != nil {
		return nil, fmt.Errorf("WRITE_TIMEOUT: %w", err)
	}
	cfg.Server.WriteTimeout = time.Duration(writeTimeout) * time.Second

	idleTimeout, err := getEnvInt("IDLE_TIMEOUT", defaultIdleTimeout)
	if err != nil {
		return nil, fmt.Errorf("IDLE_TIMEOUT: %w", err)
	}
	cfg.Server.IdleTimeout = time.Duration(idleTimeout) * time.Second

	requestTimeout, err := getEnvInt("REQUEST_TIMEOUT", defaultRequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("REQUEST_TIMEOUT: %w", err)
	}
	cfg.Server.RequestTimeout = time.Duration(requestTimeout) * time.Second

	maxWorkersNum, err := getEnvInt("MAX_WORKERS_NUM", defaultMaxWorkersNum)
	if err != nil {
		return nil, fmt.Errorf("MAX_WORKERS_NUM: %w", err)
	}
	cfg.Server.MaxWorkersNum = maxWorkersNum

	// Logger load with defaults
	cfg.Logger.LevelInfo = getEnvString("LEVEL_INFO", defaultLogLevel)
	cfg.Logger.LogPath = getEnvString("LOGGING_PATH", defaultLogPath)

	// Storage load with default
	cfg.Storage.FileStoragePath = getEnvString("FILE_STORAGE_PATH", defaultFileStoragePath)

	return &cfg, nil
}

func getConfigPath() string {
	configPath := os.Getenv(Path)
	if configPath == "" {
		configPath = defaultConfigPath
	}

	return configPath
}
