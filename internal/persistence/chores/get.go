package chores

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (p *Persister) Get(ctx context.Context, id int) (*models.Chore, error) {
	var (
		name        string
		description sql.NullString
		isComplete  bool
		completedOn sql.NullTime
		deadline    time.Time
	)
	err := p.db.QueryRowContext(ctx, `
		select name, description, is_complete, completed_on, deadline
		from chores
		where id = ?
	`, id).Scan(&name, &description, &isComplete, &completedOn, &deadline)
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
		Id:          id,
		Status:      status,
		Name:        name,
		Description: description.String,
		Deadline:    deadline,
	}
	if completedOn.Valid {
		chore.CompletedOn = completedOn.Time
	}
	return chore, nil
}
