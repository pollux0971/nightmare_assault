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

// GoogleClient implements the Provider interface for Google's Generative AI API.
type GoogleClient struct {
	providerID string
	apiKey     string
	baseURL    string
	model      string
	maxTokens  int
	httpClient *http.Client
}

// GoogleConfig contains configuration for Google client.
type GoogleConfig struct {
	ProviderID string
	APIKey     string
	BaseURL    string
	Model      string
	MaxTokens  int
}

// NewGoogleClient creates a new Google Generative AI client.
func NewGoogleClient(cfg GoogleConfig) *GoogleClient {
	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	model := cfg.Model
	if model == "" {
		model = "gemini-1.5-pro"
	}
	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}
	return &GoogleClient{
		providerID: cfg.ProviderID, apiKey: cfg.APIKey, baseURL: baseURL,
		model: model, maxTokens: maxTokens,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *GoogleClient) Name() string      { return c.providerID }
func (c *GoogleClient) ModelInfo() ModelInfo {
	return ModelInfo{Provider: c.providerID, Model: c.model, MaxTokens: c.maxTokens}
}

type googleRequest struct {
	Contents          []googleContent         `json:"contents"`
	SystemInstruction *googleContent          `json:"systemInstruction,omitempty"`
	GenerationConfig  *googleGenerationConfig `json:"generationConfig,omitempty"`
}

type googleContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []googlePart `json:"parts"`
}

type googlePart struct {
	Text string `json:"text"`
}

type googleGenerationConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
}

type googleResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct{ Text string `json:"text"` } `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func (c *GoogleClient) TestConnection(ctx context.Context) error {
	url := fmt.Sprintf("%s/models?key=%s", c.baseURL, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return NewAPIError("google", 0, "無法建立請求", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewAPIError("google", 0, ErrNetworkError.Error(), ErrNetworkError)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return NewAPIError("google", resp.StatusCode, ErrInvalidAPIKey.Error(), ErrInvalidAPIKey)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return NewAPIError("google", resp.StatusCode, string(body), nil)
	}
	return nil
}

func (c *GoogleClient) SendMessage(ctx context.Context, messages []Message) (*Response, error) {
	var systemInstruction *googleContent
	var contents []googleContent
	for _, msg := range messages {
		if msg.Role == "system" {
			systemInstruction = &googleContent{Parts: []googlePart{{Text: msg.Content}}}
		} else {
			role := msg.Role
			if role == "assistant" {
				role = "model"
			}
			contents = append(contents, googleContent{Role: role, Parts: []googlePart{{Text: msg.Content}}})
		}
	}
	reqBody := googleRequest{
		Contents: contents, SystemInstruction: systemInstruction,
		GenerationConfig: &googleGenerationConfig{MaxOutputTokens: c.maxTokens},
	}
	jsonBody, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, NewAPIError("google", 0, "無法建立請求", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewAPIError("google", 0, ErrNetworkError.Error(), ErrNetworkError)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, NewAPIError("google", resp.StatusCode, ErrInvalidAPIKey.Error(), ErrInvalidAPIKey)
	}
	if resp.StatusCode >= 400 {
		return nil, NewAPIError("google", resp.StatusCode, string(body), nil)
	}
	var googleResp googleResponse
	if err := json.Unmarshal(body, &googleResp); err != nil {
		return nil, NewAPIError("google", 0, "無法解析回應", err)
	}
	if len(googleResp.Candidates) == 0 || len(googleResp.Candidates[0].Content.Parts) == 0 {
		return nil, NewAPIError("google", 0, ErrEmptyResponse.Error(), ErrEmptyResponse)
	}
	var content strings.Builder
	for _, part := range googleResp.Candidates[0].Content.Parts {
		content.WriteString(part.Text)
	}
	return &Response{
		Content: content.String(),
		Metadata: map[string]interface{}{
			"model": c.model, "prompt_tokens": googleResp.UsageMetadata.PromptTokenCount,
			"output_tokens": googleResp.UsageMetadata.CandidatesTokenCount,
		},
	}, nil
}

func (c *GoogleClient) Stream(ctx context.Context, messages []Message, callback func(chunk string)) error {
	var systemInstruction *googleContent
	var contents []googleContent
	for _, msg := range messages {
		if msg.Role == "system" {
			systemInstruction = &googleContent{Parts: []googlePart{{Text: msg.Content}}}
		} else {
			role := msg.Role
			if role == "assistant" {
				role = "model"
			}
			contents = append(contents, googleContent{Role: role, Parts: []googlePart{{Text: msg.Content}}})
		}
	}
	reqBody := googleRequest{
		Contents: contents, SystemInstruction: systemInstruction,
		GenerationConfig: &googleGenerationConfig{MaxOutputTokens: c.maxTokens},
	}
	jsonBody, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s&alt=sse", c.baseURL, c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return NewAPIError("google", 0, "無法建立請求", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewAPIError("google", 0, ErrNetworkError.Error(), ErrNetworkError)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return NewAPIError("google", resp.StatusCode, string(body), nil)
	}
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return NewAPIError("google", 0, "串流讀取錯誤", err)
		}
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var chunk googleResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Candidates) > 0 && len(chunk.Candidates[0].Content.Parts) > 0 {
			for _, part := range chunk.Candidates[0].Content.Parts {
				if part.Text != "" {
					callback(part.Text)
				}
			}
		}
	}
	return nil
}
