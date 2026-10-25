package schedules

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
)

func (p *Persister) Get(ctx context.Context, id int) (*models.ScheduleDetails, error) {
	row := p.db.QueryRowContext(
		ctx,
		`
			select s.id, s.name, s.chore_template_id, t.name, s.interval_days,
				s.deactivated_at, c.id, c.deadline
			from schedules s
			join chore_templates t on t.id = s.chore_template_id
			left join chores c on c.schedule_id = s.id and c.is_complete = 0
			where s.id = ?
		`,
		id,
	)
	var schedule models.ScheduleDetails
	err := row.Scan(
		&schedule.Id,
		&schedule.Name,
		&schedule.ChoreTemplateId,
		&schedule.ChoreTemplateName,
		&schedule.IntervalDays,
		&schedule.DeactivatedAt,
		&schedule.ActiveChoreId,
		&schedule.ActiveChoreDeadline,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select schedule: %w", err)
	}
	return &schedule, nil
}
