package populate

import (
	"context"
	"fmt"
)

func (p *Persister) HasData(ctx context.Context) (bool, error) {
	var hasData bool
	err := p.db.GetContext(ctx, &hasData, `SELECT
		EXISTS (SELECT 1 FROM chores) OR
		EXISTS (SELECT 1 FROM chore_templates) OR
		EXISTS (SELECT 1 FROM schedules)`)
	if err != nil {
		return false, fmt.Errorf("check for existing application data: %w", err)
	}
	return hasData, nil
}
