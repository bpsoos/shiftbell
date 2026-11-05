package profiles

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

type Deps struct {
	Db *sqlx.DB
}

type Persister struct {
	db *sqlx.DB
}

func NewPersister(deps *Deps) *Persister {
	return &Persister{db: deps.Db}
}

func nameKey(name string) string {
	return norm.NFC.String(cases.Fold().String(name))
}
