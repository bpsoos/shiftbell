package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (p *Persister) EditOneOff(
	ctx context.Context,
	params *models.EditOneOffChoreParams,
) (*models.EditChoreResult, error) {
	execution, err := p.db.ExecContext(
		ctx,
		`
		update chores
		set name = ?,
			description = nullif(?, ''),
			deadline = ?
		where id = ? and is_complete = false
	`,
		params.Name,
		params.Description,
		params.Deadline,
		params.Id,
	)
	if err != nil {
		return nil, fmt.Errorf("edit one-off chore: %w", err)
	}
	affected, err := execution.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("edit one-off chore rows affected: %w", err)
	}
	if affected == 0 {
		return nil, models.ErrNotFound
	}
	return p.Get(ctx, params.Id)
}
