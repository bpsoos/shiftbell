package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
)

func (s *Service) Browse(ctx context.Context, params *models.BrowseChoresParams) (*models.ChorePage, error) {
	status := params.Status
	if status == "" {
		status = models.ChoreStatusActive
	}
	if status != models.ChoreStatusActive && status != models.ChoreStatusCompleted {
		return nil, validationerrors.ErrInvalidFilter
	}
	if params.Offset < 0 {
		return nil, validationerrors.ErrInvalidOffset
	}
	if params.Limit <= 0 {
		return nil, validationerrors.ErrInvalidLimit
	}
	search, err := s.normalizer.NormalizeSearch(params.Search)
	if err != nil {
		return nil, validationerrors.ErrInvalidSearch
	}

	page, err := s.persister.Browse(ctx, &models.BrowseChoresParams{
		Status: status,
		Search: search,
		Offset: params.Offset,
		Limit:  params.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("browse chores: %w", err)
	}

	return page, nil
}
