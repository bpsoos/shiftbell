package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/jmoiron/sqlx"
)

func (p *Persister) CreateTemplateScheduled(
	ctx context.Context,
	params *models.CreateTemplateScheduledParams,
) (_ *models.CreateChoreResult, returnErr error) {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin template scheduled chore transaction: %w", err)
	}
	defer func() {
		returnErr = combineTransactionRollbackError(
			returnErr,
			tx.Rollback(),
			"template scheduled chore",
		)
	}()

	snapshot, err := selectActiveTemplateSnapshot(ctx, tx, params.ChoreTemplateId)
	if err != nil {
		return nil, err
	}
	scheduleId, err := insertTemplateSchedule(ctx, tx, params)
	if err != nil {
		return nil, err
	}
	chore, err := insertTemplateScheduledChore(ctx, tx, params, snapshot, scheduleId)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit template scheduled chore transaction: %w", err)
	}
	return &models.CreateChoreResult{Chore: chore}, nil
}

func insertTemplateSchedule(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.CreateTemplateScheduledParams,
) (int, error) {
	result, err := tx.ExecContext(ctx, `
		insert into schedules (name, chore_template_id, interval_days)
		values (?, ?, ?)
	`, params.ScheduleName, params.ChoreTemplateId, params.IntervalDays)
	if err != nil {
		return 0, fmt.Errorf("insert schedule: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read inserted schedule id: %w", err)
	}
	return int(id), nil
}

func insertTemplateScheduledChore(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.CreateTemplateScheduledParams,
	snapshot *choreTemplateSnapshot,
	scheduleId int,
) (*models.Chore, error) {
	result, err := tx.ExecContext(ctx, `
		insert into chores (name, description, is_complete, deadline, schedule_id)
		values (?, ?, ?, ?, ?)
	`, snapshot.Name, nullableText(snapshot.Description), false, params.Deadline, scheduleId)
	if err != nil {
		return nil, fmt.Errorf("insert scheduled chore: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read inserted chore id: %w", err)
	}
	return &models.Chore{
		Id:          int(id),
		ScheduleId:  &scheduleId,
		Name:        snapshot.Name,
		Description: snapshot.Description,
		Status:      models.ChoreStatusActive,
		Deadline:    params.Deadline,
	}, nil
}
