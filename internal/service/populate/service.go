package populate

import (
	"context"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	profilemodels "github.com/bpsoos/shiftbell/internal/models/profiles"
)

type Service struct {
	profileService ProfileService
	dataPersister  DataPersister
	chorePersister ChorePersister
	now            func() time.Time
	timezone       *time.Location
}

type Deps struct {
	ProfileService ProfileService
	DataPersister  DataPersister
	ChorePersister ChorePersister
	Now            func() time.Time
}

type Config struct {
	AppTimezone *time.Location
}

func NewService(deps *Deps, config *Config) *Service {
	return &Service{
		profileService: deps.ProfileService,
		dataPersister:  deps.DataPersister,
		chorePersister: deps.ChorePersister,
		now:            deps.Now,
		timezone:       config.AppTimezone,
	}
}

type DataPersister interface {
	HasData(context.Context) (bool, error)
}

type ProfileService interface {
	Create(context.Context, *profilemodels.CreateParams) (*profilemodels.Profile, error)
}

type ChorePersister interface {
	SetAssignment(
		context.Context,
		*models.SetAssignmentParams,
	) (*models.ChoreDetails, error)
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
