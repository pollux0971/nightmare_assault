package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIClient implements the Provider interface for OpenAI-compatible APIs.
type OpenAIClient struct {
	providerID string
	apiKey     string
	baseURL    string
	model      string
	maxTokens  int
	httpClient *http.Client
}

// OpenAIConfig contains configuration for OpenAI-compatible clients.
type OpenAIConfig struct {
	ProviderID string
	APIKey     string
	BaseURL    string
	Model      string
	MaxTokens  int
}

// NewOpenAIClient creates a new OpenAI-compatible client.
func NewOpenAIClient(cfg OpenAIConfig) *OpenAIClient {
	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := cfg.Model
	if model == "" {
		model = "gpt-4o"
	}

	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	return &OpenAIClient{
		providerID: cfg.ProviderID,
		apiKey:     cfg.APIKey,
		baseURL:    baseURL,
		model:      model,
		maxTokens:  maxTokens,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Name returns the provider name.
func (c *OpenAIClient) Name() string {
	return c.providerID
}

// ModelInfo returns model information.
func (c *OpenAIClient) ModelInfo() ModelInfo {
	return ModelInfo{
		Provider:  c.providerID,
		Model:     c.model,
		MaxTokens: c.maxTokens,
	}
}

// TestConnection tests the API connection.
func (c *OpenAIClient) TestConnection(ctx context.Context) error {
	// Try to list models as a simple connection test
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/models", nil)
	if err != nil {
		return NewAPIError(c.providerID, 0, "無法建立請求", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewAPIError(c.providerID, 0, ErrNetworkError.Error(), ErrNetworkError)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return NewAPIError(c.providerID, resp.StatusCode, ErrInvalidAPIKey.Error(), ErrInvalidAPIKey)
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return NewAPIError(c.providerID, resp.StatusCode, string(body), nil)
	}

	return nil
}

// openAIChatRequest represents the request body for chat completions.
type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float64             `json:"temperature,omitempty"`
	Stream      bool                `json:"stream,omitempty"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIChatResponse represents the response from chat completions.
type openAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// SendMessage sends a message and returns the response.
func (c *OpenAIClient) SendMessage(ctx context.Context, messages []Message) (*Response, error) {
	// Convert messages to OpenAI format
	openAIMessages := make([]openAIChatMessage, len(messages))
	for i, msg := range messages {
		openAIMessages[i] = openAIChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	reqBody := openAIChatRequest{
		Model:     c.model,
		Messages:  openAIMessages,
		MaxTokens: c.maxTokens,
		Stream:    false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, NewAPIError(c.providerID, 0, "無法序列化請求", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, NewAPIError(c.providerID, 0, "無法建立請求", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewAPIError(c.providerID, 0, ErrNetworkError.Error(), ErrNetworkError)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewAPIError(c.providerID, 0, "無法讀取回應", err)
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, NewAPIError(c.providerID, resp.StatusCode, ErrInvalidAPIKey.Error(), ErrInvalidAPIKey)
	}

	if resp.StatusCode == 429 {
		return nil, NewAPIError(c.providerID, resp.StatusCode, ErrRateLimited.Error(), ErrRateLimited)
	}

	if resp.StatusCode >= 400 {
		return nil, NewAPIError(c.providerID, resp.StatusCode, string(body), nil)
	}

	var chatResp openAIChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, NewAPIError(c.providerID, 0, "無法解析回應", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, NewAPIError(c.providerID, 0, ErrEmptyResponse.Error(), ErrEmptyResponse)
	}

	return &Response{
		Content: chatResp.Choices[0].Message.Content,
		Metadata: map[string]interface{}{
			"model":             chatResp.Model,
			"prompt_tokens":     chatResp.Usage.PromptTokens,
			"completion_tokens": chatResp.Usage.CompletionTokens,
			"total_tokens":      chatResp.Usage.TotalTokens,
		},
	}, nil
}

// Stream sends a message and streams the response via callback.
func (c *OpenAIClient) Stream(ctx context.Context, messages []Message, callback func(chunk string)) error {
	// Convert messages to OpenAI format
	openAIMessages := make([]openAIChatMessage, len(messages))
	for i, msg := range messages {
		openAIMessages[i] = openAIChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	reqBody := openAIChatRequest{
		Model:     c.model,
		Messages:  openAIMessages,
		MaxTokens: c.maxTokens,
		Stream:    true,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return NewAPIError(c.providerID, 0, "無法序列化請求", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return NewAPIError(c.providerID, 0, "無法建立請求", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewAPIError(c.providerID, 0, ErrNetworkError.Error(), ErrNetworkError)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return NewAPIError(c.providerID, resp.StatusCode, ErrInvalidAPIKey.Error(), ErrInvalidAPIKey)
	}

	if resp.StatusCode == 429 {
		return NewAPIError(c.providerID, resp.StatusCode, ErrRateLimited.Error(), ErrRateLimited)
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return NewAPIError(c.providerID, resp.StatusCode, string(body), nil)
	}

	// Parse SSE stream
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return NewAPIError(c.providerID, 0, "串流讀取錯誤", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue // Skip malformed chunks
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			callback(chunk.Choices[0].Delta.Content)
		}
	}

	return nil
}

// setHeaders sets common headers for API requests.
func (c *OpenAIClient) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Add OpenRouter-specific headers if applicable
	if c.providerID == "openrouter" {
		req.Header.Set("HTTP-Referer", "https://github.com/nightmare-assault/nightmare-assault")
		req.Header.Set("X-Title", "Nightmare Assault")
	}
}
