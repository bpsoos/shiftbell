package choretemplates

import (
	"context"
	"errors"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/jmoiron/sqlx"
)

type Persister struct {
	db *sqlx.DB
}

type PersisterDeps struct {
	Db *sqlx.DB
}

func NewChoreTemplatePersister(deps *PersisterDeps) *Persister {
	return &Persister{
		db: deps.Db,
	}
}

func (p *Persister) Edit(
	context.Context,
	*models.EditChoreTemplateParams,
) (*models.ChoreTemplate, error) {
	return nil, errors.ErrUnsupported
}

func (p *Persister) Deactivate(
	context.Context,
	*models.DeactivateChoreTemplateParams,
) (*models.ChoreTemplate, error) {
	return nil, errors.ErrUnsupported
}
