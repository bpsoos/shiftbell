package choretemplates_test

import (
	"database/sql"

	choretemplatespersistence "github.com/bpsoos/shiftbell/internal/persistence/choretemplates"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {
	var (
		db        *sqlx.DB
		persister *choretemplatespersistence.Persister
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = choretemplatespersistence.NewChoreTemplatePersister(&choretemplatespersistence.PersisterDeps{Db: db})
	})

	Context("with no description", func() {
		It("persists a null description", func() {
			Expect(persister.Create("Laundry", "")).To(Succeed())

			var name string
			var description sql.NullString
			Expect(db.QueryRow(`select name, description from chore_templates`).Scan(&name, &description)).To(Succeed())
			Expect(name).To(Equal("Laundry"))
			Expect(description.Valid).To(BeFalse())
		})
	})

	Context("with one description", func() {
		It("persists the name and description", func() {
			Expect(persister.Create("Laundry", "Wash and fold clothes")).To(Succeed())

			var name string
			var description string
			Expect(db.QueryRow(`select name, description from chore_templates`).Scan(&name, &description)).To(Succeed())
			Expect(name).To(Equal("Laundry"))
			Expect(description).To(Equal("Wash and fold clothes"))
		})
	})

	Context("with many chore templates", func() {
		It("persists every chore template", func() {
			Expect(persister.Create("Laundry", "Wash and fold clothes")).To(Succeed())
			Expect(persister.Create("Dishes", "Load the dishwasher")).To(Succeed())

			var count int
			Expect(db.QueryRow(`select count(*) from chore_templates where name in (?, ?)`, "Laundry", "Dishes").Scan(&count)).To(Succeed())
			Expect(count).To(Equal(2))
		})
	})
})
