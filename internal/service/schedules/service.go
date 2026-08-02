package schedules

import (
	"context"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
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
	Browse(context.Context, *models.BrowseSchedulesParams) (*models.SchedulePage, error)
	Get(context.Context, int) (*models.ScheduleDetails, error)
	Edit(context.Context, *models.EditScheduleParams) (*models.Schedule, error)
	Deactivate(
		context.Context,
		*models.DeactivateScheduleParams,
	) (*models.Schedule, error)
}

type Normalizer interface {
	NormalizeName(string) (string, error)
	NormalizeSearch(string) (string, error)
}
