package choretypes

import "github.com/jmoiron/sqlx"

type Persister struct {
	db *sqlx.DB
}

type PersisterDeps struct {
	Db *sqlx.DB
}

func NewChoreTypePersister(deps *PersisterDeps) *Persister {
	return &Persister{
		db: deps.Db,
	}
}
