package choretemplates

import (
	"context"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
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
	Browse(context.Context, *models.BrowseChoreTemplatesParams) (*models.ChoreTemplatePage, error)
	Get(context.Context, int) (*models.ChoreTemplateDetails, error)
	Edit(context.Context, *models.EditChoreTemplateParams) (*models.ChoreTemplate, error)
}

type Normalizer interface {
	NormalizeName(string) (string, error)
	NormalizeDescription(string) (string, error)
	NormalizeSearch(string) (string, error)
}
