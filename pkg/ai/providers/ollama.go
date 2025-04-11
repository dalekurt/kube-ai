package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	config ProviderConfig
	client *http.Client
}

// OllamaRequest represents a request to the Ollama API
type OllamaRequest struct {
	Model    string          `json:"model"`
	Prompt   string          `json:"prompt"`
	Stream   bool            `json:"stream,omitempty"`
	Options  OllamaOptions   `json:"options,omitempty"`
	Messages []OllamaMessage `json:"messages,omitempty"`
}

// OllamaOptions represents options for the Ollama model
type OllamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
}

// OllamaMessage represents a message in a conversation
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaResponse represents a response from the Ollama API
type OllamaResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	CreatedAt string `json:"created_at"`
	Done      bool   `json:"done"`
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL string, modelName string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	if modelName == "" {
		modelName = "llama3.3"
	}

	return &OllamaProvider{
		config: ProviderConfig{
			BaseURL:   baseURL,
			ModelName: modelName,
		},
		client: &http.Client{},
	}
}

// GenerateResponse generates a response for a prompt
func (p *OllamaProvider) GenerateResponse(prompt string, temperature float64) (string, error) {
	request := OllamaRequest{
		Model:  p.config.ModelName,
		Prompt: prompt,
		Stream: false,
		Options: OllamaOptions{
			Temperature: temperature,
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	fmt.Printf("Debug - Sending request to: %s\n", p.config.BaseURL+"/api/generate")
	fmt.Printf("Debug - Request body: %s\n", string(requestBody))

	resp, err := p.client.Post(p.config.BaseURL+"/api/generate", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error making request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from Ollama API: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read the full response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// The response is a series of JSON objects, one per line
	lines := strings.Split(string(bodyBytes), "\n")

	// Concatenate all the responses
	var fullResponse strings.Builder
	for _, line := range lines {
		if line == "" {
			continue
		}

		var response OllamaResponse
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			continue // Skip lines that don't parse
		}

		if response.Response != "" {
			fullResponse.WriteString(response.Response)
		}
	}

	return fullResponse.String(), nil
}

// ChatCompletion sends a chat message to the Ollama API
func (p *OllamaProvider) ChatCompletion(systemPrompt string, userMessage string, temperature float32) (string, error) {
	messages := []OllamaMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	request := OllamaRequest{
		Model:    p.config.ModelName,
		Messages: messages,
		Stream:   false,
		Options: OllamaOptions{
			Temperature: float64(temperature),
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	fmt.Printf("Debug - Sending request to: %s\n", p.config.BaseURL+"/api/chat")
	fmt.Printf("Debug - Request body: %s\n", string(requestBody))

	resp, err := p.client.Post(p.config.BaseURL+"/api/chat", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error making request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from Ollama API: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read the full response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// The response is a series of JSON objects, one per line
	lines := strings.Split(string(bodyBytes), "\n")

	// Concatenate all the responses
	var fullResponse strings.Builder
	for _, line := range lines {
		if line == "" {
			continue
		}

		var response OllamaResponse
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			continue // Skip lines that don't parse
		}

		if response.Response != "" {
			fullResponse.WriteString(response.Response)
		}
	}

	return fullResponse.String(), nil
}

// ListModels returns a list of available models from Ollama
func (p *OllamaProvider) ListModels() (string, error) {
	resp, err := p.client.Get(p.config.BaseURL + "/api/tags")
	if err != nil {
		return "", fmt.Errorf("error getting models: %w", err)
	}
	defer resp.Body.Close()

	// Parse the JSON response
	type ModelInfo struct {
		Name       string `json:"name"`
		ModifiedAt string `json:"modified_at"`
		Size       int64  `json:"size"`
	}

	type TagsResponse struct {
		Models []ModelInfo `json:"models"`
	}

	var response TagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	// Format the output
	var buf strings.Builder
	buf.WriteString("Available Models:\n")

	for _, model := range response.Models {
		sizeInGB := float64(model.Size) / (1024 * 1024 * 1024)
		buf.WriteString(fmt.Sprintf("- %s (%.2f GB)\n", model.Name, sizeInGB))
	}

	return buf.String(), nil
}

// GetName returns the name of the provider
func (p *OllamaProvider) GetName() string {
	return "ollama"
}

// GetModelName returns the name of the currently used model
func (p *OllamaProvider) GetModelName() string {
	return p.config.ModelName
}

// SetModelName sets the model to use
func (p *OllamaProvider) SetModelName(modelName string) {
	p.config.ModelName = modelName
}

// RequiresAPIKey returns true if the provider requires an API key
func (p *OllamaProvider) RequiresAPIKey() bool {
	return false
}

// GenerateCompletion sends a prompt to Ollama and returns the response
func (p *OllamaProvider) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	// Create the request body
	requestBody := map[string]interface{}{
		"model":  p.config.ModelName,
		"prompt": prompt,
		"stream": false, // Set to false to get a complete response
		"options": map[string]interface{}{
			"temperature": 0.7,
		},
	}

	// Convert request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating request JSON: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	// Check for error status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from Ollama API: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Ollama may still send multiple JSON objects even with stream:false
	// Read the response line by line to handle this case
	scanner := bufio.NewScanner(resp.Body)
	var responseText strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse each line as a separate JSON object
		var responseObj OllamaResponse
		if err := json.Unmarshal([]byte(line), &responseObj); err != nil {
			return "", fmt.Errorf("error parsing response JSON: %w", err)
		}

		// Append the response text
		responseText.WriteString(responseObj.Response)

		// If done is true, this is the last message
		if responseObj.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	return responseText.String(), nil
}
