package chores

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
		returnErr = combineManualOneOffRollbackError(returnErr, tx.Rollback())
	}()

	chore, err := insertManualOneOffChore(ctx, tx, params)
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

func insertManualOneOffChore(
	ctx context.Context,
	tx *sqlx.Tx,
	params *choremodels.CreateManualOneOffParams,
) (*choremodels.Chore, error) {
	result, err := tx.ExecContext(ctx, `
		insert into chores (name, description, is_complete, deadline)
		values (?, ?, ?, ?)
	`, params.Name, nullableText(params.Description), false, params.Deadline)
	if err != nil {
		return nil, fmt.Errorf("insert manual one-off chore: %w", err)
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
		return nil, resolveChoreTemplateInsertError(ctx, tx, name, err)
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
	ctx context.Context,
	tx *sqlx.Tx,
	name string,
	insertErr error,
) error {
	existingId, err := findActiveChoreTemplateIDByName(ctx, tx, name)
	if err == nil {
		return &choretemplatemodels.NameConflictError{ExistingId: existingId}
	}
	return fmt.Errorf("insert chore template: %w", insertErr)
}

func findActiveChoreTemplateIDByName(
	ctx context.Context,
	tx *sqlx.Tx,
	name string,
) (int, error) {
	var id int
	err := tx.QueryRowContext(ctx, `
		select id
		from chore_templates
		where deactivated_at is null and lower(name) = lower(?)
		order by id desc
		limit 1
	`, name).Scan(&id)
	return id, err
}

func nullableText(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func combineManualOneOffRollbackError(operationErr error, rollbackErr error) error {
	if rollbackErr == nil || errors.Is(rollbackErr, sql.ErrTxDone) {
		return operationErr
	}
	return errors.Join(
		operationErr,
		fmt.Errorf("rollback manual one-off chore transaction: %w", rollbackErr),
	)
}
