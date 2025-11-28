package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server HTTPConfig   `validate:"true"`
	Logger LoggerConfig `validate:"true"`
}

type HTTPConfig struct {
	Host string `validate:"true"`
	Port string `validate:"true"`
}

type LoggerConfig struct {
	LevelInfo string `validate:"true"`
	LogPath   string `validate:"true"`
}

const (
	defaultConfigPath = ".env"
	Path              = "CONFIG_PATH"
)

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

	// Logger load
	cfg.Logger.LevelInfo = os.Getenv("LEVEL_INFO")
	cfg.Logger.LogPath = os.Getenv("LOGGING_PATH")

	return &cfg, nil
}

func getConfigPath() string {
	configPath := os.Getenv(Path)
	if configPath == "" {
		configPath = defaultConfigPath
	}

	return configPath
}
