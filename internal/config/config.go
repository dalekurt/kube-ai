package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AIPersona defines an AI assistant personality
type AIPersona struct {
	Description  string `json:"description"`
	SystemPrompt string `json:"systemPrompt"`
}

// DefaultPersonas provides a set of predefined AI personas
var DefaultPersonas = map[string]AIPersona{
	"kubernetes-expert": {
		Description:  "Deep Kubernetes expertise with technical, detailed responses",
		SystemPrompt: "You are a highly knowledgeable Kubernetes expert with deep technical understanding of Kubernetes architecture, deployment patterns, and troubleshooting. Provide detailed, accurate, and technical responses focusing on best practices and proper Kubernetes patterns.",
	},
	"devops-engineer": {
		Description:  "DevOps-focused approach with CI/CD and automation expertise",
		SystemPrompt: "You are a DevOps engineer specializing in Kubernetes and cloud-native technologies. Focus on CI/CD practices, automation, and infrastructure as code approaches when answering questions. Provide practical, implementation-focused advice that helps establish reliable DevOps practices.",
	},
	"teacher": {
		Description:  "Simplified explanations with examples for learning",
		SystemPrompt: "You are a Kubernetes teacher who specializes in explaining complex concepts in simple terms. Use analogies, examples, and clear step-by-step explanations to help the user understand Kubernetes concepts. Focus on building understanding rather than just solving problems.",
	},
	"security-specialist": {
		Description:  "Focus on security best practices and vulnerability mitigation",
		SystemPrompt: "You are a Kubernetes security specialist with expertise in securing container workloads and Kubernetes clusters. When answering questions, emphasize security best practices, point out potential vulnerabilities, and suggest proper security controls and configurations to mitigate risks.",
	},
	"concise": {
		Description:  "Brief, to-the-point responses without extra explanation",
		SystemPrompt: "You are a Kubernetes advisor who provides extremely concise, accurate responses. Minimize explanation and focus on direct, actionable answers. Use bullet points where appropriate and never include unnecessary information.",
	},
}

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

	// Persona configuration
	ActivePersona  string               `json:"activePersona"`
	CustomPersonas map[string]AIPersona `json:"customPersonas"`
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
	config := &Config{
		CustomPersonas: make(map[string]AIPersona),
	}

	// Try to load from saved config file first
	configPath, err := getConfigFilePath()
	if err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			if err := json.Unmarshal(data, config); err == nil {
				// Successfully loaded from config file

				// Initialize custom personas map if it's nil
				if config.CustomPersonas == nil {
					config.CustomPersonas = make(map[string]AIPersona)
				}

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

	// Set default persona
	config.ActivePersona = os.Getenv("KUBE_AI_PERSONA")
	if config.ActivePersona == "" {
		config.ActivePersona = "kubernetes-expert" // Default persona
	}

	// Save the initial config
	if err := config.SaveConfig(); err != nil {
		// Log the error but continue, as this is not critical
		fmt.Printf("Warning: Failed to save initial configuration: %v\n", err)
	}

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
	if err := c.SaveConfig(); err != nil {
		fmt.Printf("Warning: Failed to save provider configuration: %v\n", err)
	}
}

// UpdateModel updates the default model for the current provider
func (c *Config) UpdateModel(model string) {
	c.DefaultModel = model
	if err := c.SaveConfig(); err != nil {
		fmt.Printf("Warning: Failed to save model configuration: %v\n", err)
	}
}

// GetCurrentPersona returns the currently active persona
func (c *Config) GetCurrentPersona() AIPersona {
	// Check if active persona exists in custom personas
	if persona, ok := c.CustomPersonas[c.ActivePersona]; ok {
		return persona
	}

	// Check if it's a built-in persona
	if persona, ok := DefaultPersonas[c.ActivePersona]; ok {
		return persona
	}

	// Return default persona if active one doesn't exist
	return DefaultPersonas["kubernetes-expert"]
}

// ListPersonas returns all available personas (built-in and custom)
func (c *Config) ListPersonas() map[string]AIPersona {
	// Combine default and custom personas
	allPersonas := make(map[string]AIPersona)

	// Add default personas
	for id, persona := range DefaultPersonas {
		allPersonas[id] = persona
	}

	// Add custom personas
	for id, persona := range c.CustomPersonas {
		allPersonas[id] = persona
	}

	return allPersonas
}

// SetPersona sets the active persona by ID
func (c *Config) SetPersona(personaID string) error {
	// Check if the persona exists
	if _, ok := DefaultPersonas[personaID]; !ok {
		if _, ok := c.CustomPersonas[personaID]; !ok {
			return fmt.Errorf("persona '%s' not found", personaID)
		}
	}

	c.ActivePersona = personaID
	if err := c.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save persona configuration: %w", err)
	}

	return nil
}

// AddCustomPersona adds a new custom persona
func (c *Config) AddCustomPersona(name string, description string, systemPrompt string) error {
	// Initialize personas if nil
	if c.CustomPersonas == nil {
		c.CustomPersonas = make(map[string]AIPersona)
	}

	// Check if name is reserved
	if name == "default" || name == "kubernetes-expert" || name == "devops-engineer" {
		return fmt.Errorf("cannot override built-in persona '%s'", name)
	}

	// Validate the required fields
	if systemPrompt == "" {
		return fmt.Errorf("system prompt is required")
	}

	// Add persona
	c.CustomPersonas[name] = AIPersona{
		Description:  description,
		SystemPrompt: systemPrompt,
	}

	// Save config
	return c.SaveConfig()
}

// RemoveCustomPersona removes a custom persona
func (c *Config) RemoveCustomPersona(id string) error {
	// Check if we're trying to remove a built-in persona
	if _, exists := DefaultPersonas[id]; exists {
		return fmt.Errorf("cannot remove built-in persona '%s'", id)
	}

	// Check if the persona exists
	if _, ok := c.CustomPersonas[id]; !ok {
		return fmt.Errorf("custom persona '%s' not found", id)
	}

	// If we're removing the active persona, switch to default
	if c.ActivePersona == id {
		c.ActivePersona = "kubernetes-expert"
	}

	// Remove the persona
	delete(c.CustomPersonas, id)

	if err := c.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration after removing persona: %w", err)
	}

	return nil
}
