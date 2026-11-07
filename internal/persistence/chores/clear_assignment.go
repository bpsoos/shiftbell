package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/jmoiron/sqlx"
)

func (p *Persister) ClearAssignment(
	ctx context.Context,
	id int,
) (*models.ChoreDetails, error) {
	return executeInTransaction(
		ctx,
		p.db,
		"clear chore assignment",
		func(tx *sqlx.Tx) (*models.ChoreDetails, error) {
			chore, err := getChore(ctx, tx, id)
			if err != nil {
				return nil, err
			}
			if _, err := tx.ExecContext(
				ctx,
				`update chores set owner_profile_id = null where id = ?`,
				id,
			); err != nil {
				return nil, fmt.Errorf("clear chore assignment: %w", err)
			}
			chore.OwnerProfileId = nil
			return chore, nil
		},
	)
}
