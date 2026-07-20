package choretemplates

import (
	"database/sql"
	"fmt"
)

func (p *Persister) Create(name string, description string) error {
	var sqlDesc sql.NullString
	if description != "" {
		sqlDesc = sql.NullString{
			String: description,
			Valid:  true,
		}
	}
	_, err := p.db.NamedExec(`
		insert into chore_templates (name, description)
		values (:name, :description)
	`, map[string]any{
		"description": sqlDesc,
		"name":        name,
	})

	if err != nil {
		return fmt.Errorf("db exec inserting chore templates: %v", err)
	}

	return nil
}
