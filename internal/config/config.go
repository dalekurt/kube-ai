package config

import (
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	KubeConfigPath string
	AIApiKey       string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	config := &Config{}

	// Try to load kubeconfig from standard location
	homeDir, err := os.UserHomeDir()
	if err == nil {
		config.KubeConfigPath = filepath.Join(homeDir, ".kube", "config")
	}

	// Override with environment variables if present
	if kubePath := os.Getenv("KUBECONFIG"); kubePath != "" {
		config.KubeConfigPath = kubePath
	}

	// Load AI API key from environment
	config.AIApiKey = os.Getenv("AI_API_KEY")

	return config
} 