package chores

import (
	"context"
	"fmt"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/jmoiron/sqlx"
)

func insertOneOffChore(
	ctx context.Context,
	tx *sqlx.Tx,
	params *choremodels.CreateOneOffChoreParams,
) (*choremodels.Chore, error) {
	result, err := tx.ExecContext(ctx, `
		insert into chores (name, description, is_complete, deadline)
		values (?, ?, ?, ?)
	`, params.Name, nullableText(params.Description), false, params.Deadline)
	if err != nil {
		return nil, fmt.Errorf("insert one-off chore: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read inserted chore id: %w", err)
	}

	return &choremodels.Chore{
		Id:          int(id),
		Status:      choremodels.ChoreStatusActive,
		Name:        params.Name,
		Description: params.Description,
		Deadline:    params.Deadline,
	}, nil
}
