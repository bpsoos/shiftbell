package choretemplates

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func executeInTransaction[T any](
	ctx context.Context,
	db *sqlx.DB,
	operation func(*sqlx.Tx) (T, error),
) (_ T, returnErr error) {
	var zero T
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return zero, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		returnErr = combineTransactionRollbackError(returnErr, tx.Rollback())
	}()

	result, err := operation(tx)
	if err != nil {
		return zero, err
	}
	if err := tx.Commit(); err != nil {
		return zero, fmt.Errorf("commit transaction: %w", err)
	}
	return result, nil
}

func combineTransactionRollbackError(operationErr error, rollbackErr error) error {
	if rollbackErr == nil || errors.Is(rollbackErr, sql.ErrTxDone) {
		return operationErr
	}
	return errors.Join(operationErr, fmt.Errorf("rollback transaction: %w", rollbackErr))
}
