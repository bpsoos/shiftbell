package choretypes

import (
	"fmt"

	"github.com/bpsoos/shiftbell/internal/models"
)

func (p *Persister) Get(id int) (*models.ChoreType, error) {
	row := p.db.QueryRow(
		`
			select name, description
			from chore_types
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
		return nil, fmt.Errorf("db query selecting chores: %v", err)
	}

	return &models.ChoreType{
		Id:          id,
		Name:        name,
		Description: description,
	}, nil
}
