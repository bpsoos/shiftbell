package profiles

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
)

func (s *Service) List(ctx context.Context) ([]models.Profile, error) {
	profiles, err := s.persister.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	return profiles, nil
}
