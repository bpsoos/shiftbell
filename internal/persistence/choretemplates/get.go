package choretemplates

import (
	"fmt"

	"github.com/bpsoos/shiftbell/internal/models"
)

func (p *Persister) Get(id int) (*models.ChoreTemplate, error) {
	row := p.db.QueryRow(
		`
			select name, description
			from chore_templates
			where id = ?
		`,
		id,
	)
	var (
		name        string
		description string
	)
	err := row.Scan(&name, &description)
	if err != nil {
		return nil, fmt.Errorf("db query selecting chore templates: %v", err)
	}

	return &models.ChoreTemplate{
		Id:          id,
		Name:        name,
		Description: description,
	}, nil
}
