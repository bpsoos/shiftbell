package choretemplates

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/jmoiron/sqlx"
)

func (p *Persister) Deactivate(
	ctx context.Context,
	params *models.DeactivateChoreTemplateParams,
) (*models.ChoreTemplate, error) {
	return executeInTransaction(
		ctx,
		p.db,
		func(tx *sqlx.Tx) (*models.ChoreTemplate, error) {
			return deactivateActiveChoreTemplate(ctx, tx, params)
		},
	)
}

func deactivateActiveChoreTemplate(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.DeactivateChoreTemplateParams,
) (*models.ChoreTemplate, error) {
	choreTemplate, err := getActiveChoreTemplateForMutation(ctx, tx, params.Id)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(
		ctx,
		`
			update chore_templates
			set deactivated_at = ?
			where id = ? and deactivated_at is null
		`,
		params.DeactivatedAt,
		params.Id,
	)
	if err != nil {
		return nil, fmt.Errorf("deactivate chore template: %w", err)
	}

	choreTemplate.DeactivatedAt = &params.DeactivatedAt

	return choreTemplate, nil
}
