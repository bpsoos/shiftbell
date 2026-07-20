package chores

import (
	"fmt"
	"time"
)

func (p *Persister) MarkComplete(id int, completedOn time.Time) error {
	_, err := p.db.Exec(
		`
			update chores
			set is_complete = true,
				completed_on = ?
			where id = ? and is_complete = false
		`,
		completedOn,
		id,
	)
	if err != nil {
		return fmt.Errorf("update chores exec: %v", err)
	}
	return nil
}
