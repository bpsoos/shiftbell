package choretypes

import (
	"database/sql"
	"fmt"

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

func (p *Persister) Delete(id int) error {
	_, err := p.db.Exec(`
		delete from chore_types where id = $1
	`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete exec: %v", err)
	}
	return nil
}

func (p *Persister) Create(name string, description string) error {
	var sqlDesc sql.NullString
	if name != "" {
		sqlDesc = sql.NullString{
			String: description,
			Valid:  true,
		}
	}
	_, err := p.db.NamedExec(`
		insert into chore_types (name, description)
		values (:name, :description)
	`, map[string]any{
		"description": sqlDesc,
		"name":        name,
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
		id          int
		name        string
		description sql.NullString
	)
	results := make([]models.ChoreType, 0)
	for rows.Next() {
		err := rows.Scan(&id, &name, &description)
		if err != nil {
			return nil, fmt.Errorf("reading fetched rows: %v", err)
		}
		desc := ""
		if description.Valid {
			desc = description.String
		}
		results = append(results, models.ChoreType{
			Id:          id,
			Name:        name,
			Description: desc,
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
