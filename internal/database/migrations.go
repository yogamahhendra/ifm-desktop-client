package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Migration struct {
	Version     int
	Description string
	Apply       func(tx *sql.Tx) error
}

var migrations = []Migration{
	{
		Version:     1,
		Description: "Create initial schema and seed default models, presets, and settings",
		Apply: func(tx *sql.Tx) error {
			schema := `
			CREATE TABLE IF NOT EXISTS conversations (
				id TEXT PRIMARY KEY,
				title TEXT NOT NULL,
				model_id TEXT NOT NULL,
				preset_id TEXT NULL,
				session_id TEXT NOT NULL,
				system_prompt TEXT NOT NULL DEFAULT '',
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL
			);

			CREATE INDEX IF NOT EXISTS idx_conversations_updated_at ON conversations(updated_at DESC);

			CREATE TABLE IF NOT EXISTS messages (
				id TEXT PRIMARY KEY,
				conversation_id TEXT NOT NULL,
				parent_message_id TEXT NULL,
				role TEXT NOT NULL,
				content TEXT NOT NULL,
				reasoning_content TEXT NULL,
				tool_calls TEXT NULL,
				tool_call_id TEXT NULL,
				finish_reason TEXT NULL,
				prompt_tokens INTEGER DEFAULT 0,
				completion_tokens INTEGER DEFAULT 0,
				created_at DATETIME NOT NULL,
				FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
			);

			CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
			CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at ASC);

			CREATE TABLE IF NOT EXISTS models (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				model_identifier TEXT NOT NULL UNIQUE,
				description TEXT NOT NULL DEFAULT '',
				temperature REAL NOT NULL DEFAULT 1.0,
				top_p REAL NOT NULL DEFAULT 0.95,
				top_k INTEGER NULL,
				max_tokens INTEGER NOT NULL DEFAULT 4096,
				reasoning_effort TEXT NOT NULL DEFAULT '',
				is_default BOOLEAN NOT NULL DEFAULT 0,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL
			);

			CREATE INDEX IF NOT EXISTS idx_models_identifier ON models(model_identifier);

			CREATE TABLE IF NOT EXISTS presets (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				description TEXT NOT NULL DEFAULT '',
				system_prompt TEXT NOT NULL,
				model_id TEXT NULL,
				temperature REAL NULL,
				top_p REAL NULL,
				max_tokens INTEGER NULL,
				reasoning_effort TEXT NULL,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL
			);

			CREATE TABLE IF NOT EXISTS app_settings (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL,
				updated_at DATETIME NOT NULL
			);
			`
			if _, err := tx.Exec(schema); err != nil {
				return fmt.Errorf("failed executing initial schema: %w", err)
			}

			// Seed Initial Models
			now := time.Now().UTC()
			models := []struct {
				id          string
				name        string
				identifier  string
				description string
				temp        float64
				topP        float64
				maxTokens   int
				effort      string
				isDefault   bool
			}{
				{
					id:          uuid.NewString(),
					name:        "K2 Horizon 375B",
					identifier:  "IFM/K2-Horizon-375B-A23B",
					description: "IFM Flagship Reasoning & General Foundation Model (375B / 23B active)",
					temp:        1.0,
					topP:        0.95,
					maxTokens:   4096,
					effort:      "high",
					isDefault:   true,
				},
				{
					id:          uuid.NewString(),
					name:        "Jais 3 MoVA 36B",
					identifier:  "IFM/Jais-3-MoVA-36B-A4B",
					description: "IFM State-of-the-Art Bilingual Arabic/English Foundation Model (36B / 4B active)",
					temp:        0.7,
					topP:        0.9,
					maxTokens:   4096,
					effort:      "",
					isDefault:   false,
				},
			}

			for _, m := range models {
				_, err := tx.Exec(`
					INSERT OR IGNORE INTO models (id, name, model_identifier, description, temperature, top_p, max_tokens, reasoning_effort, is_default, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				`, m.id, m.name, m.identifier, m.description, m.temp, m.topP, m.maxTokens, m.effort, m.isDefault, now, now)
				if err != nil {
					return fmt.Errorf("failed seeding model %s: %w", m.identifier, err)
				}
			}

			// Seed Initial Presets
			presets := []struct {
				name        string
				description string
				prompt      string
				temp        float64
				effort      string
			}{
				{
					name:        "General Assistant",
					description: "Versatile, helpful, and concise general-purpose AI assistant",
					prompt:      "You are a helpful, accurate, and concise AI assistant powered by IFM foundation models.",
					temp:        1.0,
					effort:      "high",
				},
				{
					name:        "Software Developer",
					description: "Expert software engineer writing production-ready code",
					prompt:      "You are an expert principal software engineer. Provide clean, secure, idiomatic, and well-structured code with clear explanations when needed.",
					temp:        0.7,
					effort:      "high",
				},
				{
					name:        "Code Reviewer",
					description: "Senior engineer reviewing code for security, edge cases, and performance",
					prompt:      "You are a senior code reviewer. Review provided code thoroughly for bugs, edge cases, security vulnerabilities, performance bottlenecks, and style improvements.",
					temp:        0.5,
					effort:      "high",
				},
				{
					name:        "Technical Writer",
					description: "Documentation specialist producing clear, structured guides",
					prompt:      "You are a technical documentation specialist. Produce clear, developer-friendly documentation, tutorials, and specifications using Markdown.",
					temp:        0.7,
					effort:      "",
				},
				{
					name:        "SQL Assistant",
					description: "Database architect specializing in SQL schemas, queries, and optimization",
					prompt:      "You are an expert database administrator and SQL architect. Provide optimal schema designs, performant queries, and explain index strategies.",
					temp:        0.3,
					effort:      "high",
				},
				{
					name:        "DevOps Engineer",
					description: "Cloud infrastructure, Docker, CI/CD, and site reliability engineer",
					prompt:      "You are a Senior DevOps Engineer. Specialize in Docker, Kubernetes, CI/CD pipelines, Windows/Linux server management, and resilient infrastructure.",
					temp:        0.5,
					effort:      "high",
				},
			}

			for _, p := range presets {
				_, err := tx.Exec(`
					INSERT OR IGNORE INTO presets (id, name, description, system_prompt, temperature, top_p, max_tokens, reasoning_effort, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, 0.95, 4096, ?, ?, ?)
				`, uuid.NewString(), p.name, p.description, p.prompt, p.temp, p.effort, now, now)
				if err != nil {
					return fmt.Errorf("failed seeding preset %s: %w", p.name, err)
				}
			}

			// Seed Initial Settings (Non-sensitive only)
			settings := map[string]string{
				"theme":             "dark",
				"ifm_base_url":      "https://api.ifm.ai/v1",
				"default_model":     "IFM/K2-Horizon-375B-A23B",
				"request_timeout":   "60",
				"streaming_enabled": "true",
			}

			for k, v := range settings {
				_, err := tx.Exec(`
					INSERT OR IGNORE INTO app_settings (key, value, updated_at)
					VALUES (?, ?, ?)
				`, k, v, now)
				if err != nil {
					return fmt.Errorf("failed seeding setting %s: %w", k, err)
				}
			}

			return nil
		},
	},
	{
		Version:     2,
		Description: "Create api_keys table for multi-key management and token usage tracking",
		Apply: func(tx *sql.Tx) error {
			schema := `
			CREATE TABLE IF NOT EXISTS api_keys (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				key_mask TEXT NOT NULL,
				is_active BOOLEAN NOT NULL DEFAULT 0,
				total_prompt_tokens INTEGER DEFAULT 0,
				total_completion_tokens INTEGER DEFAULT 0,
				total_requests INTEGER DEFAULT 0,
				last_used_at DATETIME NULL,
				created_at DATETIME NOT NULL
			);
			`
			_, err := tx.Exec(schema)
			return err
		},
	},
	{
		Version:     3,
		Description: "Add folder_path to conversations table for workspace folder context",
		Apply: func(tx *sql.Tx) error {
			_, err := tx.Exec(`ALTER TABLE conversations ADD COLUMN folder_path TEXT NULL;`)
			return err
		},
	},
}

// Migrate executes all pending database migrations in a transaction.
func (db *DB) Migrate() error {
	// Create schema migrations tracker
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at DATETIME NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("failed creating schema_migrations table: %w", err)
	}

	for _, m := range migrations {
		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.Version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed checking migration status for version %d: %w", m.Version, err)
		}

		if exists > 0 {
			continue // Already applied
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed starting transaction for migration %d: %w", m.Version, err)
		}

		if err := m.Apply(tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %d failed: %w", m.Version, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, ?)",
			m.Version, m.Description, time.Now().UTC()); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed recording migration %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed committing migration %d: %w", m.Version, err)
		}
	}

	return nil
}
