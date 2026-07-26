package choretemplates

import (
	"context"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/models"
)

func (s *Service) Get(ctx context.Context, id int) (*models.ChoreTemplateDetails, error) {
	details, err := s.persister.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get chore template: %w", err)
	}

	return details, nil
}
