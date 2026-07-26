package choretemplates_test

import (
	"github.com/bpsoos/shiftbell/internal/models"
	choretemplatespersistence "github.com/bpsoos/shiftbell/internal/persistence/choretemplates"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get", func() {
	var (
		db        *sqlx.DB
		persister *choretemplatespersistence.Persister
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = choretemplatespersistence.NewChoreTemplatePersister(&choretemplatespersistence.PersisterDeps{Db: db})
	})

	Context("with no chore templates", func() {
		It("returns an error", func() {
			result, err := persister.Get(1)

			Expect(result).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Context("with one chore template", func() {
		BeforeEach(func() {
			_, err := db.Exec(`insert into chore_templates (id, name, description) values (?, ?, ?)`, 1, "Laundry", "Wash and fold clothes")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the chore template", func() {
			result, err := persister.Get(1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(&models.ChoreTemplate{Id: 1, Name: "Laundry", Description: "Wash and fold clothes"}))
		})
	})

	Context("with a chore template without a description", func() {
		BeforeEach(func() {
			_, err := db.Exec(`insert into chore_templates (id, name, description) values (?, ?, ?)`, 1, "Laundry", nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an empty description", func() {
			result, err := persister.Get(1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(&models.ChoreTemplate{Id: 1, Name: "Laundry", Description: ""}))
		})
	})

	Context("with many chore templates", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chore_templates (id, name, description) values
					(?, ?, ?),
					(?, ?, ?)`,
				1, "Laundry", "Wash and fold clothes",
				2, "Dishes", "Load the dishwasher",
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the requested chore template", func() {
			result, err := persister.Get(2)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(&models.ChoreTemplate{Id: 2, Name: "Dishes", Description: "Load the dishwasher"}))
		})
	})
})
