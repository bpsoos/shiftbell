package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (p *Persister) CorrectCompletion(
	ctx context.Context,
	params *models.CorrectCompletionParams,
) (*models.CorrectCompletionResult, error) {
	execution, err := p.db.ExecContext(
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
		return nil, fmt.Errorf("correct chore completion: %w", err)
	}
	affected, err := execution.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("correct chore completion rows affected: %w", err)
	}
	if affected == 0 {
		return nil, models.ErrNotFound
	}
	chore, err := p.Get(ctx, params.Id)
	if err != nil {
		return nil, err
	}
	return &models.CorrectCompletionResult{Chore: chore}, nil
}
