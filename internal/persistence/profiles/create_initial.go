package profiles

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
)

func (p *Persister) CreateInitial(
	ctx context.Context,
	params *models.CreateParams,
) (*models.Profile, error) {
	result, err := p.db.ExecContext(ctx, `
		INSERT INTO profiles (name, name_key)
		SELECT ?, ? WHERE NOT EXISTS (SELECT 1 FROM profiles)`, params.Name, nameKey(params.Name))
	if err != nil {
		return nil, fmt.Errorf("insert initial profile: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read initial profile count: %w", err)
	}
	if count == 0 {
		return nil, models.ErrAlreadyInitialized
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read initial profile id: %w", err)
	}
	return &models.Profile{Id: int(id), Name: params.Name}, nil
}
