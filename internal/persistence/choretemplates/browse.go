package choretemplates

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
)

func (p *Persister) Browse(
	ctx context.Context,
	params *models.BrowseChoreTemplatesParams,
) (*models.ChoreTemplatePage, error) {
	stateCondition := "deactivated_at is null"
	if params.Filter == models.ChoreTemplateFilterDeactivated {
		stateCondition = "deactivated_at is not null"
	}

	rows, err := p.db.QueryContext(ctx, fmt.Sprintf(`
		select id, name, description, deactivated_at
		from chore_templates
		where %s
			and (
				? = ''
				or instr(lower(name), lower(?)) > 0
				or instr(lower(coalesce(description, '')), lower(?)) > 0
			)
		order by id desc
		limit ?
		offset ?
	`, stateCondition), params.Search, params.Search, params.Search, params.Limit+1, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("select chore templates: %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			logging.Default().Error("error closing rows", "err", err)
		}
	}()

	choreTemplates := make([]models.ChoreTemplate, 0, params.Limit+1)
	for rows.Next() {
		var (
			choreTemplate models.ChoreTemplate
			description   sql.NullString
			deactivatedAt sql.NullTime
		)
		if err := rows.Scan(
			&choreTemplate.Id,
			&choreTemplate.Name,
			&description,
			&deactivatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chore template: %w", err)
		}
		choreTemplate.Description = description.String
		if deactivatedAt.Valid {
			value := deactivatedAt.Time
			choreTemplate.DeactivatedAt = &value
		}
		choreTemplates = append(choreTemplates, choreTemplate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read chore templates: %w", err)
	}

	more := len(choreTemplates) > params.Limit
	if more {
		choreTemplates = choreTemplates[:params.Limit]
	}

	return &models.ChoreTemplatePage{ChoreTemplates: choreTemplates, More: more}, nil
}
