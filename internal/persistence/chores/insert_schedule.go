package chores

import (
	"context"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/database"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/jmoiron/sqlx"
)

func insertSchedule(
	ctx context.Context,
	tx *sqlx.Tx,
	name string,
	templateId, intervalDays int,
) (int, error) {
	result, err := tx.ExecContext(ctx, `
		insert into schedules (name, chore_template_id, interval_days)
		values (?, ?, ?)
	`, name, templateId, intervalDays)
	if err != nil {
		return 0, resolveScheduleInsertError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read inserted schedule id: %w", err)
	}
	return int(id), nil
}

func resolveScheduleInsertError(insertErr error) error {
	if database.IsUniqueConstraintError(insertErr) {
		return models.ErrNameConflict
	}
	return fmt.Errorf("insert schedule: %w", insertErr)
}
