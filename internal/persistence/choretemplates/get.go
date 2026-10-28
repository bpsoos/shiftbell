package choretemplates

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
)

func (p *Persister) Get(
	ctx context.Context,
	id int,
) (*models.ChoreTemplateDetails, error) {
	row := p.db.QueryRowContext(
		ctx,
		`
			select t.name, t.description, t.deactivated_at,
				(select count(*) from schedules s
				 where s.chore_template_id = t.id and s.deactivated_at is null),
				(select count(*) from schedules s
				 where s.chore_template_id = t.id and s.deactivated_at is not null)
			from chore_templates t
			where t.id = ?
		`,
		id,
	)
	var (
		name                     string
		description              sql.NullString
		deactivatedAt            sql.NullTime
		activeScheduleCount      int
		deactivatedScheduleCount int
	)
	err := row.Scan(
		&name,
		&description,
		&deactivatedAt,
		&activeScheduleCount,
		&deactivatedScheduleCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("select chore template: %w", err)
	}

	var deactivatedAtValue *time.Time
	if deactivatedAt.Valid {
		deactivatedAtValue = &deactivatedAt.Time
	}

	return &models.ChoreTemplateDetails{
		Id:                       id,
		Name:                     name,
		Description:              description.String,
		DeactivatedAt:            deactivatedAtValue,
		ActiveScheduleCount:      activeScheduleCount,
		DeactivatedScheduleCount: deactivatedScheduleCount,
	}, nil
}
