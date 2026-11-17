package preset

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"ifm-chat/internal/database/repositories"
)

type Service struct {
	repo *repositories.PresetRepository
}

func NewService(repo *repositories.PresetRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]repositories.PresetItem, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (*repositories.PresetItem, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, p repositories.PresetItem) (*repositories.PresetItem, error) {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	if p.Name == "" {
		return nil, fmt.Errorf("preset name is required")
	}
	if err := s.repo.Create(ctx, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) Update(ctx context.Context, p repositories.PresetItem) error {
	if p.ID == "" {
		return fmt.Errorf("preset ID is required")
	}
	return s.repo.Update(ctx, &p)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
