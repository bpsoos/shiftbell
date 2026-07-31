package choretemplates

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
)

func (p *Persister) Get(ctx context.Context, id int) (*models.ChoreTemplateDetails, error) {
	row := p.db.QueryRowContext(
		ctx,
		`
			select name, description, deactivated_at
			from chore_templates
			where id = ?
		`,
		id,
	)
	var (
		name          string
		description   sql.NullString
		deactivatedAt sql.NullTime
	)
	err := row.Scan(&name, &description, &deactivatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("select chore template: %w", err)
	}

	var deactivatedAtValue *time.Time
	if deactivatedAt.Valid {
		deactivatedAtValue = &deactivatedAt.Time
	}

	return &models.ChoreTemplateDetails{
		ChoreTemplate: models.ChoreTemplate{
			Id:            id,
			Name:          name,
			Description:   description.String,
			DeactivatedAt: deactivatedAtValue,
		},
	}, nil
}
