package choretemplates

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/database"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/jmoiron/sqlx"
)

func (p *Persister) Edit(
	ctx context.Context,
	params *models.EditChoreTemplateParams,
) (*models.ChoreTemplate, error) {
	return executeInTransaction(
		ctx,
		p.db,
		func(tx *sqlx.Tx) (*models.ChoreTemplate, error) {
			return editActiveChoreTemplate(ctx, tx, params)
		},
	)
}

func editActiveChoreTemplate(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.EditChoreTemplateParams,
) (*models.ChoreTemplate, error) {
	choreTemplate, err := getActiveChoreTemplateForMutation(ctx, tx, params.Id)
	if err != nil {
		return nil, err
	}
	description := sql.NullString{
		String: params.Description,
		Valid:  params.Description != "",
	}

	_, err = tx.ExecContext(
		ctx,
		`
			update chore_templates
			set name = ?, description = ?
			where id = ? and deactivated_at is null
		`,
		params.Name,
		description,
		params.Id,
	)
	if err != nil {
		if database.IsUniqueConstraintError(err) {
			return nil, models.ErrNameConflict
		}
		return nil, fmt.Errorf("update chore template: %w", err)
	}

	choreTemplate.Name = params.Name
	choreTemplate.Description = params.Description

	return choreTemplate, nil
}
