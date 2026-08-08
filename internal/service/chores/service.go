package chores

import (
	"context"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
)

type Service struct {
	persister  Persister
	normalizer Normalizer
	now        func() time.Time
	timezone   *time.Location
}

type Config struct {
	AppTimezone *time.Location
}

type Deps struct {
	Persister  Persister
	Normalizer Normalizer
	Now        func() time.Time
}

func NewService(deps *Deps, config *Config) *Service {
	return &Service{
		persister:  deps.Persister,
		normalizer: deps.Normalizer,
		now:        deps.Now,
		timezone:   config.AppTimezone,
	}
}

type Persister interface {
	Browse(context.Context, *models.BrowseChoresParams) (*models.ChorePage, error)
	Get(context.Context, int) (*models.ChoreDetails, error)
	CreateManualOneOff(
		context.Context,
		*models.CreateManualOneOffParams,
	) (*models.CreateChoreResult, error)
	CreateTemplateOneOff(
		context.Context,
		*models.CreateTemplateOneOffParams,
	) (*models.CreateChoreResult, error)
	CreateManualScheduled(
		context.Context,
		*models.CreateManualScheduledParams,
	) (*models.CreateChoreResult, error)
	CreateTemplateScheduled(
		context.Context,
		*models.CreateTemplateScheduledParams,
	) (*models.CreateChoreResult, error)
	EditOneOff(
		context.Context,
		*models.EditOneOffChoreParams,
	) (*models.EditChoreResult, error)
	EditScheduled(
		context.Context,
		*models.EditScheduledChoreParams,
	) (*models.EditChoreResult, error)
	Complete(
		context.Context,
		*models.CompleteChoreParams,
	) (*models.CompleteChoreResult, error)
	CorrectCompletion(
		context.Context,
		*models.CorrectCompletionParams,
	) (*models.CorrectCompletionResult, error)
	Delete(context.Context, int) error
}

type Normalizer interface {
	NormalizeName(string) (string, error)
	NormalizeDescription(string) (string, error)
	NormalizeSearch(string) (string, error)
}
