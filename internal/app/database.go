package app

import (
	"fmt"

	"github.com/bpsoos/shiftbell/internal/database"
	"github.com/bpsoos/shiftbell/internal/migrations"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func OpenDatabase(path string) (*sqlx.DB, error) {
	if err := migrations.NewMigrator().Migrate(path); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	db, err := sqlx.Connect("sqlite", database.SQLiteDSN(path))
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	db.SetMaxOpenConns(1)
	return db, nil
}
