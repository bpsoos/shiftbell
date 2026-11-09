package accounts

import (
	"context"
	"fmt"

	models "github.com/bpsoos/shiftbell/internal/models/accounts"
	"github.com/bpsoos/shiftbell/internal/models/validation"
)

func (s *Service) Create(
	ctx context.Context,
	params *models.CreateParams,
) (*models.Account, error) {
	if params == nil {
		return nil, fmt.Errorf(
			"%w: %w",
			validation.ErrInvalidUsername,
			validation.ErrRequired,
		)
	}
	username, err := s.usernameValidator.Validate(params.Username)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", validation.ErrInvalidUsername, err)
	}
	if err := s.passwordValidator.Validate(params.Password); err != nil {
		return nil, fmt.Errorf("%w: %w", validation.ErrInvalidPassword, err)
	}
	passwordHash, err := s.passwordHasher.Hash(params.Password)
	if err != nil {
		return nil, fmt.Errorf("hash account password: %w", err)
	}
	account, err := s.persister.Create(ctx, &models.CreateValues{
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	return account, nil
}
