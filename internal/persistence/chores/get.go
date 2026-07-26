package chores

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bpsoos/shiftbell/internal/models"
)

func (p *Persister) Get(id int) (*models.Chore, error) {
	row := p.db.QueryRow(
		`
			select name,
				description,
				completed_on,
				deadline
			from chores
			where id = $1
		`,
		id,
	)

	var (
		name                string
		description         sql.NullString
		status              = models.ChoreStatusActive
		deadline            time.Time
		completedOnNullable sql.NullTime
		completedOn         time.Time
	)
	err := row.Scan(&name, &description, &completedOnNullable, &deadline)
	if err != nil {
		return nil, fmt.Errorf("get chore query: %v", err)
	}

	if completedOnNullable.Valid {
		completedOn = completedOnNullable.Time
		status = models.ChoreStatusCompleted
	}

	return &models.Chore{
		Id:          id,
		Name:        name,
		Status:      status,
		Description: description.String,
		CompletedOn: completedOn,
		Deadline:    deadline,
	}, nil
}
