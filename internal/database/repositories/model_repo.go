package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ModelRepository struct {
	db *sql.DB
}

func NewModelRepository(db *sql.DB) *ModelRepository {
	return &ModelRepository{db: db}
}

func (r *ModelRepository) List(ctx context.Context) ([]ModelItem, error) {
	query := `
		SELECT id, name, model_identifier, description, temperature, top_p, top_k, max_tokens, reasoning_effort, is_default, created_at, updated_at
		FROM models
		ORDER BY is_default DESC, name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer rows.Close()

	var list []ModelItem
	for rows.Next() {
		var m ModelItem
		if err := rows.Scan(
			&m.ID, &m.Name, &m.ModelIdentifier, &m.Description,
			&m.Temperature, &m.TopP, &m.TopK, &m.MaxTokens,
			&m.ReasoningEffort, &m.IsDefault, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed scanning model row: %w", err)
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *ModelRepository) Get(ctx context.Context, id string) (*ModelItem, error) {
	query := `
		SELECT id, name, model_identifier, description, temperature, top_p, top_k, max_tokens, reasoning_effort, is_default, created_at, updated_at
		FROM models
		WHERE id = ?
	`
	var m ModelItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.Name, &m.ModelIdentifier, &m.Description,
		&m.Temperature, &m.TopP, &m.TopK, &m.MaxTokens,
		&m.ReasoningEffort, &m.IsDefault, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}
	return &m, nil
}

func (r *ModelRepository) GetByIdentifier(ctx context.Context, identifier string) (*ModelItem, error) {
	query := `
		SELECT id, name, model_identifier, description, temperature, top_p, top_k, max_tokens, reasoning_effort, is_default, created_at, updated_at
		FROM models
		WHERE model_identifier = ?
	`
	var m ModelItem
	err := r.db.QueryRowContext(ctx, query, identifier).Scan(
		&m.ID, &m.Name, &m.ModelIdentifier, &m.Description,
		&m.Temperature, &m.TopP, &m.TopK, &m.MaxTokens,
		&m.ReasoningEffort, &m.IsDefault, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get model by identifier: %w", err)
	}
	return &m, nil
}

func (r *ModelRepository) Create(ctx context.Context, m *ModelItem) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	m.CreatedAt = now
	m.UpdatedAt = now

	query := `
		INSERT INTO models (
			id, name, model_identifier, description, temperature, top_p, top_k,
			max_tokens, reasoning_effort, is_default, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		m.ID, m.Name, m.ModelIdentifier, m.Description, m.Temperature, m.TopP, m.TopK,
		m.MaxTokens, m.ReasoningEffort, m.IsDefault, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed creating model: %w", err)
	}
	return nil
}

func (r *ModelRepository) Update(ctx context.Context, m *ModelItem) error {
	m.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE models
		SET name = ?, model_identifier = ?, description = ?, temperature = ?,
		    top_p = ?, top_k = ?, max_tokens = ?, reasoning_effort = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		m.Name, m.ModelIdentifier, m.Description, m.Temperature,
		m.TopP, m.TopK, m.MaxTokens, m.ReasoningEffort, m.UpdatedAt, m.ID,
	)
	if err != nil {
		return fmt.Errorf("failed updating model: %w", err)
	}
	return nil
}

func (r *ModelRepository) SetDefault(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "UPDATE models SET is_default = 0"); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE models SET is_default = 1 WHERE id = ?", id); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ModelRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM models WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed deleting model: %w", err)
	}
	return nil
}
