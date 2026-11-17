package settings

import (
	"context"
	"strings"

	"ifm-chat/internal/config"
	"ifm-chat/internal/database/repositories"
	"ifm-chat/internal/security"
)

type AppConfigResponse struct {
	Theme            string            `json:"theme"`
	IFMBaseURL       string            `json:"ifm_base_url"`
	DefaultModel     string            `json:"default_model"`
	DefaultPreset    string            `json:"default_preset"`
	RequestTimeout   string            `json:"request_timeout"`
	StreamingEnabled string            `json:"streaming_enabled"`
	Temperature      string            `json:"temperature"`
	TopP             string            `json:"top_p"`
	MaxTokens        string            `json:"max_tokens"`
	ReasoningEffort  string            `json:"reasoning_effort"`
	HasAPIKey        bool              `json:"has_api_key"`
	HasTavilyKey     bool              `json:"has_tavily_key"`
	DatabasePath     string            `json:"database_path"`
	AllSettings      map[string]string `json:"all_settings"`
}

type Service struct {
	repo     *repositories.SettingsRepository
	security *security.CredentialService
	paths    *config.AppPaths
}

func NewService(repo *repositories.SettingsRepository, sec *security.CredentialService, paths *config.AppPaths) *Service {
	return &Service{
		repo:     repo,
		security: sec,
		paths:    paths,
	}
}

func (s *Service) GetConfig(ctx context.Context) (*AppConfigResponse, error) {
	all, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	getOrDefault := func(k, def string) string {
		if v, ok := all[k]; ok && v != "" {
			return v
		}
		return def
	}

	return &AppConfigResponse{
		Theme:            getOrDefault("theme", "dark"),
		IFMBaseURL:       getOrDefault("ifm_base_url", "https://api.ifm.ai/v1"),
		DefaultModel:     getOrDefault("default_model", "IFM/K2-Horizon-375B-A23B"),
		DefaultPreset:    getOrDefault("default_preset", ""),
		RequestTimeout:   getOrDefault("request_timeout", "60"),
		StreamingEnabled: getOrDefault("streaming_enabled", "true"),
		Temperature:      getOrDefault("temperature", "1.0"),
		TopP:             getOrDefault("top_p", "0.95"),
		MaxTokens:        getOrDefault("max_tokens", "4096"),
		ReasoningEffort:  getOrDefault("reasoning_effort", "high"),
		HasAPIKey:        s.security.HasAPIKey(),
		HasTavilyKey:     s.hasTavilyKey(),
		DatabasePath:     s.paths.DatabaseURL,
		AllSettings:      all,
	}, nil
}

func (s *Service) hasTavilyKey() bool {
	k, err := s.security.GetTavilyKey()
	return err == nil && len(strings.TrimSpace(k)) > 0
}

func (s *Service) SaveSettings(ctx context.Context, settings map[string]string) error {
	return s.repo.SetBatch(ctx, settings)
}

func (s *Service) SaveAPIKey(apiKey string) error {
	return s.security.SetAPIKey(apiKey)
}

func (s *Service) DeleteAPIKey() error {
	return s.security.DeleteAPIKey()
}

func (s *Service) GetPaths() *config.AppPaths {
	return s.paths
}
