package choretemplates

import (
	"context"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
)

type Service struct {
	persister  Persister
	normalizer Normalizer
	now        func() time.Time
}

type Config struct{}

type Deps struct {
	Persister  Persister
	Normalizer Normalizer
	Now        func() time.Time
}

func NewService(deps *Deps, config *Config) *Service {
	now := deps.Now
	if now == nil {
		now = time.Now
	}

	return &Service{
		persister:  deps.Persister,
		normalizer: deps.Normalizer,
		now:        now,
	}
}

type Persister interface {
	Create(context.Context, *models.CreateChoreTemplateParams) (*models.ChoreTemplate, error)
	Browse(context.Context, *models.BrowseChoreTemplatesParams) (*models.ChoreTemplatePage, error)
	Get(context.Context, int) (*models.ChoreTemplateDetails, error)
	Edit(context.Context, *models.EditChoreTemplateParams) (*models.ChoreTemplate, error)
	Deactivate(context.Context, *models.DeactivateChoreTemplateParams) (*models.ChoreTemplate, error)
}

type Normalizer interface {
	NormalizeName(string) (string, error)
	NormalizeDescription(string) (string, error)
	NormalizeSearch(string) (string, error)
}
