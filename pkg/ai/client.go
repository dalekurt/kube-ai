package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OllamaClient represents a client for the Ollama API
type OllamaClient struct {
	BaseURL string
	Client  *http.Client
}

// OllamaRequest represents a request to the Ollama API
type OllamaRequest struct {
	Model    string    `json:"model"`
	Prompt   string    `json:"prompt"`
	Stream   bool      `json:"stream,omitempty"`
	Options  Options   `json:"options,omitempty"`
	Messages []Message `json:"messages,omitempty"`
}

// Options represents options for the Ollama model
type Options struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
}

// Message represents a message in a conversation
type Message struct {
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

// NewOllamaClient creates a new client for the Ollama API
func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

// GenerateResponse generates a response from the Ollama API
func (c *OllamaClient) GenerateResponse(model, prompt string, options Options) (string, error) {
	request := OllamaRequest{
		Model:   model,
		Prompt:  prompt,
		Stream:  false,
		Options: options,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	fmt.Printf("Debug - Sending request to: %s\n", c.BaseURL+"/api/generate")
	fmt.Printf("Debug - Request body: %s\n", string(requestBody))

	resp, err := c.Client.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewBuffer(requestBody))
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

// ChatCompletion sends a chat completion request to Ollama
func (c *OllamaClient) ChatCompletion(model string, messages []Message, options Options) (string, error) {
	request := OllamaRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Options:  options,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	fmt.Printf("Debug - Sending request to: %s\n", c.BaseURL+"/api/chat")
	fmt.Printf("Debug - Request body: %s\n", string(requestBody))

	resp, err := c.Client.Post(c.BaseURL+"/api/chat", "application/json", bytes.NewBuffer(requestBody))
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
	// The last one should contain the final response
	lines := strings.Split(string(bodyBytes), "\n")

	// Find the last non-empty line
	var lastLine string
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] != "" {
			lastLine = lines[i]
			break
		}
	}

	if lastLine == "" {
		return "", fmt.Errorf("no response data found")
	}

	var finalResponse OllamaResponse
	if err := json.Unmarshal([]byte(lastLine), &finalResponse); err != nil {
		return "", fmt.Errorf("error decoding final response: %w", err)
	}

	// Alternatively, we can concatenate all the responses
	var fullResponse strings.Builder
	for _, line := range lines {
		if line == "" {
			continue
		}

		var response OllamaResponse
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			continue // Skip lines that don't parse
		}

		// Extract content from the message field if it exists
		if response.Response != "" {
			fullResponse.WriteString(response.Response)
		} else {
			// For the new Ollama API format that uses message.content
			type ResponseMessage struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}

			type NewResponse struct {
				Message ResponseMessage `json:"message"`
			}

			var newResp NewResponse
			if err := json.Unmarshal([]byte(line), &newResp); err == nil {
				if newResp.Message.Content != "" {
					fullResponse.WriteString(newResp.Message.Content)
				}
			}
		}
	}

	return fullResponse.String(), nil
}
