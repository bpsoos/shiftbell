package choretemplates

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
)

func (s *Service) ListSchedules(
	ctx context.Context,
	id int,
	state schedulemodels.ScheduleFilter,
) (*models.TemplateSchedules, error) {
	if state != schedulemodels.ScheduleFilterActive &&
		state != schedulemodels.ScheduleFilterDeactivated {
		return nil, validationerrors.ErrInvalidFilter
	}
	template, err := s.persister.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get template for schedule list: %w", err)
	}
	schedules, err := s.persister.ListSchedules(ctx, id, state)
	if err != nil {
		return nil, fmt.Errorf("list template schedules: %w", err)
	}
	return &models.TemplateSchedules{
		TemplateName: template.Name,
		Schedules:    schedules,
	}, nil
}
