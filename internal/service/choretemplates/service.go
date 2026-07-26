package choretemplates

import (
	"context"

	"github.com/bpsoos/shiftbell/internal/models"
)

type Service struct{}

type Config struct{}

type Deps struct{}

func NewService(deps *Deps, config *Config) *Service {
	return &Service{}
}

type Persister interface {
	Create(context.Context, *models.CreateChoreTemplateParams) (*models.ChoreTemplate, error)
	Edit(context.Context, *models.EditChoreTemplateParams) (*models.ChoreTemplate, error)
}
