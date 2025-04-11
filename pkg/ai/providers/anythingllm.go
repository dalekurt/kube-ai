package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// AnythingLLMProvider implements the Provider interface for AnythingLLM
type AnythingLLMProvider struct {
	config ProviderConfig
	client *http.Client
}

// AnythingLLMChatRequest represents a chat request to the AnythingLLM API
type AnythingLLMChatRequest struct {
	Message      string  `json:"message"`
	SystemPrompt string  `json:"systemPrompt,omitempty"`
	Temperature  float32 `json:"temperature"`
	Stream       bool    `json:"stream"`
}

// AnythingLLMChatResponse represents a response from the AnythingLLM API
type AnythingLLMChatResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result"`
	Error   string `json:"error,omitempty"`
}

// NewAnythingLLMProvider creates a new AnythingLLM provider
func NewAnythingLLMProvider(baseURL string, apiKey string) *AnythingLLMProvider {
	if baseURL == "" {
		baseURL = "http://localhost:3001"
	}

	return &AnythingLLMProvider{
		config: ProviderConfig{
			BaseURL: baseURL,
			APIKey:  apiKey,
		},
		client: &http.Client{},
	}
}

// GenerateResponse generates a response for a prompt
func (p *AnythingLLMProvider) GenerateResponse(prompt string, temperature float64) (string, error) {
	return p.ChatCompletion("", prompt, float32(temperature))
}

// ChatCompletion generates a response from a conversation
func (p *AnythingLLMProvider) ChatCompletion(systemPrompt string, userMessage string, temperature float32) (string, error) {
	request := AnythingLLMChatRequest{
		Message:     userMessage,
		Temperature: temperature,
		Stream:      false,
	}

	if systemPrompt != "" {
		request.SystemPrompt = systemPrompt
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", p.config.BaseURL+"/api/chat", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add API key if it exists
	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to AnythingLLM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from AnythingLLM API: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response AnythingLLMChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if !response.Success {
		return "", fmt.Errorf("error from AnythingLLM: %s", response.Error)
	}

	return response.Result, nil
}

// ListModels returns a list of available models from AnythingLLM
func (p *AnythingLLMProvider) ListModels() (string, error) {
	req, err := http.NewRequest("GET", p.config.BaseURL+"/api/model/list", nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Add API key if it exists
	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to AnythingLLM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from AnythingLLM API: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	type ModelResponse struct {
		Success bool `json:"success"`
		Models  []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"models"`
	}

	var response ModelResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if !response.Success {
		return "", fmt.Errorf("error getting models from AnythingLLM")
	}

	var buf strings.Builder
	buf.WriteString("Available AnythingLLM Models:\n")

	for _, model := range response.Models {
		buf.WriteString(fmt.Sprintf("- %s\n", model.Name))
	}

	if len(response.Models) == 0 {
		buf.WriteString("No models available. Please check your AnythingLLM configuration.\n")
	}

	return buf.String(), nil
}

// GetName returns the name of the provider
func (p *AnythingLLMProvider) GetName() string {
	return "anythingllm"
}

// GetModelName returns the name of the currently used model
func (p *AnythingLLMProvider) GetModelName() string {
	// AnythingLLM doesn't have a concept of model selection at the API level
	// It uses the model configured in the server
	return "default"
}

// SetModelName sets the model to use
func (p *AnythingLLMProvider) SetModelName(modelName string) {
	// This is a no-op for AnythingLLM as the model is set on the server side
}

// RequiresAPIKey returns true if the provider requires an API key
func (p *AnythingLLMProvider) RequiresAPIKey() bool {
	// AnythingLLM might require an API key depending on its configuration
	// We'll return false as it can work without one in local deployments
	return false
}

// GenerateCompletion sends a prompt to AnythingLLM and returns the response
func (p *AnythingLLMProvider) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	// For AnythingLLM, we'll use the existing ChatCompletion method with an empty system prompt
	systemPrompt := "You are a Kubernetes expert assistant. Provide concise, accurate information about Kubernetes concepts, troubleshooting, and best practices."

	return p.ChatCompletion(systemPrompt, prompt, 0.7)
}
