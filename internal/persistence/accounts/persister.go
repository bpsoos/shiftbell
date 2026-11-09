package accounts

import "github.com/jmoiron/sqlx"

type Deps struct {
	Db *sqlx.DB
}

type Persister struct {
	db *sqlx.DB
}

func NewPersister(deps *Deps) *Persister {
	return &Persister{db: deps.Db}
}
