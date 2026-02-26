package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FastCodeConfig holds global configuration loaded from ~/.fastcode/config.yaml.
type FastCodeConfig struct {
	OpenAIAPIKey   string `yaml:"openai_api_key"`
	Model          string `yaml:"model"`
	BaseURL        string `yaml:"base_url"`
	EmbeddingURL   string `yaml:"embedding_url"`   // Separate URL for embedding API
	EmbeddingModel string `yaml:"embedding_model"` // Embedding model name
}

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".fastcode", "config.yaml")
}

// Load reads the YAML config file and sets environment variables.
// Environment variables already set take precedence over the config file.
func Load() (*FastCodeConfig, error) {
	return LoadFrom(DefaultConfigPath())
}

// LoadFrom reads a specific YAML config file and sets environment variables.
func LoadFrom(path string) (*FastCodeConfig, error) {
	cfg := &FastCodeConfig{}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // No config file, not an error
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	// Set env vars only if not already set (env vars take precedence)
	setIfEmpty("OPENAI_API_KEY", cfg.OpenAIAPIKey)
	setIfEmpty("MODEL", cfg.Model)
	setIfEmpty("BASE_URL", cfg.BaseURL)
	setIfEmpty("EMBEDDING_URL", cfg.EmbeddingURL)
	setIfEmpty("EMBEDDING_MODEL", cfg.EmbeddingModel)

	return cfg, nil
}

func setIfEmpty(key, value string) {
	if value != "" && os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
}
