package chores_test

import (
	"time"

	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetBatch", func() {
	var (
		db              *sqlx.DB
		persister       *chorespersistence.Persister
		lastCompletedAt time.Time
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = chorespersistence.NewPersister(&chorespersistence.PersisterDeps{Db: db})
		lastCompletedAt = time.Date(2026, time.January, 1, 8, 0, 0, 0, time.UTC)
	})

	Context("with no chores", func() {
		It("returns an empty batch", func() {
			result, err := persister.GetBatch(0, 1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Chores).To(BeEmpty())
			Expect(result.More).To(BeFalse())
		})
	})

	Context("with one chore", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chores (name, description, last_completed_at, is_complete, deadline) values (?, ?, ?, ?, ?)`,
				"first",
				"first",
				lastCompletedAt,
				false,
				time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the chore with more=false", func() {
			result, err := persister.GetBatch(0, 1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Chores).To(HaveLen(1))
			Expect(result.Chores[0].Description).To(Equal("first"))
			Expect(result.More).To(BeFalse())
		})
	})

	Context("with many chores", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chores (name, description, last_completed_at, is_complete, deadline) values
					(?, ?, ?, ?, ?),
					(?, ?, ?, ?, ?),
					(?, ?, ?, ?, ?)`,
				"first",
				"first",
				lastCompletedAt,
				false,
				time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				"second",
				"second",
				lastCompletedAt,
				false,
				time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC),
				"third",
				"third",
				lastCompletedAt,
				false,
				time.Date(2026, time.January, 3, 0, 0, 0, 0, time.UTC),
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the requested page with more=true", func() {
			result, err := persister.GetBatch(1, 1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Chores).To(HaveLen(1))
			Expect(result.Chores[0].Description).To(Equal("second"))
			Expect(result.More).To(BeTrue())
		})
	})
})
