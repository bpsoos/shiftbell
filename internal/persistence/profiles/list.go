package profiles

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
)

func (p *Persister) List(ctx context.Context) ([]models.Profile, error) {
	profiles := []models.Profile{}
	if err := p.db.SelectContext(
		ctx,
		&profiles,
		`SELECT id, name FROM profiles ORDER BY id`,
	); err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	return profiles, nil
}
