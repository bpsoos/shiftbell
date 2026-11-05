package profiles

import (
	"context"
	"fmt"
)

func (p *Persister) HasProfiles(ctx context.Context) (bool, error) {
	var exists bool
	if err := p.db.GetContext(
		ctx,
		&exists,
		`SELECT EXISTS(SELECT 1 FROM profiles)`,
	); err != nil {
		return false, fmt.Errorf("check profiles: %w", err)
	}
	return exists, nil
}
