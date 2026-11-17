package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ConversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(ctx context.Context, c *Conversation) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	if c.SessionID == "" {
		c.SessionID = uuid.NewString()
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now

	query := `
		INSERT INTO conversations (id, title, model_id, preset_id, session_id, system_prompt, folder_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		c.ID, c.Title, c.ModelID, c.PresetID, c.SessionID, c.SystemPrompt, c.FolderPath, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}
	return nil
}

func (r *ConversationRepository) Get(ctx context.Context, id string) (*Conversation, error) {
	query := `
		SELECT id, title, model_id, preset_id, session_id, system_prompt, folder_path, created_at, updated_at
		FROM conversations
		WHERE id = ?
	`
	var c Conversation
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Title, &c.ModelID, &c.PresetID, &c.SessionID, &c.SystemPrompt, &c.FolderPath, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	return &c, nil
}

func (r *ConversationRepository) List(ctx context.Context, search string) ([]Conversation, error) {
	var rows *sql.Rows
	var err error

	if search != "" {
		query := `
			SELECT DISTINCT c.id, c.title, c.model_id, c.preset_id, c.session_id, c.system_prompt, c.folder_path, c.created_at, c.updated_at
			FROM conversations c
			LEFT JOIN messages m ON c.id = m.conversation_id
			WHERE c.title LIKE ? OR m.content LIKE ?
			ORDER BY c.updated_at DESC
		`
		pattern := "%" + search + "%"
		rows, err = r.db.QueryContext(ctx, query, pattern, pattern)
	} else {
		query := `
			SELECT id, title, model_id, preset_id, session_id, system_prompt, folder_path, created_at, updated_at
			FROM conversations
			ORDER BY updated_at DESC
		`
		rows, err = r.db.QueryContext(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var list []Conversation
	for rows.Next() {
		var c Conversation
		if err := rows.Scan(&c.ID, &c.Title, &c.ModelID, &c.PresetID, &c.SessionID, &c.SystemPrompt, &c.FolderPath, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed scanning conversation row: %w", err)
		}
		list = append(list, c)
	}

	return list, nil
}

func (r *ConversationRepository) Update(ctx context.Context, c *Conversation) error {
	c.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE conversations
		SET title = ?, model_id = ?, preset_id = ?, session_id = ?, system_prompt = ?, folder_path = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		c.Title, c.ModelID, c.PresetID, c.SessionID, c.SystemPrompt, c.FolderPath, c.UpdatedAt, c.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}
	return nil
}

func (r *ConversationRepository) UpdateFolderPath(ctx context.Context, id string, folderPath *string) error {
	now := time.Now().UTC()
	query := `UPDATE conversations SET folder_path = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, folderPath, now, id)
	if err != nil {
		return fmt.Errorf("failed to update conversation folder_path: %w", err)
	}
	return nil
}

func (r *ConversationRepository) UpdateTitle(ctx context.Context, id, title string) error {
	now := time.Now().UTC()
	query := `UPDATE conversations SET title = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, title, now, id)
	if err != nil {
		return fmt.Errorf("failed to update conversation title: %w", err)
	}
	return nil
}

func (r *ConversationRepository) Touch(ctx context.Context, id string) error {
	now := time.Now().UTC()
	query := `UPDATE conversations SET updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}

func (r *ConversationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM conversations WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}
	return nil
}

func (r *ConversationRepository) ClearAll(ctx context.Context) error {
	query := `DELETE FROM conversations`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
