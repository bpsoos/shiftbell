package accounts

import (
	"context"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/database"
	models "github.com/bpsoos/shiftbell/internal/models/accounts"
)

func (p *Persister) Create(
	ctx context.Context,
	values *models.CreateValues,
) (*models.Account, error) {
	result, err := p.db.ExecContext(
		ctx,
		`INSERT INTO accounts (username, password_hash) VALUES (?, ?)`,
		values.Username,
		values.PasswordHash,
	)
	if err != nil {
		if database.IsUniqueConstraintError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, fmt.Errorf("insert account: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read inserted account id: %w", err)
	}
	return &models.Account{Id: int(id), Username: values.Username}, nil
}
