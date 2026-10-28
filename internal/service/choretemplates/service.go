package choretemplates

import (
	"context"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
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
	return &Service{
		persister:  deps.Persister,
		normalizer: deps.Normalizer,
		now:        deps.Now,
	}
}

type Persister interface {
	Create(
		context.Context,
		*models.CreateChoreTemplateParams,
	) (*models.ChoreTemplate, error)
	Browse(
		context.Context,
		*models.BrowseChoreTemplatesParams,
	) (*models.ChoreTemplatePage, error)
	Get(context.Context, int) (*models.ChoreTemplateDetails, error)
	ListSchedules(
		context.Context,
		int,
		schedulemodels.ScheduleFilter,
	) ([]models.ScheduleReference, error)
	Edit(context.Context, *models.EditChoreTemplateParams) (*models.ChoreTemplate, error)
	Deactivate(
		context.Context,
		*models.DeactivateChoreTemplateParams,
	) (*models.ChoreTemplate, error)
}

type Normalizer interface {
	NormalizeName(string) (string, error)
	NormalizeDescription(string) (string, error)
	NormalizeSearch(string) (string, error)
}
