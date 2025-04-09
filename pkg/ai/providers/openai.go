package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	config ProviderConfig
	client *http.Client
}

// OpenAIChatRequest represents a chat request to the OpenAI API
type OpenAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []OpenAIChatMessage `json:"messages"`
	Temperature float64             `json:"temperature"`
}

// OpenAIChatMessage represents a message in a conversation
type OpenAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIChatResponse represents a response from the OpenAI API
type OpenAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index        int               `json:"index"`
		Message      OpenAIChatMessage `json:"message"`
		FinishReason string            `json:"finish_reason"`
	} `json:"choices"`
}

// OpenAIListModelsResponse represents a response from the OpenAI list models API
type OpenAIListModelsResponse struct {
	Data []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int    `json:"created"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string, modelName string) *OpenAIProvider {
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}

	return &OpenAIProvider{
		config: ProviderConfig{
			BaseURL:   "https://api.openai.com/v1",
			APIKey:    apiKey,
			ModelName: modelName,
		},
		client: &http.Client{},
	}
}

// GenerateResponse generates a response for a prompt
func (p *OpenAIProvider) GenerateResponse(prompt string, temperature float64) (string, error) {
	// For OpenAI, we'll just use the chat endpoint with a user message
	return p.ChatCompletion("", prompt, temperature)
}

// ChatCompletion generates a response from a conversation
func (p *OpenAIProvider) ChatCompletion(systemPrompt string, userMessage string, temperature float64) (string, error) {
	if p.config.APIKey == "" {
		return "", fmt.Errorf("OpenAI API key is required")
	}

	messages := []OpenAIChatMessage{
		{Role: "user", Content: userMessage},
	}

	// Add system prompt if provided
	if systemPrompt != "" {
		messages = []OpenAIChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		}
	}

	request := OpenAIChatRequest{
		Model:       p.config.ModelName,
		Messages:    messages,
		Temperature: temperature,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", p.config.BaseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to OpenAI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from OpenAI API: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response OpenAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return response.Choices[0].Message.Content, nil
}

// ListModels returns a list of available models from OpenAI
func (p *OpenAIProvider) ListModels() (string, error) {
	if p.config.APIKey == "" {
		return "", fmt.Errorf("OpenAI API key is required")
	}

	req, err := http.NewRequest("GET", p.config.BaseURL+"/models", nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to OpenAI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from OpenAI API: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response OpenAIListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	// Format the output
	var buf strings.Builder
	buf.WriteString("Available OpenAI Models:\n")

	for _, model := range response.Data {
		// Only include GPT models
		if strings.Contains(model.ID, "gpt") {
			buf.WriteString(fmt.Sprintf("- %s\n", model.ID))
		}
	}

	buf.WriteString("\nNote: Only GPT models are shown. For a complete list, visit the OpenAI documentation.\n")

	return buf.String(), nil
}

// GetName returns the name of the provider
func (p *OpenAIProvider) GetName() string {
	return "openai"
}

// GetModelName returns the name of the currently used model
func (p *OpenAIProvider) GetModelName() string {
	return p.config.ModelName
}

// SetModelName sets the model to use
func (p *OpenAIProvider) SetModelName(modelName string) {
	p.config.ModelName = modelName
}

// RequiresAPIKey returns true if the provider requires an API key
func (p *OpenAIProvider) RequiresAPIKey() bool {
	return true
}
