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
		name             string
		description      sql.NullString
		isComplete       bool
		completedOn      sql.NullTime
		deadline         time.Time
		scheduleId       *int
		predecessorId    *int
		ownerProfileId   *int
		ownerProfileName string
	)
	err := queryer.QueryRowxContext(ctx, `
		select c.name, c.description, c.is_complete, c.completed_on, c.deadline, c.schedule_id, c.predecessor_id, c.owner_profile_id, coalesce(p.name, '')
		from chores c
		left join profiles p on p.id = c.owner_profile_id
		where c.id = ?
	`, id).Scan(&name, &description, &isComplete, &completedOn, &deadline, &scheduleId, &predecessorId, &ownerProfileId, &ownerProfileName)
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
		OwnerProfileId:   ownerProfileId,
		OwnerProfileName: ownerProfileName,
		Id:               id,
		ScheduleId:       scheduleId,
		PredecessorId:    predecessorId,
		Status:           status,
		Name:             name,
		Description:      description.String,
		Deadline:         deadline,
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
