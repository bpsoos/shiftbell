package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (p *Persister) Complete(
	ctx context.Context,
	params *models.CompleteChoreParams,
) (*models.CompleteChoreResult, error) {
	_, err := p.db.ExecContext(
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
		return nil, fmt.Errorf("complete chore: %w", err)
	}
	chore, err := p.Get(ctx, params.Id)
	if err != nil {
		return nil, err
	}
	return &models.CompleteChoreResult{Chore: chore}, nil
}
