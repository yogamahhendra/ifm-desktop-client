package preset

import (
	"context"
	"database/sql"
	"testing"

	"ifm-chat/internal/database/repositories"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed opening in-memory db: %v", err)
	}

	schema := `
	CREATE TABLE presets (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		system_prompt TEXT NOT NULL,
		model_id TEXT,
		temperature REAL,
		top_p REAL,
		max_tokens INTEGER,
		reasoning_effort TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed creating schema: %v", err)
	}
	return db
}

func TestPresetServiceCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repositories.NewPresetRepository(db)
	svc := NewService(repo)
	ctx := context.Background()

	// 1. Validation test: name required
	_, err := svc.Create(ctx, repositories.PresetItem{})
	if err == nil {
		t.Errorf("expected error for empty preset name")
	}

	// 2. Create valid preset
	temp := 0.7
	created, err := svc.Create(ctx, repositories.PresetItem{
		Name:         "Golang Architect",
		Description:  "Clean architecture guru",
		SystemPrompt: "You are a senior Golang architect.",
		Temperature:  &temp,
	})
	if err != nil {
		t.Fatalf("failed creating preset: %v", err)
	}
	if created.ID == "" || created.Name != "Golang Architect" {
		t.Fatalf("unexpected created preset: %+v", created)
	}

	// 3. Get preset
	got, err := svc.Get(ctx, created.ID)
	if err != nil || got == nil {
		t.Fatalf("failed getting preset: %v", err)
	}
	if got.Description != "Clean architecture guru" {
		t.Errorf("expected description match, got %s", got.Description)
	}

	// 4. Update preset
	got.Description = "Updated Architect description"
	if err := svc.Update(ctx, *got); err != nil {
		t.Fatalf("failed updating preset: %v", err)
	}
	// Test Update with empty ID
	if err := svc.Update(ctx, repositories.PresetItem{}); err == nil {
		t.Errorf("expected error updating preset with empty ID")
	}

	// 5. List presets
	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("failed listing presets: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 preset, got %d", len(list))
	}

	// 6. Delete preset
	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("failed deleting preset: %v", err)
	}
	deleted, err := svc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("error checking deleted preset: %v", err)
	}
	if deleted != nil {
		t.Errorf("expected deleted preset to be nil")
	}
}
