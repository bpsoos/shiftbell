package chores

import (
	"context"
	"database/sql"
	"fmt"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
)

func (p *Persister) CreateManualOneOff(
	ctx context.Context,
	params *choremodels.CreateManualOneOffParams,
) (*choremodels.CreateChoreResult, error) {
	var description sql.NullString
	if params.Description != "" {
		description = sql.NullString{String: params.Description, Valid: true}
	}

	result, err := p.db.ExecContext(ctx, `
		insert into chores (name, description, is_complete, deadline)
		values (?, ?, ?, ?)
	`, params.Name, description, false, params.Deadline)
	if err != nil {
		return nil, fmt.Errorf("insert manual one-off chore: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read inserted chore id: %w", err)
	}

	return &choremodels.CreateChoreResult{
		Chore: &choremodels.Chore{
			Id:          int(id),
			Status:      choremodels.ChoreStatusActive,
			Name:        params.Name,
			Description: params.Description,
			Deadline:    params.Deadline,
		},
	}, nil
}
