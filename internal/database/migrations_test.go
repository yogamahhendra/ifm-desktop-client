package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDatabaseMigrationAndSeeding(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	// Verify idempotency (running Migrate twice should succeed)
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate second run failed: %v", err)
	}

	// Check seeded models
	var modelCount int
	err = db.QueryRow("SELECT COUNT(*) FROM models").Scan(&modelCount)
	if err != nil {
		t.Fatalf("failed querying models: %v", err)
	}
	if modelCount < 2 {
		t.Errorf("expected at least 2 seeded models, got %d", modelCount)
	}

	// Check seeded presets
	var presetCount int
	err = db.QueryRow("SELECT COUNT(*) FROM presets").Scan(&presetCount)
	if err != nil {
		t.Fatalf("failed querying presets: %v", err)
	}
	if presetCount < 6 {
		t.Errorf("expected at least 6 seeded presets, got %d", presetCount)
	}

	// Check seeded settings
	var settingCount int
	err = db.QueryRow("SELECT COUNT(*) FROM app_settings").Scan(&settingCount)
	if err != nil {
		t.Fatalf("failed querying app_settings: %v", err)
	}
	if settingCount < 4 {
		t.Errorf("expected at least 4 default settings, got %d", settingCount)
	}

	_ = os.Remove(dbPath)
}
