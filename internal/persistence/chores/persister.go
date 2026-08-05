package chores

import (
	"context"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/jmoiron/sqlx"
)

type Persister struct {
	db *sqlx.DB
}

type PersisterDeps struct {
	Db *sqlx.DB
}

func NewPersister(deps *PersisterDeps) *Persister {
	return &Persister{
		db: deps.Db,
	}
}

func (p *Persister) CreateManualScheduled(
	_ context.Context,
	_ *models.CreateManualScheduledParams,
) (*models.CreateChoreResult, error) {
	return nil, nil
}

func (p *Persister) CreateTemplateScheduled(
	_ context.Context,
	_ *models.CreateTemplateScheduledParams,
) (*models.CreateChoreResult, error) {
	return nil, nil
}

func (p *Persister) EditScheduled(
	_ context.Context,
	_ *models.EditScheduledChoreParams,
) (*models.EditChoreResult, error) {
	return nil, nil
}
