package ifm

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientAuthenticationAndSessionRouting(t *testing.T) {
	expectedToken := "test-secret-ifm-key"
	expectedSession := "session-uuid-123"

	mockTransport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		// Verify Auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+expectedToken {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"Invalid API Key"}}`)),
				Header:     make(http.Header),
			}, nil
		}

		// Verify Session Routing header
		sessionHeader := r.Header.Get("X-Session-ID")
		if sessionHeader != expectedSession {
			t.Errorf("expected X-Session-ID '%s', got '%s'", expectedSession, sessionHeader)
		}

		respBody := `{
			"id": "chatcmpl-test",
			"object": "chat.completion",
			"created": 12345678,
			"model": "IFM/K2-Horizon-375B-A23B",
			"choices": [
				{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "Hello human!"
					},
					"finish_reason": "stop"
				}
			]
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(respBody)),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient("https://api.ifm.ai/v1", expectedToken, 5*time.Second)
	client.httpClient.Transport = mockTransport

	// Test Successful Chat Completion
	req := ChatCompletionRequest{
		Model: "IFM/K2-Horizon-375B-A23B",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.CreateChatCompletion(context.Background(), req, expectedSession)
	if err != nil {
		t.Fatalf("CreateChatCompletion failed: %v", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content != "Hello human!" {
		t.Fatalf("unexpected response content: %+v", resp)
	}

	// Test Unauthorized with wrong key
	badClient := NewClient("https://api.ifm.ai/v1", "wrong-key", 5*time.Second)
	badClient.httpClient.Transport = mockTransport
	_, err = badClient.CreateChatCompletion(context.Background(), req, expectedSession)
	if err == nil {
		t.Errorf("expected error for unauthorized request, got nil")
	}
}

func TestStreamChatCompletionAndReasoningParsing(t *testing.T) {
	ssePayload := strings.Join([]string{
		"data: {\"choices\":[{\"index\":0,\"delta\":{\"reasoning_content\":\"Thinking step 1. \"}}]}",
		"data: {\"choices\":[{\"index\":0,\"delta\":{\"reasoning_content\":\"Step 2 done.\"}}]}",
		"data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Final answer here.\"}}]}",
		"data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}",
		"data: [DONE]",
		"",
	}, "\n\n")

	mockTransport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		header := make(http.Header)
		header.Set("Content-Type", "text/event-stream")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(ssePayload)),
			Header:     header,
		}, nil
	})

	client := NewClient("https://api.ifm.ai/v1", "test-key", 5*time.Second)
	client.httpClient.Transport = mockTransport

	req := ChatCompletionRequest{
		Model: "IFM/K2-Horizon-375B-A23B",
		Messages: []ChatMessage{
			{Role: "user", Content: "Calculate 2+2"},
		},
	}

	var accumulatedContent string
	var accumulatedReasoning string
	var finalFinishReason string

	err := client.StreamChatCompletion(context.Background(), req, "sess-test", func(chunk *StreamChunk) error {
		for _, c := range chunk.Choices {
			if c.Delta.Content != "" {
				accumulatedContent += c.Delta.Content
			}
			if c.Delta.ReasoningContent != nil {
				accumulatedReasoning += *c.Delta.ReasoningContent
			}
			if c.FinishReason != nil {
				finalFinishReason = *c.FinishReason
			}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("StreamChatCompletion failed: %v", err)
	}

	expectedReasoning := "Thinking step 1. Step 2 done."
	expectedContent := "Final answer here."

	if accumulatedReasoning != expectedReasoning {
		t.Errorf("expected reasoning '%s', got '%s'", expectedReasoning, accumulatedReasoning)
	}
	if accumulatedContent != expectedContent {
		t.Errorf("expected content '%s', got '%s'", expectedContent, accumulatedContent)
	}
	if finalFinishReason != "stop" {
		t.Errorf("expected finish_reason 'stop', got '%s'", finalFinishReason)
	}
}

type blockingReader struct {
	ctx context.Context
}

func (b *blockingReader) Read(p []byte) (n int, err error) {
	<-b.ctx.Done()
	return 0, b.ctx.Err()
}

func (b *blockingReader) Close() error {
	return nil
}

func TestStreamChatCompletionCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mockTransport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		header := make(http.Header)
		header.Set("Content-Type", "text/event-stream")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       &blockingReader{ctx: ctx},
			Header:     header,
		}, nil
	})

	client := NewClient("https://api.ifm.ai/v1", "test-key", 5*time.Second)
	client.httpClient.Transport = mockTransport

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel() // Cancel context during active stream reading
	}()

	err := client.StreamChatCompletion(ctx, ChatCompletionRequest{
		Model: "IFM/K2-Horizon-375B-A23B",
	}, "sess-cancel", func(chunk *StreamChunk) error {
		return nil
	})

	if err == nil {
		t.Errorf("expected context cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected 'context canceled' error, got: %v", err)
	}
}
