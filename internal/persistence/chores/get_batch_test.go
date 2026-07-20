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
		db        *sqlx.DB
		persister *chorespersistence.Persister
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = chorespersistence.NewPersister(&chorespersistence.PersisterDeps{Db: db})
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
				`insert into chores (name, description, is_complete, deadline) values (?, ?, ?, ?)`,
				"First chore",
				"first description",
				false,
				time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the chore with more=false", func() {
			result, err := persister.GetBatch(0, 1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Chores).To(HaveLen(1))
			Expect(result.Chores[0].Name).To(Equal("First chore"))
			Expect(result.Chores[0].Description).To(Equal("first description"))
			Expect(result.More).To(BeFalse())
		})
	})

	Context("with many chores", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chores (name, description, is_complete, deadline) values
					(?, ?, ?, ?),
					(?, ?, ?, ?),
					(?, ?, ?, ?)`,
				"First chore",
				"first description",
				false,
				time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				"Second chore",
				"second description",
				false,
				time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC),
				"Third chore",
				"third description",
				false,
				time.Date(2026, time.January, 3, 0, 0, 0, 0, time.UTC),
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the requested page with more=true", func() {
			result, err := persister.GetBatch(1, 1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Chores).To(HaveLen(1))
			Expect(result.Chores[0].Name).To(Equal("Second chore"))
			Expect(result.Chores[0].Description).To(Equal("second description"))
			Expect(result.More).To(BeTrue())
		})

		It("returns the first page with more=true", func() {
			result, err := persister.GetBatch(0, 2)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Chores).To(HaveLen(2))
			Expect(result.Chores[0].Name).To(Equal("First chore"))
			Expect(result.Chores[1].Name).To(Equal("Second chore"))
			Expect(result.Chores[0].Description).To(Equal("first description"))
			Expect(result.Chores[1].Description).To(Equal("second description"))
			Expect(result.More).To(BeTrue())
		})

		It("returns the final page with more=false", func() {
			result, err := persister.GetBatch(2, 2)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Chores).To(HaveLen(1))
			Expect(result.Chores[0].Name).To(Equal("Third chore"))
			Expect(result.Chores[0].Description).To(Equal("third description"))
			Expect(result.More).To(BeFalse())
		})

		It("returns an empty page beyond the end", func() {
			result, err := persister.GetBatch(3, 2)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Chores).To(BeEmpty())
			Expect(result.More).To(BeFalse())
		})
	})
})
