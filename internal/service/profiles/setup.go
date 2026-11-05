package profiles

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
)

func (s *Service) Setup(
	ctx context.Context,
	params *models.CreateParams,
) (*models.Profile, error) {
	values, err := s.normalize(params)
	if err != nil {
		return nil, err
	}
	profile, err := s.persister.CreateInitial(ctx, values)
	if err != nil {
		return nil, fmt.Errorf("set up profile: %w", err)
	}
	return profile, nil
}

func (s *Service) SetupRequired(ctx context.Context) (bool, error) {
	exists, err := s.persister.HasProfiles(ctx)
	if err != nil {
		return false, fmt.Errorf("check profile setup: %w", err)
	}
	return !exists, nil
}
