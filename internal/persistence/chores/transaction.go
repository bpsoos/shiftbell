package chores

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
	name string,
	operation func(*sqlx.Tx) (T, error),
) (_ T, returnErr error) {
	var zero T
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return zero, fmt.Errorf("begin %s transaction: %w", name, err)
	}
	defer func() {
		returnErr = combineTransactionRollbackError(returnErr, tx.Rollback(), name)
	}()

	result, err := operation(tx)
	if err != nil {
		return zero, err
	}
	if err := tx.Commit(); err != nil {
		return zero, fmt.Errorf("commit %s transaction: %w", name, err)
	}
	return result, nil
}

func combineTransactionRollbackError(
	operationErr error,
	rollbackErr error,
	transactionName string,
) error {
	if rollbackErr == nil || errors.Is(rollbackErr, sql.ErrTxDone) {
		return operationErr
	}
	return errors.Join(
		operationErr,
		fmt.Errorf("rollback %s transaction: %w", transactionName, rollbackErr),
	)
}
