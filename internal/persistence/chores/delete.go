package chores

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (p *Persister) Delete(ctx context.Context, id int) (returnErr error) {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin chore deletion transaction: %w", err)
	}
	defer func() {
		returnErr = combineTransactionRollbackError(
			returnErr,
			tx.Rollback(),
			"chore deletion",
		)
	}()

	execution, err := tx.ExecContext(ctx, `
		delete from chores
		where id = ? and (schedule_id is null or is_complete = true)
	`, id)
	if err != nil {
		return fmt.Errorf("delete chore: %w", err)
	}
	affected, err := execution.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete chore rows affected: %w", err)
	}
	if affected == 0 {
		var existingID int
		err := tx.QueryRowContext(ctx, `select id from chores where id = ?`, id).
			Scan(&existingID)
		if errors.Is(err, sql.ErrNoRows) {
			return models.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("check chore after rejected deletion: %w", err)
		}
		return models.ErrActiveScheduledCannotDelete
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit chore deletion transaction: %w", err)
	}
	return nil
}
