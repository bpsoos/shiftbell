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

func (p *Persister) Deactivate(
	_ context.Context,
	_ *models.DeactivateScheduleParams,
) (*models.Schedule, error) {
	return nil, errors.New("deactivate schedule: not implemented")
}
