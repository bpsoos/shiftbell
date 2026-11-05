package profiles

import (
	"context"
	"errors"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
)

func (s *Service) Initialize(ctx context.Context, name string) error {
	if name == "" {
		return nil
	}
	exists, err := s.persister.HasProfiles(ctx)
	if err != nil {
		return fmt.Errorf("check profile initialization: %w", err)
	}
	if exists {
		return nil
	}
	_, err = s.Setup(ctx, &models.CreateParams{Name: name})
	if errors.Is(err, models.ErrAlreadyInitialized) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("initialize profile: %w", err)
	}
	return nil
}
