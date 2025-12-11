package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// AnthropicClient implements the Provider interface for Anthropic's Claude API.
type AnthropicClient struct {
	providerID string
	apiKey     string
	baseURL    string
	model      string
	maxTokens  int
	httpClient *http.Client
}

// AnthropicConfig contains configuration for Anthropic client.
type AnthropicConfig struct {
	ProviderID string
	APIKey     string
	BaseURL    string
	Model      string
	MaxTokens  int
}

// NewAnthropicClient creates a new Anthropic client.
func NewAnthropicClient(cfg AnthropicConfig) *AnthropicClient {
	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}

	model := cfg.Model
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	return &AnthropicClient{
		providerID: cfg.ProviderID,
		apiKey:     cfg.APIKey,
		baseURL:    baseURL,
		model:      model,
		maxTokens:  maxTokens,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *AnthropicClient) Name() string      { return c.providerID }
func (c *AnthropicClient) ModelInfo() ModelInfo {
	return ModelInfo{Provider: c.providerID, Model: c.model, MaxTokens: c.maxTokens}
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
	Stream    bool               `json:"stream,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (c *AnthropicClient) TestConnection(ctx context.Context) error {
	reqBody := anthropicRequest{
		Model: c.model, MaxTokens: 10,
		Messages: []anthropicMessage{{Role: "user", Content: "Hi"}},
	}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return NewAPIError("anthropic", 0, "無法建立請求", err)
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewAPIError("anthropic", 0, ErrNetworkError.Error(), ErrNetworkError)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return NewAPIError("anthropic", resp.StatusCode, ErrInvalidAPIKey.Error(), ErrInvalidAPIKey)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return NewAPIError("anthropic", resp.StatusCode, string(body), nil)
	}
	return nil
}

func (c *AnthropicClient) SendMessage(ctx context.Context, messages []Message) (*Response, error) {
	var systemPrompt string
	var anthropicMessages []anthropicMessage
	for _, msg := range messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			anthropicMessages = append(anthropicMessages, anthropicMessage{Role: msg.Role, Content: msg.Content})
		}
	}
	reqBody := anthropicRequest{Model: c.model, MaxTokens: c.maxTokens, Messages: anthropicMessages, System: systemPrompt}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, NewAPIError("anthropic", 0, "無法建立請求", err)
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewAPIError("anthropic", 0, ErrNetworkError.Error(), ErrNetworkError)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, NewAPIError("anthropic", resp.StatusCode, ErrInvalidAPIKey.Error(), ErrInvalidAPIKey)
	}
	if resp.StatusCode == 429 {
		return nil, NewAPIError("anthropic", resp.StatusCode, ErrRateLimited.Error(), ErrRateLimited)
	}
	if resp.StatusCode >= 400 {
		return nil, NewAPIError("anthropic", resp.StatusCode, string(body), nil)
	}
	var anthropicResp anthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, NewAPIError("anthropic", 0, "無法解析回應", err)
	}
	if len(anthropicResp.Content) == 0 {
		return nil, NewAPIError("anthropic", 0, ErrEmptyResponse.Error(), ErrEmptyResponse)
	}
	var content strings.Builder
	for _, cnt := range anthropicResp.Content {
		if cnt.Type == "text" {
			content.WriteString(cnt.Text)
		}
	}
	return &Response{
		Content: content.String(),
		Metadata: map[string]interface{}{
			"model": anthropicResp.Model, "input_tokens": anthropicResp.Usage.InputTokens,
			"output_tokens": anthropicResp.Usage.OutputTokens, "stop_reason": anthropicResp.StopReason,
		},
	}, nil
}

func (c *AnthropicClient) Stream(ctx context.Context, messages []Message, callback func(chunk string)) error {
	var systemPrompt string
	var anthropicMessages []anthropicMessage
	for _, msg := range messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			anthropicMessages = append(anthropicMessages, anthropicMessage{Role: msg.Role, Content: msg.Content})
		}
	}
	reqBody := anthropicRequest{Model: c.model, MaxTokens: c.maxTokens, Messages: anthropicMessages, System: systemPrompt, Stream: true}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return NewAPIError("anthropic", 0, "無法建立請求", err)
	}
	c.setHeaders(req)
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewAPIError("anthropic", 0, ErrNetworkError.Error(), ErrNetworkError)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return NewAPIError("anthropic", resp.StatusCode, ErrInvalidAPIKey.Error(), ErrInvalidAPIKey)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return NewAPIError("anthropic", resp.StatusCode, string(body), nil)
	}
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return NewAPIError("anthropic", 0, "串流讀取錯誤", err)
		}
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			callback(event.Delta.Text)
		}
		if event.Type == "message_stop" {
			break
		}
	}
	return nil
}

func (c *AnthropicClient) setHeaders(req *http.Request) {
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")
}
