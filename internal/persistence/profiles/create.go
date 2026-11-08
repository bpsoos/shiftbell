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
		`INSERT INTO profiles (name, name_key)
		SELECT ?, ? WHERE (SELECT count(*) FROM profiles) < ?`,
		params.Name,
		nameKey(params.Name),
		p.maxProfiles,
	)
	if err != nil {
		if database.IsUniqueConstraintError(err) {
			return nil, models.ErrNameConflict
		}
		return nil, fmt.Errorf("insert profile: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read inserted profile count: %w", err)
	}
	if count == 0 {
		return nil, models.ErrLimitReached
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read inserted profile id: %w", err)
	}
	return &models.Profile{Id: int(id), Name: params.Name}, nil
}
