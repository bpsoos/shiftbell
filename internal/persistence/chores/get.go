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
			select description,
				completed_at,
				deadline
			from chores
			where id = $1
		`,
		id,
	)

	var (
		description         string
		status              models.ChoreStatus
		deadline            time.Time
		completedAtNullable sql.NullTime
		completedAt         time.Time
	)
	err := row.Scan(&description, &completedAtNullable, &deadline)
	if err != nil {
		return nil, fmt.Errorf("get chore query: %v", err)
	}

	if completedAtNullable.Valid {
		completedAt = completedAtNullable.Time
		status = models.ChoreStatusComplete
	}
	status = models.ChoreStatusIncomplete

	return &models.Chore{
		Id:          id,
		Status:      status,
		Description: description,
		CompletedAt: completedAt,
		Deadline:    deadline,
	}, nil
}
