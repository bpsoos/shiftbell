package chores

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/jmoiron/sqlx"
)

func (p *Persister) Complete(
	ctx context.Context,
	params *models.CompleteChoreParams,
) (*models.CompleteChoreResult, error) {
	return executeInTransaction(
		ctx,
		p.db,
		"chore completion",
		func(tx *sqlx.Tx) (*models.CompleteChoreResult, error) {
			return completeChore(ctx, tx, params)
		},
	)
}

func completeChore(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.CompleteChoreParams,
) (*models.CompleteChoreResult, error) {
	affected, err := recordCompletion(ctx, tx, params)
	if err != nil {
		return nil, err
	}
	chore, err := getChore(ctx, tx, params.Id)
	if err != nil {
		return nil, err
	}
	result := &models.CompleteChoreResult{Chore: chore}
	if chore.ScheduleId == nil {
		return result, nil
	}
	if affected > 0 {
		result.Successor, err = createSuccessor(ctx, tx, chore)
	} else {
		result.Successor, err = getSuccessor(ctx, tx, chore.Id)
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func recordCompletion(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.CompleteChoreParams,
) (int64, error) {
	execution, err := tx.ExecContext(
		ctx,
		`
			update chores
			set is_complete = true,
				completed_on = ?
			where id = ? and is_complete = false
		`,
		params.CompletedOn,
		params.Id,
	)
	if err != nil {
		return 0, fmt.Errorf("complete chore: %w", err)
	}
	affected, err := execution.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("complete chore rows affected: %w", err)
	}
	return affected, nil
}

func createSuccessor(
	ctx context.Context,
	tx *sqlx.Tx,
	chore *models.Chore,
) (*models.Chore, error) {
	var (
		templateID    int
		intervalDays  int
		deactivatedAt sql.NullTime
	)
	err := tx.QueryRowContext(
		ctx,
		`
			select chore_template_id, interval_days, deactivated_at
			from schedules where id = ?
		`,
		*chore.ScheduleId,
	).Scan(
		&templateID,
		&intervalDays,
		&deactivatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, schedulemodels.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select schedule for completion: %w", err)
	}
	if deactivatedAt.Valid {
		return nil, schedulemodels.ErrInactive
	}
	snapshot, err := selectActiveTemplateSnapshot(ctx, tx, templateID)
	if err != nil {
		return nil, err
	}
	return insertSuccessor(ctx, tx, &models.Chore{
		ScheduleId:    chore.ScheduleId,
		PredecessorId: &chore.Id,
		Name:          snapshot.Name,
		Description:   snapshot.Description,
		Status:        models.ChoreStatusActive,
		Deadline:      chore.CompletedOn.AddDate(0, 0, intervalDays),
	})
}

func insertSuccessor(
	ctx context.Context,
	tx *sqlx.Tx,
	successor *models.Chore,
) (*models.Chore, error) {
	result, err := tx.ExecContext(
		ctx,
		`
			insert into chores (name, description, is_complete, deadline, schedule_id, predecessor_id)
			values (?, ?, false, ?, ?, ?)
		`,
		successor.Name,
		nullableText(successor.Description),
		successor.Deadline,
		successor.ScheduleId,
		successor.PredecessorId,
	)
	if err != nil {
		return nil, fmt.Errorf("insert successor: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read inserted successor id: %w", err)
	}
	successor.Id = int(id)
	return successor, nil
}
