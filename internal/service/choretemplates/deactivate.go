package choretemplates

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
)

func (s *Service) Deactivate(ctx context.Context, id int) (*models.ChoreTemplate, error) {
	choreTemplate, err := s.persister.Deactivate(ctx, &models.DeactivateChoreTemplateParams{
		Id:            id,
		DeactivatedAt: s.now(),
	})
	if err != nil {
		return nil, fmt.Errorf("deactivate chore template: %w", err)
	}

	return choreTemplate, nil
}
