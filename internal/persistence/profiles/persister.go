package profiles

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

type Deps struct {
	Db *sqlx.DB
}

type Config struct {
	MaxProfiles int
}

type Persister struct {
	db          *sqlx.DB
	maxProfiles int
}

func NewPersister(deps *Deps, config *Config) *Persister {
	return &Persister{db: deps.Db, maxProfiles: config.MaxProfiles}
}

func nameKey(name string) string {
	return norm.NFC.String(cases.Fold().String(name))
}
