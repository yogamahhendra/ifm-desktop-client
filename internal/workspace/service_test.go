package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceServiceScanAndBuildPrompt(t *testing.T) {
	tempDir := t.TempDir()

	// Create sample project structure
	_ = os.MkdirAll(filepath.Join(tempDir, "src"), 0755)
	_ = os.MkdirAll(filepath.Join(tempDir, "node_modules", "pkg"), 0755)
	_ = os.MkdirAll(filepath.Join(tempDir, ".git"), 0755)

	mainFile := filepath.Join(tempDir, "src", "main.go")
	_ = os.WriteFile(mainFile, []byte("package main\n\nfunc main() {\n\tprintln(\"Hello IFM\")\n}\n"), 0644)

	readmeFile := filepath.Join(tempDir, "README.md")
	_ = os.WriteFile(readmeFile, []byte("# Test Project\nContext test."), 0644)

	// Ignored file in node_modules
	_ = os.WriteFile(filepath.Join(tempDir, "node_modules", "pkg", "index.js"), []byte("console.log('ignored')"), 0644)

	svc := NewService()
	summary, err := svc.Scan(tempDir)
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}

	if summary.TotalFiles != 2 {
		t.Errorf("expected 2 files (main.go, README.md), got %d", summary.TotalFiles)
	}

	foundMain := false
	foundReadme := false
	for _, f := range summary.Files {
		if f.RelativePath == "src/main.go" {
			foundMain = true
		}
		if f.RelativePath == "README.md" {
			foundReadme = true
		}
		if strings.Contains(f.RelativePath, "node_modules") {
			t.Errorf("node_modules was not ignored: %s", f.RelativePath)
		}
	}

	if !foundMain || !foundReadme {
		t.Errorf("expected files not found in summary")
	}

	prompt, err := svc.BuildPrompt(tempDir, nil, 10000)
	if err != nil {
		t.Fatalf("unexpected BuildPrompt error: %v", err)
	}

	if !strings.Contains(prompt, "ATTACHED WORKSPACE CONTEXT") {
		t.Errorf("expected prompt to contain ATTACHED WORKSPACE CONTEXT header")
	}
	if !strings.Contains(prompt, "Hello IFM") {
		t.Errorf("expected prompt to contain code content 'Hello IFM'")
	}
}
