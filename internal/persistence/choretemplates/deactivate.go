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
	references, err := activeScheduleReferences(ctx, tx, params.Id)
	if err != nil {
		return nil, err
	}
	if len(references) > 0 {
		return nil, &models.ActiveScheduleReferencesError{Schedules: references}
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

func activeScheduleReferences(
	ctx context.Context,
	tx *sqlx.Tx,
	choreTemplateID int,
) ([]models.ActiveScheduleReference, error) {
	references := make([]models.ActiveScheduleReference, 0)
	err := tx.SelectContext(ctx, &references, `
		select id, name
		from schedules
		where chore_template_id = ? and deactivated_at is null
		order by id desc
	`, choreTemplateID)
	if err != nil {
		return nil, fmt.Errorf("select active schedule references: %w", err)
	}
	return references, nil
}
