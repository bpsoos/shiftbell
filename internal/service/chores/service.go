package chores

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
	Browse(context.Context, *models.BrowseChoresParams) (*models.ChorePage, error)
	CreateManualOneOff(context.Context, *models.CreateManualOneOffInput) (*models.CreateChoreResult, error)
	CreateManualScheduled(context.Context, *models.CreateManualScheduledInput) (*models.CreateChoreResult, error)
	CreateTemplateScheduled(context.Context, *models.CreateTemplateScheduledInput) (*models.CreateChoreResult, error)
}

type Normalizer interface {
	NormalizeName(string) (string, bool)
	NormalizeDescription(string) (string, bool)
}
