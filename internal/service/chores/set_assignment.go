package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (s *Service) SetAssignment(
	ctx context.Context,
	input *models.SetAssignmentParams,
) (*models.ChoreDetails, error) {
	result, err := s.persister.SetAssignment(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("set chore assignment: %w", err)
	}
	return result, nil
}
