package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PresetRepository struct {
	db *sql.DB
}

func NewPresetRepository(db *sql.DB) *PresetRepository {
	return &PresetRepository{db: db}
}

func (r *PresetRepository) List(ctx context.Context) ([]PresetItem, error) {
	query := `
		SELECT id, name, description, system_prompt, model_id, temperature, top_p, max_tokens, reasoning_effort, created_at, updated_at
		FROM presets
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed listing presets: %w", err)
	}
	defer rows.Close()

	var list []PresetItem
	for rows.Next() {
		var p PresetItem
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.SystemPrompt, &p.ModelID,
			&p.Temperature, &p.TopP, &p.MaxTokens, &p.ReasoningEffort,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed scanning preset row: %w", err)
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *PresetRepository) Get(ctx context.Context, id string) (*PresetItem, error) {
	query := `
		SELECT id, name, description, system_prompt, model_id, temperature, top_p, max_tokens, reasoning_effort, created_at, updated_at
		FROM presets
		WHERE id = ?
	`
	var p PresetItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.SystemPrompt, &p.ModelID,
		&p.Temperature, &p.TopP, &p.MaxTokens, &p.ReasoningEffort,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed getting preset: %w", err)
	}
	return &p, nil
}

func (r *PresetRepository) Create(ctx context.Context, p *PresetItem) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	query := `
		INSERT INTO presets (
			id, name, description, system_prompt, model_id, temperature, top_p,
			max_tokens, reasoning_effort, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.Name, p.Description, p.SystemPrompt, p.ModelID,
		p.Temperature, p.TopP, p.MaxTokens, p.ReasoningEffort,
		p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed creating preset: %w", err)
	}
	return nil
}

func (r *PresetRepository) Update(ctx context.Context, p *PresetItem) error {
	p.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE presets
		SET name = ?, description = ?, system_prompt = ?, model_id = ?,
		    temperature = ?, top_p = ?, max_tokens = ?, reasoning_effort = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		p.Name, p.Description, p.SystemPrompt, p.ModelID,
		p.Temperature, p.TopP, p.MaxTokens, p.ReasoningEffort,
		p.UpdatedAt, p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed updating preset: %w", err)
	}
	return nil
}

func (r *PresetRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM presets WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed deleting preset: %w", err)
	}
	return nil
}
