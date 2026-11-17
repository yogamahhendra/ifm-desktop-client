package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) GetAll(ctx context.Context) (map[string]string, error) {
	query := `SELECT key, value FROM app_settings`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed getting app settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("failed scanning app setting: %w", err)
		}
		settings[k] = v
	}
	return settings, nil
}

func (r *SettingsRepository) Get(ctx context.Context, key string) (string, error) {
	query := `SELECT value FROM app_settings WHERE key = ?`
	var val string
	err := r.db.QueryRowContext(ctx, query, key).Scan(&val)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed getting setting %s: %w", key, err)
	}
	return val, nil
}

func (r *SettingsRepository) Set(ctx context.Context, key, value string) error {
	now := time.Now().UTC()
	query := `
		INSERT INTO app_settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`
	_, err := r.db.ExecContext(ctx, query, key, value, now)
	if err != nil {
		return fmt.Errorf("failed setting %s: %w", key, err)
	}
	return nil
}

func (r *SettingsRepository) SetBatch(ctx context.Context, settings map[string]string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO app_settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for k, v := range settings {
		if _, err := stmt.ExecContext(ctx, k, v, now); err != nil {
			return fmt.Errorf("failed setting batch key %s: %w", k, err)
		}
	}

	return tx.Commit()
}
