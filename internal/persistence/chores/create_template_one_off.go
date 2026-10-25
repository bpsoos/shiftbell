package chores

import (
	"context"
	"fmt"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (p *Persister) CreateTemplateOneOff(
	ctx context.Context,
	params *choremodels.CreateTemplateOneOffParams,
) (_ *choremodels.CreateChoreResult, returnErr error) {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin template one-off chore transaction: %w", err)
	}
	defer func() {
		returnErr = combineTransactionRollbackError(
			returnErr,
			tx.Rollback(),
			"template one-off chore",
		)
	}()

	snapshot, err := selectActiveTemplateSnapshot(ctx, tx, params.ChoreTemplateId)
	if err != nil {
		return nil, err
	}
	chore, err := insertOneOffChore(ctx, tx, &choremodels.CreateOneOffChoreParams{
		Name:        snapshot.Name,
		Description: snapshot.Description,
		Deadline:    params.Deadline,
	})
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit template one-off chore transaction: %w", err)
	}

	return &choremodels.CreateChoreResult{Chore: chore}, nil
}
