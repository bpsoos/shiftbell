package chores

import (
	"context"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
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
	Browse(context.Context, *models.BrowseChoresParams) (*models.ChorePage, error)
	CreateManualOneOff(context.Context, *models.CreateManualOneOffParams) (*models.CreateChoreResult, error)
	CreateManualScheduled(context.Context, *models.CreateManualScheduledParams) (*models.CreateChoreResult, error)
	CreateTemplateScheduled(context.Context, *models.CreateTemplateScheduledParams) (*models.CreateChoreResult, error)
	EditOneOff(context.Context, *models.EditOneOffChoreParams) (*models.EditChoreResult, error)
	EditScheduled(context.Context, *models.EditScheduledChoreParams) (*models.EditChoreResult, error)
}

type Normalizer interface {
	NormalizeName(string) (string, bool)
	NormalizeDescription(string) (string, bool)
}
