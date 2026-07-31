package choretemplates

import (
	"database/sql"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
)

func (p *Persister) GetBatch(offset int, limit int) (*models.GetChoreTemplateBatchResult, error) {
	rows, err := p.db.Query(
		`
			select id, name, description
			from chore_templates
			order by id desc
			limit ?
			offset ?
		`,
		limit+1,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("db query selecting chore templates: %v", err)
	}
	var (
		id          int
		name        string
		description sql.NullString
	)
	results := make([]models.ChoreTemplate, 0)
	for rows.Next() {
		err := rows.Scan(&id, &name, &description)
		if err != nil {
			return nil, fmt.Errorf("reading fetched rows: %v", err)
		}
		desc := ""
		if description.Valid {
			desc = description.String
		}
		results = append(results, models.ChoreTemplate{
			Id:          id,
			Name:        name,
			Description: desc,
		})
	}
	more := len(results) > limit
	if more {
		results = results[:limit]
	}

	return &models.GetChoreTemplateBatchResult{
		ChoreTemplates: results,
		More:           more,
	}, nil
}
