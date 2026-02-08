package migrations

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Migrator struct{}

func NewMigrator() *Migrator {
	return &Migrator{}
}

//go:embed *.sql
var fs embed.FS

func (*Migrator) Migrate(postgresURL string) {
	driver, err := iofs.New(fs, ".")
	if err != nil {
		panic(err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", driver, postgresURL)
	if err != nil {
		panic(err)
	}
	m.Log = MigrationLogger{}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no change")

			return
		}
		panic(err)
	}
}

type MigrationLogger struct{}

func (MigrationLogger) Printf(format string, v ...any) {
	slog.Info("migrating", "msg", fmt.Sprintf(format, v...))
}

func (MigrationLogger) Verbose() bool {
	return true
}
