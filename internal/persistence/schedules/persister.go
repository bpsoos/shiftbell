package schedules

import (
	"context"
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/jmoiron/sqlx"
)

type Persister struct {
	db *sqlx.DB
}

type PersisterDeps struct {
	Db *sqlx.DB
}

func NewPersister(deps *PersisterDeps) *Persister {
	return &Persister{db: deps.Db}
}

func (p *Persister) Get(_ context.Context, _ int) (*models.ScheduleDetails, error) {
	return nil, errors.New("get schedule: not implemented")
}

func (p *Persister) Edit(
	_ context.Context,
	_ *models.EditScheduleParams,
) (*models.Schedule, error) {
	return nil, errors.New("edit schedule: not implemented")
}

func (p *Persister) Deactivate(
	_ context.Context,
	_ *models.DeactivateScheduleParams,
) (*models.Schedule, error) {
	return nil, errors.New("deactivate schedule: not implemented")
}
