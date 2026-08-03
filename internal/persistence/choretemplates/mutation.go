package choretemplates

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/jmoiron/sqlx"
)

func getActiveChoreTemplateForMutation(
	ctx context.Context,
	tx *sqlx.Tx,
	id int,
) (*models.ChoreTemplate, error) {
	var (
		name          string
		description   sql.NullString
		deactivatedAt sql.NullTime
	)
	err := tx.QueryRowxContext(
		ctx,
		`select name, description, deactivated_at from chore_templates where id = ?`,
		id,
	).Scan(&name, &description, &deactivatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("select chore template for mutation: %w", err)
	}

	if deactivatedAt.Valid {
		return nil, models.ErrInactive
	}

	return &models.ChoreTemplate{
		Id:          id,
		Name:        name,
		Description: description.String,
	}, nil
}
