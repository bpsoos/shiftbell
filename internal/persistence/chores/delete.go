package chores

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (p *Persister) Delete(ctx context.Context, id int) error {
	execution, err := p.db.ExecContext(ctx, `delete from chores where id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete chore: %w", err)
	}
	affected, err := execution.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete chore rows affected: %w", err)
	}
	if affected == 0 {
		return models.ErrNotFound
	}
	return nil
}
