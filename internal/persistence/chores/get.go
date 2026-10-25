package chores

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/jmoiron/sqlx"
)

func (p *Persister) Get(ctx context.Context, id int) (*models.Chore, error) {
	return getChore(ctx, p.db, id)
}

func getChore(
	ctx context.Context,
	queryer sqlx.QueryerContext,
	id int,
) (*models.Chore, error) {
	var (
		name          string
		description   sql.NullString
		isComplete    bool
		completedOn   sql.NullTime
		deadline      time.Time
		scheduleId    *int
		predecessorId *int
	)
	err := queryer.QueryRowxContext(ctx, `
		select name, description, is_complete, completed_on, deadline, schedule_id, predecessor_id
		from chores
		where id = ?
	`, id).Scan(&name, &description, &isComplete, &completedOn, &deadline, &scheduleId, &predecessorId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get chore: %w", err)
	}

	status := models.ChoreStatusActive
	if isComplete {
		status = models.ChoreStatusCompleted
	}
	chore := &models.Chore{
		Id:            id,
		ScheduleId:    scheduleId,
		PredecessorId: predecessorId,
		Status:        status,
		Name:          name,
		Description:   description.String,
		Deadline:      deadline,
	}
	if completedOn.Valid {
		chore.CompletedOn = completedOn.Time
	}
	return chore, nil
}

func getSuccessor(
	ctx context.Context,
	tx *sqlx.Tx,
	predecessorID int,
) (*models.Chore, error) {
	var id int
	err := tx.QueryRowContext(
		ctx,
		`select id from chores where predecessor_id = ?`,
		predecessorID,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select successor: %w", err)
	}
	return getChore(ctx, tx, id)
}
