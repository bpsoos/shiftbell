package populate

import (
	"context"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
)

type Service struct {
	dataPersister  DataPersister
	chorePersister ChorePersister
	now            func() time.Time
	timezone       *time.Location
}

type Deps struct {
	DataPersister  DataPersister
	ChorePersister ChorePersister
	Now            func() time.Time
}

type Config struct {
	AppTimezone *time.Location
}

func NewService(deps *Deps, config *Config) *Service {
	return &Service{
		dataPersister:  deps.DataPersister,
		chorePersister: deps.ChorePersister,
		now:            deps.Now,
		timezone:       config.AppTimezone,
	}
}

type DataPersister interface {
	HasData(context.Context) (bool, error)
}

type ChorePersister interface {
	CreateManualOneOff(
		context.Context,
		*models.CreateManualOneOffParams,
	) (*models.CreateChoreResult, error)
	CreateManualScheduled(
		context.Context,
		*models.CreateManualScheduledParams,
	) (*models.CreateChoreResult, error)
	MarkComplete(int, time.Time) error
}
