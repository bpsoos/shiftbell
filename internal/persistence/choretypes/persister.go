package choretypes

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

func NewChoreTypePersister(deps *PersisterDeps) *Persister {
	return &Persister{
		db: deps.Db,
	}
}

func (p *Persister) Create(description string, intervalDays int) error {
	now := time.Now()
	_, err := p.db.NamedExec(`
		with chore_types_insert as (
			insert into chore_types (description, interval_days)
			values (:description, :interval_days)
			returning id, interval_days
		)
		insert into chores (chore_type_id, last_completed_at, deadline, is_complete)
		select id, :now, :deadline, false from chore_types_insert
	`, map[string]any{
		"description":   description,
		"interval_days": intervalDays,
		"now":           now,
		"deadline":      now.Add(time.Hour * time.Duration(intervalDays) * 24),
	})

	if err != nil {
		return fmt.Errorf("db exec inserting chores: %v", err)
	}

	return nil
}

func (p *Persister) GetBatch(offset int, limit int) (*models.GetChoreTypeBatchResult, error) {
	rows, err := p.db.Query(
		`
			select *
			from chore_types
			order by id desc
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
		intervalDays int
	)
	results := make([]models.ChoreType, 0)
	for rows.Next() {
		err := rows.Scan(&id, &description, &intervalDays)
		if err != nil {
			return nil, fmt.Errorf("reading fetched rows: %v", err)
		}
		results = append(results, models.ChoreType{
			Id:           id,
			Description:  description,
			IntervalDays: intervalDays,
		})
	}
	if len(results) > limit {
		results = results[:len(results)-1]
	}

	return &models.GetChoreTypeBatchResult{
		ChoreTypes: results,
		More:       len(results) == limit,
	}, nil
}
