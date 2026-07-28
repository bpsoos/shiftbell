package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	serviceerrors "github.com/bpsoos/shiftbell/internal/service"
)

func (s *Service) Browse(ctx context.Context, params *models.BrowseChoresParams) (*models.ChorePage, error) {
	status := params.Status
	if status == "" {
		status = models.ChoreStatusActive
	}
	if status != models.ChoreStatusActive && status != models.ChoreStatusCompleted {
		return nil, serviceerrors.ErrInvalidFilter
	}
	if params.Offset < 0 {
		return nil, serviceerrors.ErrInvalidOffset
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.ErrInvalidLimit
	}

	page, err := s.persister.Browse(ctx, &models.BrowseChoresParams{
		Status: status,
		Offset: params.Offset,
		Limit:  params.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("browse chores: %w", err)
	}

	return page, nil
}
