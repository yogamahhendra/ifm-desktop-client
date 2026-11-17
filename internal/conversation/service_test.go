package conversation

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"ifm-chat/internal/database/repositories"
)

func setupTestDB(t *testing.T) *sql.DB {
	dbPath := filepath.Join(t.TempDir(), "test_service.db")
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
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed creating schema: %v", err)
	}
	return db
}

func TestConversationService(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	convRepo := repositories.NewConversationRepository(db)
	msgRepo := repositories.NewMessageRepository(db)
	svc := NewService(convRepo, msgRepo)

	// Create
	conv, err := svc.Create(ctx, "Test Conversation", "IFM/K2-Horizon-375B-A23B", "", "You are helpful.")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Add user message
	userMsg, err := svc.AddUserMessage(ctx, conv.ID, "Explain Go channels")
	if err != nil {
		t.Fatalf("add user message failed: %v", err)
	}
	if userMsg.Role != "user" {
		t.Errorf("expected role 'user', got '%s'", userMsg.Role)
	}

	// Add assistant message with reasoning
	reasoning := "1. Think about unbuffered vs buffered channels.\n2. Summarize channel select patterns."
	asstMsg, err := svc.AddAssistantMessage(ctx, conv.ID, "Channels synchronize goroutines.", &reasoning, 12, 45)
	if err != nil {
		t.Fatalf("add assistant message failed: %v", err)
	}

	// Verify messages
	messages, err := svc.GetMessages(ctx, conv.ID)
	if err != nil || len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	if messages[1].ReasoningContent == nil || *messages[1].ReasoningContent != reasoning {
		t.Errorf("assistant reasoning content mismatch")
	}

	// Prepare regenerate: deleting asstMsg and any following messages
	time.Sleep(10 * time.Millisecond)
	if err := svc.PrepareRegenerate(ctx, conv.ID, asstMsg.ID); err != nil {
		t.Fatalf("PrepareRegenerate failed: %v", err)
	}
	remaining, err := svc.GetMessages(ctx, conv.ID)
	if err != nil || len(remaining) != 1 {
		t.Fatalf("expected 1 remaining message after prepare regenerate, got %d", len(remaining))
	}
	if remaining[0].Role != "user" {
		t.Errorf("expected remaining message to be user message")
	}

	// Test UpdateFolderPath
	folder := "F:/test/folder"
	if err := svc.UpdateFolderPath(ctx, conv.ID, &folder); err != nil {
		t.Fatalf("UpdateFolderPath failed: %v", err)
	}
	updatedConv, err := svc.Get(ctx, conv.ID)
	if err != nil || updatedConv.FolderPath == nil || *updatedConv.FolderPath != folder {
		t.Fatalf("expected folder path to be updated to %s", folder)
	}
}

func TestGenerateTitleFromContent(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Explain Go concurrency patterns", "Explain Go concurrency patterns"},
		{"Write a Python script that calculates Fibonacci numbers efficiently", "Write a Python script that"},
		{"", "New Conversation"},
		{"   \n\n   What is Docker?   ", "What is Docker"},
	}

	for _, c := range cases {
		got := GenerateTitleFromContent(c.input)
		if got != c.expected {
			t.Errorf("GenerateTitleFromContent(%q) = %q; want %q", c.input, got, c.expected)
		}
	}
}
