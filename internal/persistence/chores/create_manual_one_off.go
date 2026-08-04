package chores

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/database"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/jmoiron/sqlx"
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
		created.ChoreTemplate, err = insertSavedChoreTemplate(
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

func insertOneOffChore(
	ctx context.Context,
	tx *sqlx.Tx,
	params *choremodels.CreateOneOffChoreParams,
) (*choremodels.Chore, error) {
	result, err := tx.ExecContext(ctx, `
		insert into chores (name, description, is_complete, deadline)
		values (?, ?, ?, ?)
	`, params.Name, nullableText(params.Description), false, params.Deadline)
	if err != nil {
		return nil, fmt.Errorf("insert one-off chore: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read inserted chore id: %w", err)
	}

	return &choremodels.Chore{
		Id:          int(id),
		Status:      choremodels.ChoreStatusActive,
		Name:        params.Name,
		Description: params.Description,
		Deadline:    params.Deadline,
	}, nil
}

func insertSavedChoreTemplate(
	ctx context.Context,
	tx *sqlx.Tx,
	name string,
	description string,
) (*choretemplatemodels.ChoreTemplate, error) {
	result, err := tx.ExecContext(ctx, `
		insert into chore_templates (name, description)
		values (?, ?)
	`, name, nullableText(description))
	if err != nil {
		return nil, resolveChoreTemplateInsertError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read inserted chore template id: %w", err)
	}

	return &choretemplatemodels.ChoreTemplate{
		Id:          int(id),
		Name:        name,
		Description: description,
	}, nil
}

func resolveChoreTemplateInsertError(
	insertErr error,
) error {
	if database.IsUniqueConstraintError(insertErr) {
		return choretemplatemodels.ErrNameConflict
	}
	return fmt.Errorf("insert chore template: %w", insertErr)
}

func nullableText(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
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
