package profiles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
)

func (p *Persister) Delete(ctx context.Context, id int) (returnErr error) {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin profile deletion transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			returnErr = errors.Join(
				returnErr,
				fmt.Errorf("rollback profile deletion: %w", err),
			)
		}
	}()
	if _, err := tx.ExecContext(ctx,
		`UPDATE chores SET owner_profile_id = NULL WHERE owner_profile_id = ?`, id,
	); err != nil {
		return fmt.Errorf("unlink profile ownership: %w", err)
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM profiles WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete profile rows affected: %w", err)
	}
	if affected == 0 {
		return models.ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit profile deletion: %w", err)
	}
	return nil
}
