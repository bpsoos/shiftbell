package schedules

import (
	"context"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
)

func (p *Persister) Browse(
	ctx context.Context,
	params *models.BrowseSchedulesParams,
) (*models.SchedulePage, error) {
	stateCondition := "s.deactivated_at is null"
	if params.Filter == models.ScheduleFilterDeactivated {
		stateCondition = "s.deactivated_at is not null"
	}
	rows, err := p.db.QueryContext(ctx, fmt.Sprintf(`
		select s.id, s.name, s.chore_template_id, t.name, s.interval_days,
			s.deactivated_at, c.id, c.deadline
		from schedules s
		join chore_templates t on t.id = s.chore_template_id
		left join chores c on c.schedule_id = s.id and c.is_complete = 0
		where %s and (? = '' or instr(lower(s.name), lower(?)) > 0)
		order by s.id desc
		limit ? offset ?
	`, stateCondition), params.Search, params.Search, params.Limit+1, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("select schedules: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logging.Default().Error("error closing rows", "err", err)
		}
	}()
	schedules := make([]models.Schedule, 0, params.Limit+1)
	for rows.Next() {
		var schedule models.Schedule
		if err := rows.Scan(&schedule.Id, &schedule.Name, &schedule.ChoreTemplateId,
			&schedule.ChoreTemplateName, &schedule.IntervalDays, &schedule.DeactivatedAt,
			&schedule.ActiveChoreId, &schedule.ActiveChoreDeadline); err != nil {
			return nil, fmt.Errorf("scan schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read schedules: %w", err)
	}
	more := len(schedules) > params.Limit
	if more {
		schedules = schedules[:params.Limit]
	}
	return &models.SchedulePage{Schedules: schedules, More: more}, nil
}
