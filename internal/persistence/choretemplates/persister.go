package choretemplates

import "github.com/jmoiron/sqlx"

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
