package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	OpenWebUI OpenWebUIConfig `yaml:"openwebui"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	HTTPPort int    `yaml:"http_port"`
	GRPCPort int    `yaml:"grpc_port"`
	Host     string `yaml:"host"`
}

// OpenWebUIConfig holds OpenWebUI API configuration
type OpenWebUIConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
	Timeout int    `yaml:"timeout"` // timeout in seconds
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			HTTPPort: 8080,
			GRPCPort: 9090,
			Host:     "0.0.0.0",
		},
		OpenWebUI: OpenWebUIConfig{
			BaseURL: "http://localhost:3000",
			Timeout: 30,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// Load from file if it exists
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}

			if err := yaml.Unmarshal(data, config); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Override with environment variables
	if port := os.Getenv("HTTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.HTTPPort = p
		}
	}

	if port := os.Getenv("GRPC_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.GRPCPort = p
		}
	}

	if host := os.Getenv("HOST"); host != "" {
		config.Server.Host = host
	}

	if baseURL := os.Getenv("OPENWEBUI_BASE_URL"); baseURL != "" {
		config.OpenWebUI.BaseURL = baseURL
	}

	if apiKey := os.Getenv("OPENWEBUI_API_KEY"); apiKey != "" {
		config.OpenWebUI.APIKey = apiKey
	}

	if timeout := os.Getenv("OPENWEBUI_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			config.OpenWebUI.Timeout = t
		}
	}

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}

	return config, nil
}
