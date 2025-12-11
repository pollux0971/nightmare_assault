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

// CohereClient implements the Provider interface for Cohere's API.
type CohereClient struct {
	providerID string
	apiKey     string
	baseURL    string
	model      string
	maxTokens  int
	httpClient *http.Client
}

// CohereConfig contains configuration for Cohere client.
type CohereConfig struct {
	ProviderID string
	APIKey     string
	BaseURL    string
	Model      string
	MaxTokens  int
}

// NewCohereClient creates a new Cohere client.
func NewCohereClient(cfg CohereConfig) *CohereClient {
	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.cohere.ai/v1"
	}
	model := cfg.Model
	if model == "" {
		model = "command-r-plus"
	}
	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}
	return &CohereClient{
		providerID: cfg.ProviderID, apiKey: cfg.APIKey, baseURL: baseURL,
		model: model, maxTokens: maxTokens,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *CohereClient) Name() string      { return c.providerID }
func (c *CohereClient) ModelInfo() ModelInfo {
	return ModelInfo{Provider: c.providerID, Model: c.model, MaxTokens: c.maxTokens}
}

type cohereRequest struct {
	Model       string          `json:"model"`
	Message     string          `json:"message"`
	ChatHistory []cohereMessage `json:"chat_history,omitempty"`
	Preamble    string          `json:"preamble,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type cohereMessage struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

type cohereResponse struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
	Meta         struct {
		Tokens struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"tokens"`
	} `json:"meta"`
}

func (c *CohereClient) TestConnection(ctx context.Context) error {
	reqBody := cohereRequest{Model: c.model, Message: "Hi", MaxTokens: 10}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return NewAPIError("cohere", 0, "無法建立請求", err)
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewAPIError("cohere", 0, ErrNetworkError.Error(), ErrNetworkError)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return NewAPIError("cohere", resp.StatusCode, ErrInvalidAPIKey.Error(), ErrInvalidAPIKey)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return NewAPIError("cohere", resp.StatusCode, string(body), nil)
	}
	return nil
}

func (c *CohereClient) SendMessage(ctx context.Context, messages []Message) (*Response, error) {
	var preamble string
	var chatHistory []cohereMessage
	var lastUserMessage string
	for i, msg := range messages {
		if msg.Role == "system" {
			preamble = msg.Content
		} else if i == len(messages)-1 && msg.Role == "user" {
			lastUserMessage = msg.Content
		} else {
			role := "USER"
			if msg.Role == "assistant" {
				role = "CHATBOT"
			}
			chatHistory = append(chatHistory, cohereMessage{Role: role, Message: msg.Content})
		}
	}
	if lastUserMessage == "" {
		return nil, NewAPIError("cohere", 0, "必須提供使用者訊息", nil)
	}
	reqBody := cohereRequest{Model: c.model, Message: lastUserMessage, ChatHistory: chatHistory, Preamble: preamble, MaxTokens: c.maxTokens}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, NewAPIError("cohere", 0, "無法建立請求", err)
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewAPIError("cohere", 0, ErrNetworkError.Error(), ErrNetworkError)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, NewAPIError("cohere", resp.StatusCode, ErrInvalidAPIKey.Error(), ErrInvalidAPIKey)
	}
	if resp.StatusCode >= 400 {
		return nil, NewAPIError("cohere", resp.StatusCode, string(body), nil)
	}
	var cohereResp cohereResponse
	if err := json.Unmarshal(body, &cohereResp); err != nil {
		return nil, NewAPIError("cohere", 0, "無法解析回應", err)
	}
	if cohereResp.Text == "" {
		return nil, NewAPIError("cohere", 0, ErrEmptyResponse.Error(), ErrEmptyResponse)
	}
	return &Response{
		Content: cohereResp.Text,
		Metadata: map[string]interface{}{
			"model": c.model, "input_tokens": cohereResp.Meta.Tokens.InputTokens,
			"output_tokens": cohereResp.Meta.Tokens.OutputTokens,
		},
	}, nil
}

func (c *CohereClient) Stream(ctx context.Context, messages []Message, callback func(chunk string)) error {
	var preamble string
	var chatHistory []cohereMessage
	var lastUserMessage string
	for i, msg := range messages {
		if msg.Role == "system" {
			preamble = msg.Content
		} else if i == len(messages)-1 && msg.Role == "user" {
			lastUserMessage = msg.Content
		} else {
			role := "USER"
			if msg.Role == "assistant" {
				role = "CHATBOT"
			}
			chatHistory = append(chatHistory, cohereMessage{Role: role, Message: msg.Content})
		}
	}
	if lastUserMessage == "" {
		return NewAPIError("cohere", 0, "必須提供使用者訊息", nil)
	}
	reqBody := cohereRequest{Model: c.model, Message: lastUserMessage, ChatHistory: chatHistory, Preamble: preamble, MaxTokens: c.maxTokens, Stream: true}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return NewAPIError("cohere", 0, "無法建立請求", err)
	}
	c.setHeaders(req)
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewAPIError("cohere", 0, ErrNetworkError.Error(), ErrNetworkError)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return NewAPIError("cohere", resp.StatusCode, string(body), nil)
	}
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return NewAPIError("cohere", 0, "串流讀取錯誤", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var event struct {
			EventType string `json:"event_type"`
			Text      string `json:"text"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		if event.EventType == "text-generation" && event.Text != "" {
			callback(event.Text)
		}
		if event.EventType == "stream-end" {
			break
		}
	}
	return nil
}

func (c *CohereClient) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
}
