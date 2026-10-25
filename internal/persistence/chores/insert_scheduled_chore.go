package chores

import (
	"context"
	"fmt"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/jmoiron/sqlx"
)

func insertScheduledChore(
	ctx context.Context,
	tx *sqlx.Tx,
	snapshot *choreTemplateSnapshot,
	deadline time.Time,
	scheduleId int,
) (*models.Chore, error) {
	result, err := tx.ExecContext(ctx, `
		insert into chores (name, description, is_complete, deadline, schedule_id)
		values (?, ?, ?, ?, ?)
	`, snapshot.Name, nullableText(snapshot.Description), false, deadline, scheduleId)
	if err != nil {
		return nil, fmt.Errorf("insert scheduled chore: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read inserted chore id: %w", err)
	}
	return &models.Chore{
		Id: int(id), ScheduleId: &scheduleId,
		Name: snapshot.Name, Description: snapshot.Description,
		Status: models.ChoreStatusActive, Deadline: deadline,
	}, nil
}
