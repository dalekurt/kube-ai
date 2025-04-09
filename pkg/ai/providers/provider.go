package providers

// Provider defines the interface for AI service providers
type Provider interface {
	// GenerateResponse generates a response for a prompt
	GenerateResponse(prompt string, temperature float64) (string, error)

	// ChatCompletion generates a response from a conversation
	ChatCompletion(systemPrompt string, userMessage string, temperature float64) (string, error)

	// ListModels returns a list of available models
	ListModels() (string, error)

	// GetName returns the name of the provider
	GetName() string

	// GetModelName returns the name of the currently used model
	GetModelName() string

	// SetModelName sets the model to use
	SetModelName(modelName string)

	// RequiresAPIKey returns true if the provider requires an API key
	RequiresAPIKey() bool
}

// ProviderConfig contains common configuration for providers
type ProviderConfig struct {
	BaseURL   string
	APIKey    string
	ModelName string
}
