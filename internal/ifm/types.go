package ifm

// ChatMessage represents a single message in the conversation payload.
type ChatMessage struct {
	Role             string     `json:"role"`
	Content          string     `json:"content"`
	ReasoningContent *string    `json:"reasoning_content,omitempty"`
	Think            *string    `json:"think,omitempty"`
	Thinking         *string    `json:"thinking,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID       *string    `json:"tool_call_id,omitempty"`
	Name             *string    `json:"name,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatCompletionRequest is the payload sent to POST /chat/completions.
type ChatCompletionRequest struct {
	Model           string        `json:"model"`
	Messages        []ChatMessage `json:"messages"`
	Temperature     *float64      `json:"temperature,omitempty"`
	TopP            *float64      `json:"top_p,omitempty"`
	TopK            *int          `json:"top_k,omitempty"`
	MaxTokens       *int          `json:"max_tokens,omitempty"`
	Stop            []string      `json:"stop,omitempty"`
	Seed            *int          `json:"seed,omitempty"`
	Stream          bool          `json:"stream"`
	ReasoningEffort string        `json:"reasoning_effort,omitempty"`
	Tools           any           `json:"tools,omitempty"`
	ToolChoice      any           `json:"tool_choice,omitempty"`
	ResponseFormat  any           `json:"response_format,omitempty"`
}

// ChatCompletionResponse is the non-streaming response format.
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
}

type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason *string     `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents SSE data chunks from /chat/completions.
type StreamChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
	Usage   *Usage         `json:"usage,omitempty"`
}

type StreamChoice struct {
	Index        int         `json:"index"`
	Delta        StreamDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

type StreamDelta struct {
	Role             string     `json:"role,omitempty"`
	Content          string     `json:"content,omitempty"`
	ReasoningContent *string    `json:"reasoning_content,omitempty"`
	Think            *string    `json:"think,omitempty"`
	Thinking         *string    `json:"thinking,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
}

// StreamEvent represents the event emitted to the frontend during streaming.
type StreamEvent struct {
	ConversationID   string  `json:"conversation_id"`
	MessageID        string  `json:"message_id"`
	ContentDelta     string  `json:"content_delta,omitempty"`
	ReasoningDelta   string  `json:"reasoning_delta,omitempty"`
	FullContent      string  `json:"full_content"`
	FullReasoning    string  `json:"full_reasoning"`
	FinishReason     *string `json:"finish_reason,omitempty"`
	IsDone           bool    `json:"is_done"`
	Error            string  `json:"error,omitempty"`
	PromptTokens     int     `json:"prompt_tokens,omitempty"`
	CompletionTokens int     `json:"completion_tokens,omitempty"`
	SearchSource     string  `json:"search_source,omitempty"`
	SearchCredits    float64 `json:"search_credits,omitempty"`
}
