package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, m *Message) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO messages (
			id, conversation_id, parent_message_id, role, content,
			reasoning_content, tool_calls, tool_call_id, finish_reason,
			prompt_tokens, completion_tokens, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		m.ID, m.ConversationID, m.ParentMessageID, m.Role, m.Content,
		m.ReasoningContent, m.ToolCalls, m.ToolCallID, m.FinishReason,
		m.PromptTokens, m.CompletionTokens, m.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

func (r *MessageRepository) Get(ctx context.Context, id string) (*Message, error) {
	query := `
		SELECT id, conversation_id, parent_message_id, role, content,
		       reasoning_content, tool_calls, tool_call_id, finish_reason,
		       prompt_tokens, completion_tokens, created_at
		FROM messages
		WHERE id = ?
	`
	var m Message
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.ConversationID, &m.ParentMessageID, &m.Role, &m.Content,
		&m.ReasoningContent, &m.ToolCalls, &m.ToolCallID, &m.FinishReason,
		&m.PromptTokens, &m.CompletionTokens, &m.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	return &m, nil
}

func (r *MessageRepository) ListByConversation(ctx context.Context, conversationID string) ([]Message, error) {
	query := `
		SELECT id, conversation_id, parent_message_id, role, content,
		       reasoning_content, tool_calls, tool_call_id, finish_reason,
		       prompt_tokens, completion_tokens, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(
			&m.ID, &m.ConversationID, &m.ParentMessageID, &m.Role, &m.Content,
			&m.ReasoningContent, &m.ToolCalls, &m.ToolCallID, &m.FinishReason,
			&m.PromptTokens, &m.CompletionTokens, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed scanning message row: %w", err)
		}
		messages = append(messages, m)
	}

	return messages, nil
}

func (r *MessageRepository) Update(ctx context.Context, m *Message) error {
	query := `
		UPDATE messages
		SET content = ?, reasoning_content = ?, tool_calls = ?, finish_reason = ?,
		    prompt_tokens = ?, completion_tokens = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		m.Content, m.ReasoningContent, m.ToolCalls, m.FinishReason,
		m.PromptTokens, m.CompletionTokens, m.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	return nil
}

func (r *MessageRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM messages WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

func (r *MessageRepository) DeleteAfter(ctx context.Context, conversationID string, createdAt time.Time) error {
	query := `DELETE FROM messages WHERE conversation_id = ? AND created_at >= ?`
	_, err := r.db.ExecContext(ctx, query, conversationID, createdAt)
	if err != nil {
		return fmt.Errorf("failed deleting messages after timestamp: %w", err)
	}
	return nil
}

func (r *MessageRepository) ClearByConversation(ctx context.Context, conversationID string) error {
	query := `DELETE FROM messages WHERE conversation_id = ?`
	_, err := r.db.ExecContext(ctx, query, conversationID)
	if err != nil {
		return fmt.Errorf("failed to clear messages: %w", err)
	}
	return nil
}
