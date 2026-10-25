package chores

import (
	"context"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/database"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/jmoiron/sqlx"
)

func insertChoreTemplate(
	ctx context.Context,
	tx *sqlx.Tx,
	name, description string,
) (*models.ChoreTemplate, error) {
	result, err := tx.ExecContext(ctx, `
		insert into chore_templates (name, description)
		values (?, ?)
	`, name, nullableText(description))
	if err != nil {
		return nil, resolveChoreTemplateInsertError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read inserted chore template id: %w", err)
	}
	return &models.ChoreTemplate{Id: int(id), Name: name, Description: description}, nil
}

func resolveChoreTemplateInsertError(insertErr error) error {
	if database.IsUniqueConstraintError(insertErr) {
		return models.ErrNameConflict
	}
	return fmt.Errorf("insert chore template: %w", insertErr)
}
