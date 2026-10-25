package schedules

import (
	"context"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/database"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/jmoiron/sqlx"
)

func (p *Persister) Edit(
	ctx context.Context,
	params *models.EditScheduleParams,
) (*models.Schedule, error) {
	return executeInTransaction(ctx, p.db, func(tx *sqlx.Tx) (*models.Schedule, error) {
		return editActiveSchedule(ctx, tx, params)
	})
}

func editActiveSchedule(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.EditScheduleParams,
) (*models.Schedule, error) {
	schedule, err := getSchedule(ctx, tx, params.Id)
	if err != nil {
		return nil, err
	}
	if schedule.DeactivatedAt != nil {
		return nil, models.ErrInactive
	}
	if err := updateSchedule(ctx, tx, params); err != nil {
		return nil, err
	}
	schedule.Name = params.Name
	schedule.IntervalDays = params.IntervalDays
	return schedule, nil
}

func updateSchedule(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.EditScheduleParams,
) error {
	_, err := tx.ExecContext(ctx, `
		update schedules set name = ?, interval_days = ?
		where id = ? and deactivated_at is null
	`, params.Name, params.IntervalDays, params.Id)
	if database.IsUniqueConstraintError(err) {
		return models.ErrNameConflict
	}
	if err != nil {
		return fmt.Errorf("update schedule: %w", err)
	}
	return nil
}
