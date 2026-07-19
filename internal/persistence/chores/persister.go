package chores

import "github.com/jmoiron/sqlx"

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
