package chores

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/database"
	models "github.com/bpsoos/shiftbell/internal/models/chores"
	templatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/jmoiron/sqlx"
)

func (p *Persister) EditScheduled(
	ctx context.Context,
	params *models.EditScheduledChoreParams,
) (*models.EditChoreResult, error) {
	return executeInTransaction(
		ctx,
		p.db,
		"scheduled chore editing",
		func(tx *sqlx.Tx) (*models.EditChoreResult, error) {
			return editScheduledChore(ctx, tx, params)
		},
	)
}

func editScheduledChore(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.EditScheduledChoreParams,
) (*models.EditChoreResult, error) {
	chore, err := getChore(ctx, tx, params.Id)
	if err != nil {
		return nil, err
	}
	if chore.Status != models.ChoreStatusActive || chore.ScheduleId == nil {
		return nil, models.ErrNotFound
	}
	if err := updateScheduledChoreSnapshot(ctx, tx, params); err != nil {
		return nil, err
	}
	if params.AlsoUpdateChoreTemplate {
		if err := updateLinkedChoreTemplate(
			ctx,
			tx,
			*chore.ScheduleId,
			params.Name,
			params.Description,
		); err != nil {
			return nil, err
		}
	}
	return getChore(ctx, tx, params.Id)
}

func updateScheduledChoreSnapshot(
	ctx context.Context,
	tx *sqlx.Tx,
	params *models.EditScheduledChoreParams,
) error {
	_, err := tx.ExecContext(ctx, `
		update chores set name = ?, description = ? where id = ?
	`, params.Name, nullableText(params.Description), params.Id)
	if err != nil {
		return fmt.Errorf("update scheduled chore snapshot: %w", err)
	}
	return nil
}

func updateLinkedChoreTemplate(
	ctx context.Context,
	tx *sqlx.Tx,
	scheduleId int,
	name, description string,
) error {
	var templateId int
	err := tx.QueryRowContext(ctx, `select chore_template_id from schedules where id = ?`, scheduleId).
		Scan(&templateId)
	if errors.Is(err, sql.ErrNoRows) {
		return schedulemodels.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("select linked chore template id: %w", err)
	}
	if _, err := selectActiveTemplateSnapshot(ctx, tx, templateId); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		update chore_templates set name = ?, description = ? where id = ?
	`, name, nullableText(description), templateId)
	if err != nil {
		if database.IsUniqueConstraintError(err) {
			return templatemodels.ErrNameConflict
		}
		return fmt.Errorf("update linked chore template: %w", err)
	}
	return nil
}
