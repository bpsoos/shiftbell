package chores

import (
	"context"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/jmoiron/sqlx"
)

func (p *Persister) CreateManualScheduled(
	ctx context.Context,
	params *models.CreateManualScheduledParams,
) (*models.CreateChoreResult, error) {
	return executeInTransaction(
		ctx,
		p.db,
		"manual scheduled chore",
		func(tx *sqlx.Tx) (*models.CreateChoreResult, error) {
			template, err := insertChoreTemplate(ctx, tx, params.Name, params.Description)
			if err != nil {
				return nil, err
			}
			scheduleId, err := insertSchedule(
				ctx,
				tx,
				params.ScheduleName,
				template.Id,
				params.IntervalDays,
			)
			if err != nil {
				return nil, err
			}
			chore, err := insertScheduledChore(ctx, tx, &choreTemplateSnapshot{
				Name: template.Name, Description: template.Description,
			}, params.Deadline, scheduleId)
			if err != nil {
				return nil, err
			}
			return &models.CreateChoreResult{Chore: chore, ChoreTemplate: template}, nil
		},
	)
}
