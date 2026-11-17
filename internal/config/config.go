package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AppPaths struct {
	BaseDir     string
	DatabaseDir string
	DatabaseURL string
	LogsDir     string
	ExportsDir  string
}

// GetAppPaths returns the application data directories in %APPDATA%/IFMChat/
func GetAppPaths() (*AppPaths, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		appData = filepath.Join(homeDir, "AppData", "Roaming")
	}

	baseDir := filepath.Join(appData, "IFMChat")
	dbDir := filepath.Join(baseDir, "database")
	logsDir := filepath.Join(baseDir, "logs")
	exportsDir := filepath.Join(baseDir, "exports")

	for _, dir := range []string{baseDir, dbDir, logsDir, exportsDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	configFile := filepath.Join(baseDir, "config.json")
	dbFile := filepath.Join(dbDir, "ifm-chat.db")

	// If user moved database, read custom path from config.json
	if data, err := os.ReadFile(configFile); err == nil {
		type customCfg struct {
			CustomDatabasePath string `json:"custom_database_path"`
		}
		var cc customCfg
		if jsonErr := json.Unmarshal(data, &cc); jsonErr == nil && cc.CustomDatabasePath != "" {
			dbFile = cc.CustomDatabasePath
			dbDir = filepath.Dir(dbFile)
		}
	}

	return &AppPaths{
		BaseDir:     baseDir,
		DatabaseDir: dbDir,
		DatabaseURL: dbFile,
		LogsDir:     logsDir,
		ExportsDir:  exportsDir,
	}, nil
}

// SaveCustomDatabasePath updates custom_database_path in %APPDATA%/IFMChat/config.json
func SaveCustomDatabasePath(newDbPath string) error {
	paths, err := GetAppPaths()
	if err != nil {
		return err
	}
	configFile := filepath.Join(paths.BaseDir, "config.json")

	type customCfg struct {
		CustomDatabasePath string `json:"custom_database_path"`
	}
	cfg := customCfg{
		CustomDatabasePath: newDbPath,
	}

	bytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, bytes, 0644)
}
