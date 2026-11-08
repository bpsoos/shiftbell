package profiles

import (
	"context"
	"fmt"
)

func (s *Service) HasMultipleProfiles(ctx context.Context) (bool, error) {
	multiple, err := s.persister.HasMultipleProfiles(ctx)
	if err != nil {
		return false, fmt.Errorf("check multiple profiles: %w", err)
	}
	return multiple, nil
}
