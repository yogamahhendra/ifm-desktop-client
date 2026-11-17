package repositories

import "time"

type Conversation struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	ModelID      string    `json:"model_id"`
	PresetID     *string   `json:"preset_id"`
	SessionID    string    `json:"session_id"`
	SystemPrompt string    `json:"system_prompt"`
	FolderPath   *string   `json:"folder_path"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Message struct {
	ID               string    `json:"id"`
	ConversationID   string    `json:"conversation_id"`
	ParentMessageID  *string   `json:"parent_message_id"`
	Role             string    `json:"role"`
	Content          string    `json:"content"`
	ReasoningContent *string   `json:"reasoning_content"`
	ToolCalls        *string   `json:"tool_calls"`
	ToolCallID       *string   `json:"tool_call_id"`
	FinishReason     *string   `json:"finish_reason"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	CreatedAt        time.Time `json:"created_at"`
}

type ModelItem struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	ModelIdentifier string    `json:"model_identifier"`
	Description     string    `json:"description"`
	Temperature     float64   `json:"temperature"`
	TopP            float64   `json:"top_p"`
	TopK            *int      `json:"top_k"`
	MaxTokens       int       `json:"max_tokens"`
	ReasoningEffort string    `json:"reasoning_effort"`
	IsDefault       bool      `json:"is_default"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PresetItem struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	SystemPrompt    string    `json:"system_prompt"`
	ModelID         *string   `json:"model_id"`
	Temperature     *float64  `json:"temperature"`
	TopP            *float64  `json:"top_p"`
	MaxTokens       *int      `json:"max_tokens"`
	ReasoningEffort *string   `json:"reasoning_effort"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AppSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}
