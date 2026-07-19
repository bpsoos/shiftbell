package chores

import (
	"fmt"
	"time"
)

func (p *Persister) MarkComplete(id int, completedAt time.Time) error {
	_, err := p.db.Exec(
		`
			with candidate as (
				select id
				from chores
				where id = $1 and is_complete = false
			)
			update chores c
			set is_complete = true,
				completed_at = $2
			from candidate
			where c.id = candidate.id
		`,
		id,
		completedAt,
	)
	if err != nil {
		return fmt.Errorf("update chores exec: %v", err)
	}
	return nil
}
