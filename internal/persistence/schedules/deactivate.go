package schedules

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/jmoiron/sqlx"
)

func (p *Persister) Deactivate(
	ctx context.Context,
	params *models.DeactivateScheduleParams,
) (*models.Schedule, error) {
	return executeInTransaction(ctx, p.db, func(tx *sqlx.Tx) (*models.Schedule, error) {
		return deactivateSchedule(ctx, tx, params)
	})
}

func deactivateSchedule(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.DeactivateScheduleParams,
) (*models.Schedule, error) {
	schedule, err := getSchedule(ctx, tx, params.Id)
	if err != nil {
		return nil, err
	}
	if schedule.DeactivatedAt != nil {
		return schedule, nil
	}
	if err := recordScheduleDeactivation(ctx, tx, params); err != nil {
		return nil, err
	}
	if err := removeActiveChore(ctx, tx, params.Id); err != nil {
		return nil, err
	}
	schedule.DeactivatedAt = &params.DeactivatedAt
	schedule.ActiveChoreId = nil
	schedule.ActiveChoreDeadline = nil
	return schedule, nil
}

func recordScheduleDeactivation(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.DeactivateScheduleParams,
) error {
	_, err := tx.ExecContext(ctx, `
		update schedules set deactivated_at = ?
		where id = ? and deactivated_at is null
	`, params.DeactivatedAt, params.Id)
	if err != nil {
		return fmt.Errorf("deactivate schedule: %w", err)
	}
	return nil
}

func removeActiveChore(ctx context.Context, tx *sqlx.Tx, scheduleID int) error {
	_, err := tx.ExecContext(ctx, `
		delete from chores where schedule_id = ? and is_complete = false
	`, scheduleID)
	if err != nil {
		return fmt.Errorf("remove active schedule chore: %w", err)
	}
	return nil
}
