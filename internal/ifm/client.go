package ifm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type APIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    any    `json:"code"`
	} `json:"error"`
}

func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.ifm.ai/v1"
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = strings.TrimRight(baseURL, "/")
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// TestConnection tests whether the API endpoint and credentials are valid.
func (c *Client) TestConnection(ctx context.Context) error {
	if c.apiKey == "" {
		return errors.New("API key is not configured")
	}

	// Test with a minimal chat completion request (low reasoning effort for fast ping)
	reqBody := ChatCompletionRequest{
		Model: "IFM/K2-Horizon-375B-A23B",
		Messages: []ChatMessage{
			{Role: "user", Content: "ping"},
		},
		MaxTokens:       func(v int) *int { return &v }(1),
		ReasoningEffort: "low",
		Stream:          false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return errors.New("invalid API key or unauthorized")
	}

	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		var apiErr APIErrorResponse
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.Error.Message)
		}
		return fmt.Errorf("API returned HTTP status %d", resp.StatusCode)
	}

	return nil
}

// CreateChatCompletion sends a non-streaming chat completion request.
func (c *Client) CreateChatCompletion(ctx context.Context, reqBody ChatCompletionRequest, sessionID string) (*ChatCompletionResponse, error) {
	reqBody.Stream = false
	c.sanitizeModelParams(&reqBody)

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed encoding request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.Header.Set("X-Session-ID", sessionID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var apiErr APIErrorResponse
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.Error.Message)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed decoding response: %w", err)
	}

	return &result, nil
}

// StreamChatCompletion sends a streaming request and reads SSE chunks.
func (c *Client) StreamChatCompletion(ctx context.Context, reqBody ChatCompletionRequest, sessionID string, onChunk func(chunk *StreamChunk) error) error {
	reqBody.Stream = true
	c.sanitizeModelParams(&reqBody)

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed encoding stream request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed creating stream request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if sessionID != "" {
		req.Header.Set("X-Session-ID", sessionID)
	}

	// Use transport directly with no client timeout so streaming can take as long as needed,
	// controlled strictly by ctx
	streamClient := &http.Client{
		Transport: c.httpClient.Transport,
		Timeout:   0,
	}

	resp, err := streamClient.Do(req)
	if err != nil {
		return fmt.Errorf("stream connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var apiErr APIErrorResponse
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.Error.Message)
		}
		return fmt.Errorf("stream request failed with status %d: %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("reading SSE stream failed: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ":") {
			continue // comment or empty heartbeat
		}

		if strings.HasPrefix(line, "data: ") {
			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				break
			}

			var chunk StreamChunk
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				// Continue on partial formatting glitches
				continue
			}

			if err := onChunk(&chunk); err != nil {
				return err
			}
		}
	}

	return nil
}

// sanitizeModelParams ensures model-specific parameter handling:
// - Omits reasoning_effort for models that do not support reasoning or when not set.
// - Sets recommended defaults if parameters are omitted.
func (c *Client) sanitizeModelParams(req *ChatCompletionRequest) {
	if strings.Contains(req.Model, "K2-Horizon") {
		if req.Temperature == nil {
			t := 1.0
			req.Temperature = &t
		}
		if req.TopP == nil {
			p := 0.95
			req.TopP = &p
		}
		if req.ReasoningEffort == "" {
			req.ReasoningEffort = "high"
		}
	} else {
		// Non-reasoning models: avoid sending reasoning_effort if not supported
		if req.ReasoningEffort == "" {
			req.ReasoningEffort = ""
		}
	}
}
