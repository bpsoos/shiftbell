package chores

import (
	"database/sql"
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

func (p *Persister) Create(params *models.CreateChoreParams) (*models.Chore, error) {
	var descNullable sql.NullString
	if params.Description != "" {
		descNullable = sql.NullString{
			String: params.Description,
			Valid:  true,
		}
	}
	var deadlineNullable sql.NullTime
	if !time.Now().IsZero() {
		deadlineNullable = sql.NullTime{
			Time:  params.Deadline,
			Valid: true,
		}
	}
	p.db.NamedExec(
		`
			insert into chores
				(name, description, is_complete, deadline)
			values
				(:name, :description, :is_complete, :deadline)
		`,
		map[string]any{
			"name":        params.Name,
			"description": descNullable,
			"is_complete": false,
			"deadline":    deadlineNullable,
		},
	)

	return nil, nil
}

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

func (p *Persister) Get(id int) (*models.Chore, error) {
	row := p.db.QueryRow(
		`
			select description,
				last_completed_at,
				completed_at,
				deadline
			from chores
			where id = $1
		`,
		id,
	)

	var (
		description         string
		lastCompletedAt     time.Time
		status              models.ChoreStatus
		deadline            time.Time
		completedAtNullable sql.NullTime
		completedAt         time.Time
	)
	err := row.Scan(&description, &lastCompletedAt, &completedAtNullable, &deadline)
	if err != nil {
		return nil, fmt.Errorf("get chore query: %v", err)
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
		LastCompletedAt: lastCompletedAt,
		CompletedAt:     completedAt,
		Deadline:        deadline,
	}, nil
}

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
	if len(results) > limit {
		results = results[:len(results)-1]
	}

	return &models.GetChoreBatchResult{
		Chores: results,
		More:   len(results) == limit,
	}, nil
}

func (p *Persister) MarkComplete(id int, completedAt time.Time) error {
	_, err := p.db.Exec(
		`
			with candidate as (
				select id
				from chores
				where id = $1 and is_complete = false
			)
			update chores c
			set is_complete = true,
				completed_at = $2
			from candidate
			where c.id = candidate.id
		`,
		id,
		completedAt,
	)
	if err != nil {
		return fmt.Errorf("update chores exec: %v", err)
	}
	return nil
}
