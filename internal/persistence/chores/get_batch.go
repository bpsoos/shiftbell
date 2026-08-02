package chores

import (
	"database/sql"
	"fmt"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (p *Persister) GetBatch(offset int, limit int) (*models.GetChoreBatchResult, error) {
	rows, err := p.db.Query(
		`
			select
				id,
				name,
				description,
				deadline
			from chores
			where is_complete = false
			order by deadline asc, id asc
			limit ?
			offset ?
		`,
		limit+1,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("db query selecting chores: %w", err)
	}
	var (
		id          int
		name        string
		description sql.NullString
		deadline    time.Time
	)
	results := make([]models.Chore, 0)
	for rows.Next() {
		err := rows.Scan(&id, &name, &description, &deadline)
		if err != nil {
			return nil, fmt.Errorf("reading fetched rows: %w", err)
		}
		results = append(results, models.Chore{
			Id:          id,
			Name:        name,
			Status:      models.ChoreStatusActive,
			Description: description.String,
			Deadline:    deadline,
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
