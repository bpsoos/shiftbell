package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (s *Service) ClearAssignment(
	ctx context.Context,
	id int,
) (*models.ChoreDetails, error) {
	chore, err := s.persister.ClearAssignment(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("clear chore assignment: %w", err)
	}
	return chore, nil
}
