package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_DefaultValues(t *testing.T) {
	// Test loading config with empty path (should use defaults)
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Check default values
	if cfg.Server.HTTPPort != 8080 {
		t.Errorf("expected default HTTP port 8080, got %d", cfg.Server.HTTPPort)
	}

	if cfg.Server.GRPCPort != 9090 {
		t.Errorf("expected default gRPC port 9090, got %d", cfg.Server.GRPCPort)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected default host 0.0.0.0, got %s", cfg.Server.Host)
	}

	if cfg.OpenWebUI.BaseURL != "http://localhost:3000" {
		t.Errorf("expected default OpenWebUI URL http://localhost:3000, got %s", cfg.OpenWebUI.BaseURL)
	}

	if cfg.OpenWebUI.Timeout != 30 {
		t.Errorf("expected default timeout 30, got %d", cfg.OpenWebUI.Timeout)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	configContent := `
server:
  http_port: 8081
  grpc_port: 9091
  host: "127.0.0.1"

openwebui:
  base_url: "http://test.example.com"
  api_key: "test-key"
  timeout: 60

logging:
  level: "debug"
  format: "json"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load config from file
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Check loaded values
	if cfg.Server.HTTPPort != 8081 {
		t.Errorf("expected HTTP port 8081, got %d", cfg.Server.HTTPPort)
	}

	if cfg.Server.GRPCPort != 9091 {
		t.Errorf("expected gRPC port 9091, got %d", cfg.Server.GRPCPort)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Server.Host)
	}

	if cfg.OpenWebUI.BaseURL != "http://test.example.com" {
		t.Errorf("expected OpenWebUI URL http://test.example.com, got %s", cfg.OpenWebUI.BaseURL)
	}

	if cfg.OpenWebUI.APIKey != "test-key" {
		t.Errorf("expected API key test-key, got %s", cfg.OpenWebUI.APIKey)
	}

	if cfg.OpenWebUI.Timeout != 60 {
		t.Errorf("expected timeout 60, got %d", cfg.OpenWebUI.Timeout)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "json" {
		t.Errorf("expected log format json, got %s", cfg.Logging.Format)
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	_, err := LoadConfig("/non/existent/path/config.yaml")
	if err == nil {
		t.Error("expected error for non-existent config file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create temporary invalid config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid_config.yaml")

	invalidContent := `
server:
  http_port: not_a_number
  invalid_yaml: [
`

	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML config file")
	}
}
