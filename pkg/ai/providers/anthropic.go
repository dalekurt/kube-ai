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

// AnthropicProvider implements the Provider interface for Anthropic
type AnthropicProvider struct {
	config ProviderConfig
	client *http.Client
}

// AnthropicRequest represents a chat request to the Anthropic API
type AnthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Messages    []AnthropicMessage `json:"messages"`
	Temperature float64            `json:"temperature"`
}

// AnthropicMessage represents a message in a conversation
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents a response from the Anthropic API
type AnthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey string, modelName string) *AnthropicProvider {
	if modelName == "" {
		modelName = "claude-3-haiku-20240307"
	}

	return &AnthropicProvider{
		config: ProviderConfig{
			BaseURL:   "https://api.anthropic.com",
			APIKey:    apiKey,
			ModelName: modelName,
		},
		client: &http.Client{},
	}
}

// GenerateResponse generates a response for a prompt
func (p *AnthropicProvider) GenerateResponse(prompt string, temperature float64) (string, error) {
	// For Anthropic, we'll use the messages API with a user message
	return p.ChatCompletion("", prompt, temperature)
}

// ChatCompletion generates a response from a conversation
func (p *AnthropicProvider) ChatCompletion(systemPrompt string, userMessage string, temperature float64) (string, error) {
	if p.config.APIKey == "" {
		return "", fmt.Errorf("Anthropic API key is required")
	}

	messages := []AnthropicMessage{
		{Role: "user", Content: userMessage},
	}

	// Add system prompt if provided
	if systemPrompt != "" {
		messages = []AnthropicMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		}
	}

	request := AnthropicRequest{
		Model:       p.config.ModelName,
		MaxTokens:   4096,
		Messages:    messages,
		Temperature: temperature,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", p.config.BaseURL+"/v1/messages", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.config.APIKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to Anthropic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from Anthropic API: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	// Combine all text blocks in the response
	var result strings.Builder
	for _, content := range response.Content {
		if content.Type == "text" {
			result.WriteString(content.Text)
		}
	}

	return result.String(), nil
}

// ListModels returns a list of available models from Anthropic
func (p *AnthropicProvider) ListModels() (string, error) {
	// Anthropic doesn't have a list models API, so we'll hardcode the available models
	var buf strings.Builder
	buf.WriteString("Available Anthropic Models:\n")
	buf.WriteString("- claude-3-opus-20240229\n")
	buf.WriteString("- claude-3-sonnet-20240229\n")
	buf.WriteString("- claude-3-haiku-20240307\n")
	buf.WriteString("- claude-2.1\n")
	buf.WriteString("- claude-2.0\n")
	buf.WriteString("- claude-instant-1.2\n")

	return buf.String(), nil
}

// GetName returns the name of the provider
func (p *AnthropicProvider) GetName() string {
	return "anthropic"
}

// GetModelName returns the name of the currently used model
func (p *AnthropicProvider) GetModelName() string {
	return p.config.ModelName
}

// SetModelName sets the model to use
func (p *AnthropicProvider) SetModelName(modelName string) {
	p.config.ModelName = modelName
}

// RequiresAPIKey returns true if the provider requires an API key
func (p *AnthropicProvider) RequiresAPIKey() bool {
	return true
}

// GenerateCompletion sends a prompt to Anthropic and returns the response
func (p *AnthropicProvider) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	// For Anthropic, we'll use a Kubernetes-specific system prompt
	systemPrompt := "You are a Kubernetes expert assistant. Provide concise, accurate information about Kubernetes concepts, troubleshooting, and best practices."

	return p.ChatCompletion(systemPrompt, prompt, 0.7)
}
