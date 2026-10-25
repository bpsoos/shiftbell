package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
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
	scheduleId, err := insertSchedule(
		ctx,
		tx,
		params.ScheduleName,
		params.ChoreTemplateId,
		params.IntervalDays,
	)
	if err != nil {
		return nil, err
	}
	chore, err := insertScheduledChore(ctx, tx, snapshot, params.Deadline, scheduleId)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit template scheduled chore transaction: %w", err)
	}
	return &models.CreateChoreResult{Chore: chore}, nil
}
