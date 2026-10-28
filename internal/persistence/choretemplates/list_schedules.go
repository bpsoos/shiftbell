package choretemplates

import (
	"context"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
)

func (p *Persister) ListSchedules(
	ctx context.Context,
	id int,
	state schedulemodels.ScheduleFilter,
) ([]models.ScheduleReference, error) {
	var statePredicate string
	switch state {
	case schedulemodels.ScheduleFilterActive:
		statePredicate = "deactivated_at is null"
	case schedulemodels.ScheduleFilterDeactivated:
		statePredicate = "deactivated_at is not null"
	default:
		return nil, validationerrors.ErrInvalidFilter
	}
	rows, err := p.db.QueryContext(
		ctx,
		`select id, name from schedules where chore_template_id = ? and `+statePredicate+` order by name collate nocase, id`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("select template schedules: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logging.Default().Error("close template schedule rows", "err", err)
		}
	}()
	schedules := []models.ScheduleReference{}
	for rows.Next() {
		var schedule models.ScheduleReference
		if err := rows.Scan(&schedule.Id, &schedule.Name); err != nil {
			return nil, fmt.Errorf("scan template schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate template schedules: %w", err)
	}
	return schedules, nil
}
