package database_test

import (
	"net/url"
	"testing"

	"github.com/bpsoos/shiftbell/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDatabase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Database Suite")
}

var _ = Describe("SQLiteDSN", func() {
	It("builds a file DSN with foreign keys enabled", func() {
		dsn, err := url.Parse(database.SQLiteDSN("/tmp/shift bell.db"))

		Expect(err).NotTo(HaveOccurred())
		Expect(dsn.Scheme).To(Equal("file"))
		Expect(dsn.Path).To(Equal("/tmp/shift bell.db"))
		Expect(dsn.Query().Get("_pragma")).To(Equal("foreign_keys(1)"))
	})
})

var _ = Describe("SQLiteMigrationURL", func() {
	It("builds a migration URL with foreign keys enabled", func() {
		dsn, err := url.Parse(database.SQLiteMigrationURL("/tmp/shift bell.db"))

		Expect(err).NotTo(HaveOccurred())
		Expect(dsn.Scheme).To(Equal("sqlite"))
		Expect(dsn.Path).To(Equal("/tmp/shift bell.db"))
		Expect(dsn.Query().Get("_pragma")).To(Equal("foreign_keys(1)"))
	})
})
