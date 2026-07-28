package choretemplates

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	serviceerrors "github.com/bpsoos/shiftbell/internal/service"
)

func (s *Service) Browse(ctx context.Context, params *models.BrowseChoreTemplatesParams) (*models.ChoreTemplatePage, error) {
	if params == nil {
		return nil, serviceerrors.ErrInvalidLimit
	}

	filter := params.Filter
	if filter == "" {
		filter = models.ChoreTemplateFilterActive
	}
	if filter != models.ChoreTemplateFilterActive && filter != models.ChoreTemplateFilterDeactivated {
		return nil, serviceerrors.ErrInvalidFilter
	}
	if params.Offset < 0 {
		return nil, serviceerrors.ErrInvalidOffset
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.ErrInvalidLimit
	}

	search, valid := s.normalizer.NormalizeSearch(params.Search)
	if !valid {
		return nil, serviceerrors.ErrInvalidSearch
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
