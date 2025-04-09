package providers

import (
	"fmt"
)

// ProviderType represents the type of AI provider
type ProviderType string

// Provider types
const (
	ProviderTypeOllama      ProviderType = "ollama"
	ProviderTypeOpenAI      ProviderType = "openai"
	ProviderTypeAnthropicAI ProviderType = "anthropic"
	ProviderTypeGemini      ProviderType = "gemini"
	ProviderTypeAnythingLLM ProviderType = "anythingllm"
)

// GetProviderTypes returns a list of supported provider types
func GetProviderTypes() []ProviderType {
	return []ProviderType{
		ProviderTypeOllama,
		ProviderTypeOpenAI,
		ProviderTypeAnthropicAI,
		ProviderTypeGemini,
		ProviderTypeAnythingLLM,
	}
}

// CreateProvider creates a provider of the specified type
func CreateProvider(providerType ProviderType, config ProviderConfig) (Provider, error) {
	switch providerType {
	case ProviderTypeOllama:
		return NewOllamaProvider(config.BaseURL, config.ModelName), nil
	case ProviderTypeOpenAI:
		return NewOpenAIProvider(config.APIKey, config.ModelName), nil
	case ProviderTypeAnthropicAI:
		return NewAnthropicProvider(config.APIKey, config.ModelName), nil
	case ProviderTypeGemini:
		return NewGeminiProvider(config.APIKey, config.ModelName), nil
	case ProviderTypeAnythingLLM:
		return NewAnythingLLMProvider(config.BaseURL, config.APIKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}
