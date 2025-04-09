package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	KubeConfigPath string `json:"kubeConfigPath"`

	// AI Provider configuration
	AIProvider string `json:"aiProvider"`

	// API Keys for various providers
	OpenAIApiKey    string `json:"openaiApiKey"`
	AnthropicApiKey string `json:"anthropicApiKey"`
	GeminiApiKey    string `json:"geminiApiKey"`

	// Provider URLs
	OllamaURL      string `json:"ollamaUrl"`
	AnythingLLMURL string `json:"anythingLlmUrl"`

	// Default model name for the active provider
	DefaultModel string `json:"defaultModel"`
}

// getConfigFilePath returns the path to the configuration file
func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Create the .kube-ai directory if it doesn't exist
	kubeAIDir := filepath.Join(homeDir, ".kube-ai")
	if err := os.MkdirAll(kubeAIDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(kubeAIDir, "config.json"), nil
}

// SaveConfig saves the configuration to a file
func (c *Config) SaveConfig() error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// LoadConfig loads configuration from environment variables or saved config
func LoadConfig() *Config {
	config := &Config{}

	// Try to load from saved config file first
	configPath, err := getConfigFilePath()
	if err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			if err := json.Unmarshal(data, config); err == nil {
				// Successfully loaded from config file
				return config
			}
		}
	}

	// If config file doesn't exist or couldn't be loaded, use defaults and environment variables

	// Try to load kubeconfig from standard location
	homeDir, err := os.UserHomeDir()
	if err == nil {
		config.KubeConfigPath = filepath.Join(homeDir, ".kube", "config")
	}

	// Override with environment variables if present
	if kubePath := os.Getenv("KUBECONFIG"); kubePath != "" {
		config.KubeConfigPath = kubePath
	}

	// Load AI Provider configuration
	config.AIProvider = os.Getenv("AI_PROVIDER")
	if config.AIProvider == "" {
		config.AIProvider = "ollama" // Default provider
	}

	// Load API keys for various providers
	config.OpenAIApiKey = os.Getenv("OPENAI_API_KEY")
	config.AnthropicApiKey = os.Getenv("ANTHROPIC_API_KEY")
	config.GeminiApiKey = os.Getenv("GEMINI_API_KEY")

	// Load provider URLs
	config.OllamaURL = os.Getenv("OLLAMA_URL")
	if config.OllamaURL == "" {
		config.OllamaURL = "http://localhost:11434"
	}

	config.AnythingLLMURL = os.Getenv("ANYTHINGLLM_URL")
	if config.AnythingLLMURL == "" {
		config.AnythingLLMURL = "http://localhost:3001"
	}

	// Load default model based on provider
	switch config.AIProvider {
	case "ollama":
		config.DefaultModel = os.Getenv("OLLAMA_DEFAULT_MODEL")
		if config.DefaultModel == "" {
			config.DefaultModel = "llama3.3"
		}
	case "openai":
		config.DefaultModel = os.Getenv("OPENAI_DEFAULT_MODEL")
		if config.DefaultModel == "" {
			config.DefaultModel = "gpt-3.5-turbo"
		}
	case "anthropic":
		config.DefaultModel = os.Getenv("ANTHROPIC_DEFAULT_MODEL")
		if config.DefaultModel == "" {
			config.DefaultModel = "claude-3-haiku-20240307"
		}
	case "gemini":
		config.DefaultModel = os.Getenv("GEMINI_DEFAULT_MODEL")
		if config.DefaultModel == "" {
			config.DefaultModel = "gemini-1.5-pro"
		}
	case "anythingllm":
		// AnythingLLM doesn't need a default model as it's configured on the server
		config.DefaultModel = "default"
	}

	// Save the initial config
	config.SaveConfig()

	return config
}

// GetAPIKey returns the API key for the specified provider
func (c *Config) GetAPIKey(provider string) string {
	switch provider {
	case "openai":
		return c.OpenAIApiKey
	case "anthropic":
		return c.AnthropicApiKey
	case "gemini":
		return c.GeminiApiKey
	default:
		return ""
	}
}

// GetProviderURL returns the URL for the specified provider
func (c *Config) GetProviderURL(provider string) string {
	switch provider {
	case "ollama":
		return c.OllamaURL
	case "anythingllm":
		return c.AnythingLLMURL
	default:
		return ""
	}
}

// UpdateProvider updates the current AI provider
func (c *Config) UpdateProvider(provider string) {
	c.AIProvider = provider
	c.SaveConfig()
}

// UpdateModel updates the default model for the current provider
func (c *Config) UpdateModel(model string) {
	c.DefaultModel = model
	c.SaveConfig()
}
