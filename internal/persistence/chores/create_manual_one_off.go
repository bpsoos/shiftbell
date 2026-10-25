package chores

import (
	"context"
	"fmt"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (p *Persister) CreateManualOneOff(
	ctx context.Context,
	params *choremodels.CreateManualOneOffParams,
) (_ *choremodels.CreateChoreResult, returnErr error) {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin manual one-off chore transaction: %w", err)
	}
	defer func() {
		returnErr = combineTransactionRollbackError(
			returnErr,
			tx.Rollback(),
			"manual one-off chore",
		)
	}()

	chore, err := insertOneOffChore(ctx, tx, &choremodels.CreateOneOffChoreParams{
		Name:        params.Name,
		Description: params.Description,
		Deadline:    params.Deadline,
	})
	if err != nil {
		return nil, err
	}
	created := &choremodels.CreateChoreResult{Chore: chore}
	if params.SaveAsChoreTemplate {
		created.ChoreTemplate, err = insertChoreTemplate(
			ctx,
			tx,
			params.Name,
			params.Description,
		)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit manual one-off chore transaction: %w", err)
	}

	return created, nil
}
