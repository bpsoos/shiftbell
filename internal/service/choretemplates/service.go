package choretemplates

import (
	"context"

	"github.com/bpsoos/shiftbell/internal/models"
)

type Service struct {
	persister  Persister
	normalizer Normalizer
}

type Config struct{}

type Deps struct {
	Persister  Persister
	Normalizer Normalizer
}

func NewService(deps *Deps, config *Config) *Service {
	return &Service{
		persister:  deps.Persister,
		normalizer: deps.Normalizer,
	}
}

type Persister interface {
	Create(context.Context, *models.CreateChoreTemplateParams) (*models.ChoreTemplate, error)
	Edit(context.Context, *models.EditChoreTemplateParams) (*models.ChoreTemplate, error)
}

type Normalizer interface {
	NormalizeName(string) (string, bool)
	NormalizeDescription(string) (string, bool)
}
