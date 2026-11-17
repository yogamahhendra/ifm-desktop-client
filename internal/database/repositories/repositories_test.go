package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	dbPath := filepath.Join(t.TempDir(), "test_repo.db")
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	schema := `
	CREATE TABLE conversations (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		model_id TEXT NOT NULL,
		preset_id TEXT NULL,
		session_id TEXT NOT NULL,
		system_prompt TEXT NOT NULL DEFAULT '',
		folder_path TEXT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE TABLE messages (
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
	CREATE TABLE models (
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
	CREATE TABLE presets (
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
	CREATE TABLE app_settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed creating schema: %v", err)
	}
	return db
}

func TestConversationCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()
	repo := NewConversationRepository(db)

	conv := &Conversation{
		ID:        "conv-1",
		Title:     "Initial Title",
		ModelID:   "IFM/K2-Horizon-375B-A23B",
		SessionID: "sess-1",
	}

	// Create
	if err := repo.Create(ctx, conv); err != nil {
		t.Fatalf("create conversation failed: %v", err)
	}

	// Get
	found, err := repo.Get(ctx, "conv-1")
	if err != nil || found == nil {
		t.Fatalf("get conversation failed: %v", err)
	}
	if found.Title != "Initial Title" {
		t.Errorf("expected title 'Initial Title', got '%s'", found.Title)
	}

	// Update Title
	if err := repo.UpdateTitle(ctx, "conv-1", "Renamed Title"); err != nil {
		t.Fatalf("update title failed: %v", err)
	}
	found, _ = repo.Get(ctx, "conv-1")
	if found.Title != "Renamed Title" {
		t.Errorf("expected title 'Renamed Title', got '%s'", found.Title)
	}

	// List
	list, err := repo.List(ctx, "")
	if err != nil || len(list) != 1 {
		t.Fatalf("list conversations failed: %v", err)
	}

	// Delete
	if err := repo.Delete(ctx, "conv-1"); err != nil {
		t.Fatalf("delete conversation failed: %v", err)
	}
	deleted, _ := repo.Get(ctx, "conv-1")
	if deleted != nil {
		t.Errorf("expected nil after delete, got %v", deleted)
	}
}

func TestMessageCRUDAndReasoningPersistence(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	convRepo := NewConversationRepository(db)
	msgRepo := NewMessageRepository(db)

	_ = convRepo.Create(ctx, &Conversation{
		ID:        "conv-msg",
		Title:     "Chat Test",
		ModelID:   "IFM/K2-Horizon-375B-A23B",
		SessionID: "sess-msg",
	})

	reasoning := "Step 1: Analyzed concurrency patterns.\nStep 2: Structured goroutine channels."
	msg := &Message{
		ID:               "msg-1",
		ConversationID:   "conv-msg",
		Role:             "assistant",
		Content:          "Here is how Go concurrency works.",
		ReasoningContent: &reasoning,
		PromptTokens:     10,
		CompletionTokens: 50,
		CreatedAt:        time.Now().UTC(),
	}

	// Create Message with Reasoning
	if err := msgRepo.Create(ctx, msg); err != nil {
		t.Fatalf("create message failed: %v", err)
	}

	// Retrieve & Verify Reasoning Content
	found, err := msgRepo.Get(ctx, "msg-1")
	if err != nil || found == nil {
		t.Fatalf("get message failed: %v", err)
	}
	if found.ReasoningContent == nil || *found.ReasoningContent != reasoning {
		t.Errorf("reasoning content was not preserved accurately: got %v", found.ReasoningContent)
	}

	// List By Conversation
	list, err := msgRepo.ListByConversation(ctx, "conv-msg")
	if err != nil || len(list) != 1 {
		t.Fatalf("list by conversation failed: %v", err)
	}

	// Clear By Conversation
	if err := msgRepo.ClearByConversation(ctx, "conv-msg"); err != nil {
		t.Fatalf("clear by conversation failed: %v", err)
	}
	listAfter, _ := msgRepo.ListByConversation(ctx, "conv-msg")
	if len(listAfter) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(listAfter))
	}
}

func TestModelAndPresetRepositories(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	modelRepo := NewModelRepository(db)
	presetRepo := NewPresetRepository(db)

	// Test Model
	model := &ModelItem{
		ID:              "m-1",
		Name:            "Custom K2",
		ModelIdentifier: "custom/k2",
		Temperature:     1.0,
		TopP:            0.95,
		MaxTokens:       4096,
		IsDefault:       false,
	}
	if err := modelRepo.Create(ctx, model); err != nil {
		t.Fatalf("create model failed: %v", err)
	}
	if err := modelRepo.SetDefault(ctx, "m-1"); err != nil {
		t.Fatalf("set default model failed: %v", err)
	}
	gotModel, _ := modelRepo.Get(ctx, "m-1")
	if gotModel == nil || !gotModel.IsDefault {
		t.Errorf("expected model to be default")
	}

	// Test Preset
	preset := &PresetItem{
		ID:           "p-1",
		Name:         "Dev Assistant",
		SystemPrompt: "Be a developer",
	}
	if err := presetRepo.Create(ctx, preset); err != nil {
		t.Fatalf("create preset failed: %v", err)
	}
	gotPreset, _ := presetRepo.Get(ctx, "p-1")
	if gotPreset == nil || gotPreset.Name != "Dev Assistant" {
		t.Errorf("expected preset name 'Dev Assistant', got %v", gotPreset)
	}
}
