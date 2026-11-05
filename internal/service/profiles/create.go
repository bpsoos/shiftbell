package profiles

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
)

func (s *Service) Create(
	ctx context.Context,
	params *models.CreateParams,
) (*models.Profile, error) {
	values, err := s.normalize(params)
	if err != nil {
		return nil, err
	}
	profile, err := s.persister.Create(ctx, values)
	if err != nil {
		return nil, fmt.Errorf("create profile: %w", err)
	}
	return profile, nil
}
