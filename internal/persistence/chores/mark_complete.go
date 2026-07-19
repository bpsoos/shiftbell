package chores

import (
	"fmt"
	"time"
)

func (p *Persister) MarkComplete(id int, completedAt time.Time) error {
	_, err := p.db.Exec(
		`
			update chores
			set is_complete = true,
				completed_at = ?
			where id = ? and is_complete = false
		`,
		completedAt,
		id,
	)
	if err != nil {
		return fmt.Errorf("update chores exec: %v", err)
	}
	return nil
}
