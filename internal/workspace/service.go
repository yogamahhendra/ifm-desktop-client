package workspace

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// WorkspaceFile represents a single file scanned in the workspace folder.
type WorkspaceFile struct {
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	Size         int64  `json:"size"`
	Extension    string `json:"extension"`
	Tokens       int    `json:"tokens"`
	Included     bool   `json:"included"`
}

// WorkspaceSummary contains overall statistics and file list for an attached folder.
type WorkspaceSummary struct {
	FolderPath  string          `json:"folder_path"`
	FolderName  string          `json:"folder_name"`
	TotalFiles  int             `json:"total_files"`
	TotalSize   int64           `json:"total_size"`
	TotalTokens int             `json:"total_tokens"`
	Files       []WorkspaceFile `json:"files"`
}

type Service struct {
	MaxFiles         int
	MaxFileSizeBytes int64
	MaxContextTokens int
}

func NewService() *Service {
	return &Service{
		MaxFiles:         150,
		MaxFileSizeBytes: 150 * 1024, // 150 KB max per file
		MaxContextTokens: 60000,      // ~60k tokens safe for K2-Horizon
	}
}

// Ignored directory names
var ignoredDirs = map[string]bool{
	".git":             true,
	".github":          true,
	".svn":             true,
	".hg":              true,
	".idea":            true,
	".vscode":          true,
	".gemini":          true,
	"node_modules":     true,
	"vendor":           true,
	"bower_components": true,
	"dist":             true,
	"build":            true,
	"bin":              true,
	"obj":              true,
	"target":           true,
	"out":              true,
	".next":            true,
	".nuxt":            true,
	".turbo":           true,
	".cache":           true,
	"__pycache__":      true,
	".pytest_cache":    true,
	".venv":            true,
	"env":              true,
	"coverage":         true,
	"tmp":              true,
	"temp":             true,
}

// Allowed code and text extensions
var allowedExtensions = map[string]bool{
	".go": true, ".ts": true, ".js": true, ".tsx": true, ".jsx": true, ".vue": true, ".svelte": true,
	".py": true, ".rs": true, ".java": true, ".c": true, ".cpp": true, ".h": true, ".hpp": true,
	".cs": true, ".php": true, ".rb": true, ".swift": true, ".kt": true, ".scala": true,
	".sh": true, ".bash": true, ".zsh": true, ".bat": true, ".cmd": true, ".ps1": true,
	".html": true, ".htm": true, ".css": true, ".scss": true, ".sass": true, ".less": true,
	".json": true, ".jsonc": true, ".json5": true, ".yaml": true, ".yml": true, ".toml": true,
	".xml": true, ".sql": true, ".prisma": true, ".graphql": true, ".gql": true,
	".md": true, ".markdown": true, ".txt": true, ".csv": true, ".tsv": true,
	".ini": true, ".cfg": true, ".conf": true, ".env": true, ".env.example": true,
	".proto": true,
}

var allowedFilenames = map[string]bool{
	"dockerfile":    true,
	"makefile":      true,
	"procfile":      true,
	"gemfile":       true,
	"caddyfile":     true,
	"jenkinsfile":   true,
	"license":       true,
	".dockerignore": true,
	".gitignore":    true,
}

