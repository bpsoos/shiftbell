package chores

import (
	"fmt"
	"time"

	"github.com/bpsoos/shiftbell/internal/models"
)

func (p *Persister) GetBatch(offset int, limit int) (*models.GetChoreBatchResult, error) {
	rows, err := p.db.Query(
		`
			select
				id,
				description,
				last_completed_at,
				deadline
			from chores
			where is_complete = false
			order by deadline asc
			limit ?
			offset ?
		`,
		limit+1,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("db query selecting chores: %v", err)
	}
	var (
		id              int
		description     string
		lastCompletedAt time.Time
		deadline        time.Time
	)
	results := make([]models.Chore, 0)
	for rows.Next() {
		err := rows.Scan(&id, &description, &lastCompletedAt, &deadline)
		if err != nil {
			return nil, fmt.Errorf("reading fetched rows: %v", err)
		}
		results = append(results, models.Chore{
			Id:              id,
			Status:          models.ChoreStatusIncomplete,
			Description:     description,
			LastCompletedAt: lastCompletedAt,
			Deadline:        deadline,
		})
	}
	more := len(results) > limit
	if more {
		results = results[:limit]
	}

	return &models.GetChoreBatchResult{
		Chores: results,
		More:   more,
	}, nil
}
