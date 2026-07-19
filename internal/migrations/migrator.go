package migrations

import (
	"embed"
	"errors"
	"fmt"

	"github.com/bpsoos/shiftbell/internal/database"
	"github.com/bpsoos/shiftbell/internal/logging"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Migrator struct{}

func NewMigrator() *Migrator {
	return &Migrator{}
}

//go:embed *.sql
var fs embed.FS

func (*Migrator) Migrate(sqliteFilepath string) error {
	logger := logging.Default()
	driver, err := iofs.New(fs, ".")
	if err != nil {
		return fmt.Errorf("new iofs: %v", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", driver, database.SQLiteMigrationURL(sqliteFilepath))
	if err != nil {
		return fmt.Errorf("new source instance: %v", err)
	}
	defer func() {
		errSource, errDb := m.Close()
		if errSource != nil {
			logger.Error("closing source", "err", errSource)
		}
		if errDb != nil {
			logger.Error("closing database", "err", errDb)
		}
	}()

	m.Log = MigrationLogger{}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("no change")

			return nil
		}
		return fmt.Errorf("migrate up: %v", err)
	}
	return nil
}

type MigrationLogger struct{}

func (MigrationLogger) Printf(format string, v ...any) {
	logging.Default().Info("migrating", "msg", fmt.Sprintf(format, v...))
}

func (MigrationLogger) Verbose() bool {
	return true
}
