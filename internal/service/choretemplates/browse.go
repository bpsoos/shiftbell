package choretemplates

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
)

func (s *Service) Browse(
	ctx context.Context,
	params *models.BrowseChoreTemplatesParams,
) (*models.ChoreTemplatePage, error) {
	if params == nil {
		return nil, validationerrors.ErrInvalidLimit
	}

	filter := params.Filter
	if filter == "" {
		filter = models.ChoreTemplateFilterActive
	}
	if filter != models.ChoreTemplateFilterActive &&
		filter != models.ChoreTemplateFilterDeactivated {
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

	page, err := s.persister.Browse(ctx, &models.BrowseChoreTemplatesParams{
		Filter: filter,
		Search: search,
		Offset: params.Offset,
		Limit:  params.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("browse chore templates: %w", err)
	}

	return page, nil
}
