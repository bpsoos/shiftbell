package choretypes_test

import (
	choretypespersistence "github.com/bpsoos/shiftbell/internal/persistence/choretypes"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete", func() {
	var (
		db        *sqlx.DB
		persister *choretypespersistence.Persister
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = choretypespersistence.NewChoreTypePersister(&choretypespersistence.PersisterDeps{Db: db})
	})

	Context("with no chore types", func() {
		It("succeeds without deleting anything", func() {
			Expect(persister.Delete(1)).To(Succeed())

			var count int
			Expect(db.QueryRow(`select count(*) from chore_types`).Scan(&count)).To(Succeed())
			Expect(count).To(Equal(0))
		})
	})

	Context("with one chore type", func() {
		BeforeEach(func() {
			_, err := db.Exec(`insert into chore_types (id, name, description) values (?, ?, ?)`, 1, "Laundry", "Wash and fold clothes")
			Expect(err).NotTo(HaveOccurred())
		})

		It("deletes the chore type", func() {
			Expect(persister.Delete(1)).To(Succeed())

			var count int
			Expect(db.QueryRow(`select count(*) from chore_types`).Scan(&count)).To(Succeed())
			Expect(count).To(Equal(0))
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

		It("deletes only the requested chore type", func() {
			Expect(persister.Delete(1)).To(Succeed())

			var count int
			Expect(db.QueryRow(`select count(*) from chore_types`).Scan(&count)).To(Succeed())
			Expect(count).To(Equal(1))

			var name string
			Expect(db.QueryRow(`select name from chore_types`).Scan(&name)).To(Succeed())
			Expect(name).To(Equal("Dishes"))
		})
	})
})