// Scan traverses the given directory and returns a summary of text/code files.
func (s *Service) Scan(folderPath string) (*WorkspaceSummary, error) {
	info, err := os.Stat(folderPath)
	if err != nil {
		return nil, fmt.Errorf("folder does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", folderPath)
	}

	cleanPath := filepath.Clean(folderPath)
	folderName := filepath.Base(cleanPath)

	var files []WorkspaceFile
	var totalSize int64
	var totalTokens int

	err = filepath.Walk(cleanPath, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil // Skip unreadable paths
		}

		// Skip directories
		if fi.IsDir() {
			base := strings.ToLower(fi.Name())
			if ignoredDirs[base] {
				return filepath.SkipDir
			}
			return nil
		}

		// Enforce max files limit
		if len(files) >= s.MaxFiles {
			return filepath.SkipDir
		}

		// Check size limit per file
		if fi.Size() == 0 || fi.Size() > s.MaxFileSizeBytes {
			return nil
		}

		baseName := strings.ToLower(fi.Name())
		ext := strings.ToLower(filepath.Ext(baseName))

		if !allowedExtensions[ext] && !allowedFilenames[baseName] {
			return nil
		}

		// Verify file isn't binary
		if isBinary(path) {
			return nil
		}

		relPath, relErr := filepath.Rel(cleanPath, path)
		if relErr != nil {
			relPath = filepath.Base(path)
		}
		// Standardize to forward slashes
		relPath = filepath.ToSlash(relPath)

		tokens := int(fi.Size()/4) + 1
		totalSize += fi.Size()
		totalTokens += tokens

		files = append(files, WorkspaceFile{
			Path:         path,
			RelativePath: relPath,
			Size:         fi.Size(),
			Extension:    ext,
			Tokens:       tokens,
			Included:     true,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking workspace folder: %w", err)
	}

	// Sort files alphabetically by relative path
	sort.Slice(files, func(i, j int) bool {
		return files[i].RelativePath < files[j].RelativePath
	})

	return &WorkspaceSummary{
		FolderPath:  cleanPath,
		FolderName:  folderName,
		TotalFiles:  len(files),
		TotalSize:   totalSize,
		TotalTokens: totalTokens,
		Files:       files,
	}, nil
}

// BuildPrompt creates the markdown workspace context prompt to be fed to the model.
func (s *Service) BuildPrompt(folderPath string, selectedRelativePaths []string, maxTokens int) (string, error) {
	summary, err := s.Scan(folderPath)
	if err != nil {
		return "", err
	}

	if len(summary.Files) == 0 {
		return "", nil
	}

	if maxTokens <= 0 {
		maxTokens = s.MaxContextTokens
	}

	selectedMap := make(map[string]bool)
	for _, p := range selectedRelativePaths {
		selectedMap[p] = true
	}

	var sb strings.Builder
	sb.WriteString("================================================================================\n")
	sb.WriteString(fmt.Sprintf("ATTACHED WORKSPACE CONTEXT: %s\n", summary.FolderName))
	sb.WriteString(fmt.Sprintf("Directory Path: %s\n", summary.FolderPath))
	sb.WriteString("================================================================================\n\n")

	sb.WriteString("### Workspace File Tree:\n")
	for _, f := range summary.Files {
		if len(selectedMap) == 0 || selectedMap[f.RelativePath] {
			sb.WriteString(fmt.Sprintf("- %s (%s, ~%d tokens)\n", f.RelativePath, formatBytes(f.Size), f.Tokens))
		}
	}
	sb.WriteString("\n### Workspace File Contents:\n\n")

	currentTokens := summary.TotalFiles * 15 // Header tokens estimate

	for _, f := range summary.Files {
		if len(selectedMap) > 0 && !selectedMap[f.RelativePath] {
			continue
		}

		if currentTokens+f.Tokens > maxTokens {
			sb.WriteString(fmt.Sprintf("\n[Note: Additional files omitted to stay within token limits. %d files included]\n", len(summary.Files)))
			break
		}

		contentBytes, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}

		content := string(contentBytes)
		sb.WriteString("--------------------------------------------------------------------------------\n")
		sb.WriteString(fmt.Sprintf("FILE: %s\n", f.RelativePath))
		sb.WriteString("--------------------------------------------------------------------------------\n")

		lang := strings.TrimPrefix(f.Extension, ".")
		if lang == "" {
			lang = "text"
		}
		sb.WriteString(fmt.Sprintf("```%s\n", lang))
		sb.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("```\n\n")

		currentTokens += f.Tokens
	}

	sb.WriteString("================================================================================\n")
	sb.WriteString("END WORKSPACE CONTEXT. You may refer to any file, code structure, or logic above.\n")
	sb.WriteString("================================================================================\n")

	return sb.String(), nil
}

// OpenInExplorer opens the folder in the native Windows Explorer.
func (s *Service) OpenInExplorer(folderPath string) error {
	cleanPath := filepath.Clean(folderPath)
	cmd := exec.Command("explorer.exe", cleanPath)
	return cmd.Start()
}

func isBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return true
	}

	// If there's a null byte in the first 512 bytes, it's almost certainly binary
	return bytes.IndexByte(buf[:n], 0) != -1
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
