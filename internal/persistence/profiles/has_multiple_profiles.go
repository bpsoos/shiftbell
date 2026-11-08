package profiles

import (
	"context"
	"fmt"
)

func (p *Persister) HasMultipleProfiles(ctx context.Context) (bool, error) {
	var multiple bool
	if err := p.db.GetContext(
		ctx,
		&multiple,
		`SELECT EXISTS(SELECT 1 FROM profiles LIMIT 1 OFFSET 1)`,
	); err != nil {
		return false, fmt.Errorf("check multiple profiles: %w", err)
	}
	return multiple, nil
}
