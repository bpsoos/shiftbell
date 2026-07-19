package chores

import (
	"database/sql"
	"time"

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
