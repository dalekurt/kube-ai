package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GeminiProvider implements the Provider interface for Google's Gemini
type GeminiProvider struct {
	config ProviderConfig
	client *http.Client
}

// GeminiRequest represents a request to the Gemini API
type GeminiRequest struct {
	Contents         []GeminiContent        `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig"`
}

// GeminiContent represents a content part in a Gemini request
type GeminiContent struct {
	Role  string `json:"role,omitempty"`
	Parts []struct {
		Text string `json:"text"`
	} `json:"parts"`
}

// GeminiGenerationConfig represents the generation config for Gemini
type GeminiGenerationConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

// GeminiResponse represents a response from the Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Role  string `json:"role"`
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string, modelName string) *GeminiProvider {
	if modelName == "" {
		modelName = "gemini-1.5-pro"
	}

	return &GeminiProvider{
		config: ProviderConfig{
			BaseURL:   "https://generativelanguage.googleapis.com/v1beta",
			APIKey:    apiKey,
			ModelName: modelName,
		},
		client: &http.Client{},
	}
}

// GenerateResponse generates a response for a prompt
func (p *GeminiProvider) GenerateResponse(prompt string, temperature float64) (string, error) {
	if p.config.APIKey == "" {
		return "", fmt.Errorf("Gemini API key is required")
	}

	content := GeminiContent{
		Parts: []struct {
			Text string `json:"text"`
		}{
			{Text: prompt},
		},
	}

	request := GeminiRequest{
		Contents: []GeminiContent{content},
		GenerationConfig: GeminiGenerationConfig{
			Temperature:     temperature,
			MaxOutputTokens: 4096,
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.config.BaseURL, p.config.ModelName, p.config.APIKey)
	resp, err := p.client.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error making request to Gemini: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from Gemini API: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response text returned")
	}

	return response.Candidates[0].Content.Parts[0].Text, nil
}

// ChatCompletion generates a response from a conversation
func (p *GeminiProvider) ChatCompletion(systemPrompt string, userMessage string, temperature float64) (string, error) {
	if p.config.APIKey == "" {
		return "", fmt.Errorf("Gemini API key is required")
	}

	// For Gemini, we'll merge system prompt and user message if both provided
	var messageText string
	if systemPrompt != "" {
		messageText = fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, userMessage)
	} else {
		messageText = userMessage
	}

	content := GeminiContent{
		Role: "user",
		Parts: []struct {
			Text string `json:"text"`
		}{
			{Text: messageText},
		},
	}

	request := GeminiRequest{
		Contents: []GeminiContent{content},
		GenerationConfig: GeminiGenerationConfig{
			Temperature:     temperature,
			MaxOutputTokens: 4096,
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.config.BaseURL, p.config.ModelName, p.config.APIKey)
	resp, err := p.client.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error making request to Gemini: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from Gemini API: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response text returned")
	}

	return response.Candidates[0].Content.Parts[0].Text, nil
}

// ListModels returns a list of available models from Gemini
func (p *GeminiProvider) ListModels() (string, error) {
	var buf strings.Builder
	buf.WriteString("Available Gemini Models:\n")
	buf.WriteString("- gemini-1.5-pro\n")
	buf.WriteString("- gemini-1.5-flash\n")
	buf.WriteString("- gemini-pro\n")
	buf.WriteString("- gemini-pro-vision\n")

	return buf.String(), nil
}

// GetName returns the name of the provider
func (p *GeminiProvider) GetName() string {
	return "gemini"
}

// GetModelName returns the name of the currently used model
func (p *GeminiProvider) GetModelName() string {
	return p.config.ModelName
}

// SetModelName sets the model to use
func (p *GeminiProvider) SetModelName(modelName string) {
	p.config.ModelName = modelName
}

// RequiresAPIKey returns true if the provider requires an API key
func (p *GeminiProvider) RequiresAPIKey() bool {
	return true
}
