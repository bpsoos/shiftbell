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

func (p *Persister) SetLastCompletedAt(id int, lastUpdatedAt time.Time) (*models.Chore, error) {
	row := p.db.QueryRow(
		`
			update chores as c
			set last_completed_at = $2
			from chore_types as ct
			where c.id = $1
			and c.chore_type_id = ct.id
			returning ct.description,
				c.last_completed_at,
				ct.interval_days,
				c.completed_at
		`,
		id,
		lastUpdatedAt,
	)
	var (
		description         string
		lastCompletedAt     time.Time
		status              models.ChoreStatus
		intervalDays        int
		completedAtNullable sql.NullTime
		completedAt         time.Time
	)
	err := row.Scan(&description, &lastCompletedAt, &intervalDays, &completedAtNullable)
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
		IntervalDays:    intervalDays,
		Deadline:        lastCompletedAt.Add(24 * time.Hour * time.Duration(intervalDays)),
		LastCompletedAt: lastCompletedAt,
		CompletedAt:     completedAt,
	}, nil
}

func (p *Persister) Get(id int) (*models.Chore, error) {
	row := p.db.QueryRow(
		`
			select description,
				last_completed_at,
				interval_days,
				completed_at,
				deadline
			from chores_full
			where id = $1
		`,
		id,
	)

	var (
		description         string
		lastCompletedAt     time.Time
		status              models.ChoreStatus
		deadline            time.Time
		intervalDays        int
		completedAtNullable sql.NullTime
		completedAt         time.Time
	)
	err := row.Scan(&description, &lastCompletedAt, &intervalDays, &completedAtNullable, &deadline)
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
		IntervalDays:    intervalDays,
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
				interval_days,
				deadline
			from chores_full c
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
		intervalDays    int
		deadline        time.Time
	)
	results := make([]models.Chore, 0)
	for rows.Next() {
		err := rows.Scan(&id, &description, &lastCompletedAt, &intervalDays, &deadline)
		if err != nil {
			return nil, fmt.Errorf("reading fetched rows: %v", err)
		}
		results = append(results, models.Chore{
			Id:              id,
			Status:          models.ChoreStatusIncomplete,
			Description:     description,
			LastCompletedAt: lastCompletedAt,
			Deadline:        deadline,
			IntervalDays:    intervalDays,
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
		markCompleteQuery,
		id,
		completedAt,
	)
	if err != nil {
		return fmt.Errorf("update chores exec: %v", err)
	}
	return nil
}

const markCompleteQuery = `
	with candidate as (
		select
			id,
			chore_type_id
		from chores
		where id = $1 and is_complete = false
	),
	updated as (
		update chores c
		set is_complete = true,
			completed_at = $2
		from candidate
		where c.id = candidate.id
		returning candidate.chore_type_id
	)
	insert into chores (chore_type_id, last_completed_at, is_complete)
	select
		u.chore_type_id,
		$2,
		false
	from updated u
	join chore_types ct on ct.id = u.chore_type_id
`
