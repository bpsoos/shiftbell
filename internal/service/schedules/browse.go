package schedules

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
)

func (s *Service) Browse(
	ctx context.Context,
	params *models.BrowseSchedulesParams,
) (*models.SchedulePage, error) {
	filter := params.Filter
	if filter == "" {
		filter = models.ScheduleFilterActive
	}
	if filter != models.ScheduleFilterActive &&
		filter != models.ScheduleFilterDeactivated {
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

	page, err := s.persister.Browse(ctx, &models.BrowseSchedulesParams{
		Filter: filter,
		Search: search,
		Offset: params.Offset,
		Limit:  params.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("browse schedules: %w", err)
	}

	return page, nil
}
