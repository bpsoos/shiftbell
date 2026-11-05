package profiles

import (
	"context"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/database"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
)

func (p *Persister) Create(
	ctx context.Context,
	params *models.CreateParams,
) (*models.Profile, error) {
	result, err := p.db.ExecContext(
		ctx,
		`INSERT INTO profiles (name, name_key) VALUES (?, ?)`,
		params.Name,
		nameKey(params.Name),
	)
	if err != nil {
		if database.IsUniqueConstraintError(err) {
			return nil, models.ErrNameConflict
		}
		return nil, fmt.Errorf("insert profile: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read inserted profile id: %w", err)
	}
	return &models.Profile{Id: int(id), Name: params.Name}, nil
}
