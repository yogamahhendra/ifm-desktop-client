package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type APIKeyItem struct {
	ID                    string     `json:"id"`
	Name                  string     `json:"name"`
	KeyMask               string     `json:"key_mask"`
	IsActive              bool       `json:"is_active"`
	TotalPromptTokens     int        `json:"total_prompt_tokens"`
	TotalCompletionTokens int        `json:"total_completion_tokens"`
	TotalRequests         int        `json:"total_requests"`
	LastUsedAt            *time.Time `json:"last_used_at"`
	CreatedAt             time.Time  `json:"created_at"`
}

type APIKeyRepository struct {
	db *sql.DB
}

func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, item *APIKeyItem) error {
	if item.ID == "" {
		item.ID = uuid.NewString()
	}
	item.CreatedAt = time.Now().UTC()

	// If this is set as active, deactivate others
	if item.IsActive {
		_, _ = r.db.ExecContext(ctx, "UPDATE api_keys SET is_active = 0")
	}

	query := `
		INSERT INTO api_keys (
			id, name, key_mask, is_active, total_prompt_tokens,
			total_completion_tokens, total_requests, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.Name, item.KeyMask, item.IsActive,
		item.TotalPromptTokens, item.TotalCompletionTokens, item.TotalRequests,
		item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed creating api key record: %w", err)
	}
	return nil
}

func (r *APIKeyRepository) List(ctx context.Context) ([]APIKeyItem, error) {
	query := `
		SELECT id, name, key_mask, is_active, total_prompt_tokens,
		       total_completion_tokens, total_requests, last_used_at, created_at
		FROM api_keys
		ORDER BY is_active DESC, created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed listing api keys: %w", err)
	}
	defer rows.Close()

	var list []APIKeyItem
	for rows.Next() {
		var k APIKeyItem
		if err := rows.Scan(
			&k.ID, &k.Name, &k.KeyMask, &k.IsActive, &k.TotalPromptTokens,
			&k.TotalCompletionTokens, &k.TotalRequests, &k.LastUsedAt, &k.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed scanning api key: %w", err)
		}
		list = append(list, k)
	}
	return list, nil
}

func (r *APIKeyRepository) GetActive(ctx context.Context) (*APIKeyItem, error) {
	query := `
		SELECT id, name, key_mask, is_active, total_prompt_tokens,
		       total_completion_tokens, total_requests, last_used_at, created_at
		FROM api_keys
		WHERE is_active = 1
		LIMIT 1
	`
	var k APIKeyItem
	err := r.db.QueryRowContext(ctx, query).Scan(
		&k.ID, &k.Name, &k.KeyMask, &k.IsActive, &k.TotalPromptTokens,
		&k.TotalCompletionTokens, &k.TotalRequests, &k.LastUsedAt, &k.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed getting active api key: %w", err)
	}
	return &k, nil
}

func (r *APIKeyRepository) SetActive(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "UPDATE api_keys SET is_active = 0"); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE api_keys SET is_active = 1 WHERE id = ?", id); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM api_keys WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *APIKeyRepository) RecordUsage(ctx context.Context, id string, promptTokens, completionTokens int) error {
	now := time.Now().UTC()
	query := `
		UPDATE api_keys
		SET total_prompt_tokens = total_prompt_tokens + ?,
		    total_completion_tokens = total_completion_tokens + ?,
		    total_requests = total_requests + 1,
		    last_used_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, promptTokens, completionTokens, now, id)
	return err
}
