package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/jmoiron/sqlx"
)

func (p *Persister) CorrectCompletion(
	ctx context.Context,
	params *models.CorrectCompletionParams,
) (*models.CorrectCompletionResult, error) {
	return executeInTransaction(
		ctx,
		p.db,
		"chore completion correction",
		func(tx *sqlx.Tx) (*models.CorrectCompletionResult, error) {
			return correctCompletedChore(ctx, tx, params)
		},
	)
}

func correctCompletedChore(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.CorrectCompletionParams,
) (*models.CorrectCompletionResult, error) {
	if err := recordCompletionCorrection(ctx, tx, params); err != nil {
		return nil, err
	}
	chore, err := getChore(ctx, tx, params.Id)
	if err != nil {
		return nil, err
	}
	result := &models.CorrectCompletionResult{Chore: chore}
	if chore.ScheduleId == nil {
		return result, nil
	}
	result.Successor, err = getSuccessor(ctx, tx, chore.Id)
	if err != nil {
		return nil, err
	}
	if !isActiveSuccessor(result.Successor) {
		return result, nil
	}
	if err := updateSuccessorDeadline(ctx, tx, chore, result.Successor); err != nil {
		return nil, err
	}
	return result, nil
}

func isActiveSuccessor(successor *models.Chore) bool {
	return successor != nil && successor.Status == models.ChoreStatusActive
}

func recordCompletionCorrection(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.CorrectCompletionParams,
) error {
	execution, err := tx.ExecContext(
		ctx,
		`
		update chores
		set completed_on = ?
		where id = ? and is_complete = true
	`,
		params.CompletedOn,
		params.Id,
	)
	if err != nil {
		return fmt.Errorf("correct chore completion: %w", err)
	}
	affected, err := execution.RowsAffected()
	if err != nil {
		return fmt.Errorf("correct chore completion rows affected: %w", err)
	}
	if affected == 0 {
		return models.ErrNotFound
	}
	return nil
}

func updateSuccessorDeadline(
	ctx context.Context,
	tx *sqlx.Tx,
	predecessor *models.Chore,
	successor *models.Chore,
) error {
	var intervalDays int
	err := tx.QueryRowContext(
		ctx,
		`select interval_days from schedules where id = ?`,
		*predecessor.ScheduleId,
	).Scan(&intervalDays)
	if err != nil {
		return fmt.Errorf("select schedule interval for correction: %w", err)
	}
	deadline := predecessor.CompletedOn.AddDate(0, 0, intervalDays)
	_, err = tx.ExecContext(
		ctx,
		`
			update chores set deadline = ?
			where id = ? and is_complete = false
		`,
		deadline,
		successor.Id,
	)
	if err != nil {
		return fmt.Errorf("update successor deadline: %w", err)
	}
	successor.Deadline = deadline
	return nil
}
