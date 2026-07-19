package sqlitetest

import (
	"path/filepath"

	"github.com/bpsoos/shiftbell/internal/database"
	"github.com/bpsoos/shiftbell/internal/migrations"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	_ "modernc.org/sqlite"
)

func NewMigratedDB() *sqlx.DB {
	GinkgoHelper()

	databaseFilepath := filepath.Join(GinkgoT().TempDir(), "test.db")
	Expect(migrations.NewMigrator().Migrate(databaseFilepath)).To(Succeed())

	db, err := sqlx.Connect("sqlite", database.SQLiteDSN(databaseFilepath))
	Expect(err).NotTo(HaveOccurred())
	db.SetMaxOpenConns(1)

	DeferCleanup(func() {
		Expect(db.Close()).To(Succeed())
	})

	return db
}
