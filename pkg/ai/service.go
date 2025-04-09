package ai

import (
	"context"
	"fmt"
	"strings"

	"kube-ai/internal/config"
	"kube-ai/pkg/ai/providers"
)

// Service provides AI capabilities for Kubernetes operations
type Service struct {
	provider providers.Provider
	config   *config.Config
}

// NewService creates a new AI service
func NewService(cfg *config.Config) *Service {
	// Create provider based on configuration
	providerType := providers.ProviderType(cfg.AIProvider)
	providerConfig := providers.ProviderConfig{
		BaseURL:   cfg.GetProviderURL(cfg.AIProvider),
		APIKey:    cfg.GetAPIKey(cfg.AIProvider),
		ModelName: cfg.DefaultModel,
	}

	provider, err := providers.CreateProvider(providerType, providerConfig)
	if err != nil {
		// Fallback to Ollama if provider creation fails
		fmt.Printf("Error initializing provider '%s': %v, falling back to Ollama\n", cfg.AIProvider, err)
		provider = providers.NewOllamaProvider(cfg.OllamaURL, cfg.DefaultModel)
		// Also update config to reflect the fallback
		cfg.AIProvider = "ollama"
		cfg.SaveConfig()
	}

	return &Service{
		provider: provider,
		config:   cfg,
	}
}

// SwitchProvider changes the AI provider
func (s *Service) SwitchProvider(providerName string) error {
	providerType := providers.ProviderType(providerName)

	// Check if provider type is supported
	supported := false
	for _, pt := range providers.GetProviderTypes() {
		if pt == providerType {
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Update configuration
	s.config.AIProvider = providerName

	// Set default model based on provider if not already set
	switch providerName {
	case "ollama":
		if s.config.DefaultModel == "" || !strings.Contains(s.config.DefaultModel, "llama") {
			s.config.DefaultModel = "llama3.3"
		}
	case "openai":
		if s.config.DefaultModel == "" || strings.Contains(s.config.DefaultModel, "llama") {
			s.config.DefaultModel = "gpt-3.5-turbo"
		}
	case "anthropic":
		if s.config.DefaultModel == "" || strings.Contains(s.config.DefaultModel, "llama") {
			s.config.DefaultModel = "claude-3-haiku-20240307"
		}
	case "gemini":
		if s.config.DefaultModel == "" || strings.Contains(s.config.DefaultModel, "llama") {
			s.config.DefaultModel = "gemini-1.5-pro"
		}
	case "anythingllm":
		s.config.DefaultModel = "default"
	}

	// Create new provider
	providerConfig := providers.ProviderConfig{
		BaseURL:   s.config.GetProviderURL(providerName),
		APIKey:    s.config.GetAPIKey(providerName),
		ModelName: s.config.DefaultModel,
	}

	provider, err := providers.CreateProvider(providerType, providerConfig)
	if err != nil {
		return fmt.Errorf("error creating provider: %w", err)
	}

	// Update service provider
	s.provider = provider

	// Save configuration
	if err := s.config.SaveConfig(); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	return nil
}

// SetModelName sets the model name for the current provider
func (s *Service) SetModelName(modelName string) {
	s.provider.SetModelName(modelName)
	s.config.UpdateModel(modelName)
}

// GetProvider returns the current provider
func (s *Service) GetProvider() providers.Provider {
	return s.provider
}

// AnalyzeDeployment analyzes a Kubernetes deployment
func (s *Service) AnalyzeDeployment(deploymentYAML string) (string, error) {
	prompt := fmt.Sprintf("Analyze this Kubernetes deployment and provide insights and recommendations:\n\n%s", deploymentYAML)

	return s.provider.GenerateResponse(prompt, 0.7)
}

// OptimizeResources suggests optimizations for resource usage
func (s *Service) OptimizeResources(resourcesYAML string) (string, error) {
	prompt := fmt.Sprintf("Suggest optimizations for these Kubernetes resource definitions to improve efficiency and performance:\n\n%s", resourcesYAML)

	return s.provider.GenerateResponse(prompt, 0.7)
}

// SuggestScalingStrategy suggests scaling strategies
func (s *Service) SuggestScalingStrategy(metricsData, currentConfig string) (string, error) {
	prompt := fmt.Sprintf("Based on the following metrics and current configuration, suggest an optimal scaling strategy for this Kubernetes workload:\n\nMetrics:\n%s\n\nCurrent Configuration:\n%s",
		metricsData, currentConfig)

	return s.provider.GenerateResponse(prompt, 0.7)
}

// GenerateManifest generates a Kubernetes manifest
func (s *Service) GenerateManifest(description string) (string, error) {
	prompt := fmt.Sprintf("Generate a valid Kubernetes manifest for the following description:\n\n%s\n\nPlease provide a complete YAML manifest.",
		description)

	return s.provider.GenerateResponse(prompt, 0.7)
}

// ExplainError explains Kubernetes errors
func (s *Service) ExplainError(errorMessage string) (string, error) {
	prompt := fmt.Sprintf("Explain the following Kubernetes error in simple terms and suggest how to fix it:\n\n%s",
		errorMessage)

	return s.provider.GenerateResponse(prompt, 0.7)
}

// Chat allows general conversation about Kubernetes
func (s *Service) Chat(userMessage string) (string, error) {
	systemPrompt := "You are a helpful Kubernetes expert assistant. Provide accurate, concise information about Kubernetes concepts, resources, and best practices."

	return s.provider.ChatCompletion(systemPrompt, userMessage, 0.7)
}

// ListModels lists available models from the current provider
func (s *Service) ListModels() (string, error) {
	return s.provider.ListModels()
}

// ListProviders returns a list of available AI providers
func (s *Service) ListProviders() string {
	var buf strings.Builder
	buf.WriteString("Available AI Providers:\n")

	currentProvider := s.provider.GetName()

	for _, provider := range providers.GetProviderTypes() {
		var active string
		if string(provider) == currentProvider {
			active = " (active)"
		}
		buf.WriteString(fmt.Sprintf("- %s%s\n", provider, active))
	}

	return buf.String()
}

// GetCurrentProvider returns the name of the currently active provider
func (s *Service) GetCurrentProvider() string {
	return s.provider.GetName()
}

// GetCurrentModel returns the name of the currently active model
func (s *Service) GetCurrentModel() string {
	return s.provider.GetModelName()
}

// Query sends a single query to the AI provider and returns the response
func (s *Service) Query(ctx context.Context, prompt string) (string, error) {
	response, err := s.provider.GenerateCompletion(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("error querying AI provider: %w", err)
	}
	return response, nil
}
