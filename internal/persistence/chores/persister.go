package chores

import (
	"fmt"
	"time"

	"github.com/bpsoos/shiftbell/internal/models"
	"github.com/jmoiron/sqlx"
)

type Persister struct {
	db *sqlx.DB
}

type PersisterDeps struct {
	Db *sqlx.DB
}

func NewPersister(deps *PersisterDeps) *Persister {
	return &Persister{
		db: deps.Db,
	}
}

func (p *Persister) GetBatch(offset int, limit int) (*models.GetChoreBatchResult, error) {
	rows, err := p.db.Query(
		`
			select
				c.id,
				ct.description,
				c.created_at,
				c.deadline,
				ct.interval_days
			from chores c
			join chore_types ct
				on c.chore_type_id = ct.id
			order by c.deadline desc
			offset $1
			limit $2 + 1
		`,
		offset,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("db query selecting chores: %v", err)
	}
	var (
		id           int
		description  string
		createdAt    time.Time
		deadline     time.Time
		intervalDays int
	)
	results := make([]models.Chore, 0)
	for rows.Next() {
		err := rows.Scan(&id, &description, &createdAt, &deadline, &intervalDays)
		if err != nil {
			return nil, fmt.Errorf("reading fetched rows: %v", err)
		}
		results = append(results, models.Chore{
			Id:           id,
			Description:  description,
			CreatedAt:    createdAt,
			Deadline:     deadline,
			IntervalDays: intervalDays,
		})
	}
	if len(results) > limit {
		results = results[:len(results)-1]
	}

	return &models.GetChoreBatchResult{
		Chores: results,
		More:   len(results) == limit,
	}, nil
}

func (p *Persister) PatchStatus(id int, isComplete bool) error {
	return nil
}
