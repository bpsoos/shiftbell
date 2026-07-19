package chores

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bpsoos/shiftbell/internal/models"
)

func (p *Persister) SetLastCompletedAt(id int, lastUpdatedAt time.Time) (*models.Chore, error) {
	row := p.db.QueryRow(
		`
			update chores
			set last_completed_at = $2
			where c.id = $1
			returning description,
				last_completed_at,
				deadline,
				completed_at
		`,
		id,
		lastUpdatedAt,
	)
	var (
		description         string
		lastCompletedAt     time.Time
		deadline            time.Time
		status              models.ChoreStatus
		completedAtNullable sql.NullTime
		completedAt         time.Time
	)
	err := row.Scan(&description, &lastCompletedAt, &deadline, &completedAtNullable)
	if err != nil {
		return nil, fmt.Errorf("update chore query: %v", err)
	}

	if completedAtNullable.Valid {
		completedAt = completedAtNullable.Time
		status = models.ChoreStatusComplete
	}
	status = models.ChoreStatusIncomplete

	return &models.Chore{
		Id:              id,
		Status:          status,
		Description:     description,
		Deadline:        deadline,
		LastCompletedAt: lastCompletedAt,
		CompletedAt:     completedAt,
	}, nil
}
