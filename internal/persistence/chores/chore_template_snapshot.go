package chores

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/jmoiron/sqlx"
)

type choreTemplateSnapshot struct {
	Name        string
	Description string
}

func selectActiveTemplateSnapshot(
	ctx context.Context,
	tx *sqlx.Tx,
	choreTemplateId int,
) (*choreTemplateSnapshot, error) {
	var (
		name        string
		description sql.NullString
		deactivated sql.NullTime
	)
	err := tx.QueryRowContext(ctx, `
		select name, description, deactivated_at
		from chore_templates where id = ?
	`, choreTemplateId).Scan(&name, &description, &deactivated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select chore template snapshot: %w", err)
	}
	if deactivated.Valid {
		return nil, models.ErrInactive
	}
	return &choreTemplateSnapshot{Name: name, Description: description.String}, nil
}
