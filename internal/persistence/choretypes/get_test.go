package choretypes_test

import (
	"github.com/bpsoos/shiftbell/internal/models"
	choretypespersistence "github.com/bpsoos/shiftbell/internal/persistence/choretypes"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get", func() {
	var (
		db        *sqlx.DB
		persister *choretypespersistence.Persister
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = choretypespersistence.NewChoreTypePersister(&choretypespersistence.PersisterDeps{Db: db})
	})

	Context("with no chore types", func() {
		It("returns an error", func() {
			result, err := persister.Get(1)

			Expect(result).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Context("with one chore type", func() {
		BeforeEach(func() {
			_, err := db.Exec(`insert into chore_types (id, name, description) values (?, ?, ?)`, 1, "Laundry", "Wash and fold clothes")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the chore type", func() {
			result, err := persister.Get(1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(&models.ChoreType{Id: 1, Name: "Laundry", Description: "Wash and fold clothes"}))
		})
	})

	Context("with many chore types", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chore_types (id, name, description) values
					(?, ?, ?),
					(?, ?, ?)`,
				1, "Laundry", "Wash and fold clothes",
				2, "Dishes", "Load the dishwasher",
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the requested chore type", func() {
			result, err := persister.Get(2)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(&models.ChoreType{Id: 2, Name: "Dishes", Description: "Load the dishwasher"}))
		})
	})
})
