package accounts

import (
	"context"

	models "github.com/bpsoos/shiftbell/internal/models/accounts"
)

type Persister interface {
	Create(context.Context, *models.CreateValues) (*models.Account, error)
}

type UsernameValidator interface {
	Validate(string) (string, error)
}

type PasswordHasher interface {
	Hash(string) (string, error)
}

type PasswordValidator interface {
	Validate(string) error
}

type Deps struct {
	Persister         Persister
	UsernameValidator UsernameValidator
	PasswordValidator PasswordValidator
	PasswordHasher    PasswordHasher
}

type Service struct {
	persister         Persister
	usernameValidator UsernameValidator
	passwordValidator PasswordValidator
	passwordHasher    PasswordHasher
}

func NewService(deps *Deps) *Service {
	return &Service{
		persister:         deps.Persister,
		usernameValidator: deps.UsernameValidator,
		passwordValidator: deps.PasswordValidator,
		passwordHasher:    deps.PasswordHasher,
	}
}
