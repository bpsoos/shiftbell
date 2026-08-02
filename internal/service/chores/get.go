package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (s *Service) Get(
	ctx context.Context,
	id int,
) (*models.ChoreDetails, error) {
	chore, err := s.persister.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting chore: %w", err)
	}

	return chore, nil
}
