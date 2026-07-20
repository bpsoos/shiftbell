package chores

import (
	"database/sql"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/models"
)

func (p *Persister) Create(params *models.CreateChoreParams) (*models.Chore, error) {
	var descNullable sql.NullString
	if params.Description != "" {
		descNullable = sql.NullString{
			String: params.Description,
			Valid:  true,
		}
	}
	var deadlineNullable sql.NullTime
	if !params.Deadline.IsZero() {
		deadlineNullable = sql.NullTime{
			Time:  params.Deadline,
			Valid: true,
		}
	}
	result, err := p.db.NamedExec(
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
	if err != nil {
		return nil, fmt.Errorf("db exec inserting chore: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id for chore: %w", err)
	}

	return &models.Chore{
		Id:          int(id),
		Name:        params.Name,
		Status:      models.ChoreStatusActive,
		Description: params.Description,
		Deadline:    params.Deadline,
	}, nil
}
