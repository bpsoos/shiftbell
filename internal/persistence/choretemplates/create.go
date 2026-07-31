package choretemplates

import (
	"context"
	"database/sql"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
)

func (p *Persister) Create(ctx context.Context, params *models.CreateChoreTemplateParams) (*models.ChoreTemplate, error) {
	var sqlDesc sql.NullString
	if params.Description != "" {
		sqlDesc = sql.NullString{
			String: params.Description,
			Valid:  true,
		}
	}
	result, err := p.db.ExecContext(ctx, `
		insert into chore_templates (name, description)
		values (?, ?)
	`, params.Name, sqlDesc)
	if err != nil {
		return nil, fmt.Errorf("insert chore template: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read inserted chore template id: %w", err)
	}

	return &models.ChoreTemplate{
		Id:          int(id),
		Name:        params.Name,
		Description: params.Description,
	}, nil
}
