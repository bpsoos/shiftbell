package chores_test

import (
	"time"

	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MarkComplete", func() {
	var (
		db          *sqlx.DB
		persister   *chorespersistence.Persister
		completedAt time.Time
		deadline    time.Time
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = chorespersistence.NewPersister(&chorespersistence.PersisterDeps{Db: db})
		completedAt = time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
		deadline = time.Date(2026, time.July, 20, 0, 0, 0, 0, time.UTC)
	})

	Context("with an incomplete chore", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chores (id, name, description, is_complete, deadline) values (?, ?, ?, ?, ?)`,
				1,
				"Laundry",
				"Wash and fold clothes",
				false,
				deadline,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("persists the completed status and timestamp", func() {
			Expect(persister.MarkComplete(1, completedAt)).To(Succeed())

			var isComplete bool
			var persistedCompletedAt time.Time
			Expect(db.QueryRow(
				`select is_complete, completed_at from chores where id = ?`,
				1,
			).Scan(&isComplete, &persistedCompletedAt)).To(Succeed())
			Expect(isComplete).To(BeTrue())
			Expect(persistedCompletedAt).To(Equal(completedAt))
		})
	})

	Context("with an already complete chore", func() {
		var originalCompletedAt time.Time

		BeforeEach(func() {
			originalCompletedAt = time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)
			_, err := db.Exec(
				`insert into chores (id, name, description, is_complete, completed_at, deadline) values (?, ?, ?, ?, ?, ?)`,
				1,
				"Laundry",
				"Wash and fold clothes",
				true,
				originalCompletedAt,
				deadline,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("keeps the original completion timestamp", func() {
			Expect(persister.MarkComplete(1, completedAt)).To(Succeed())

			var persistedCompletedAt time.Time
			Expect(db.QueryRow(
				`select completed_at from chores where id = ?`,
				1,
			).Scan(&persistedCompletedAt)).To(Succeed())
			Expect(persistedCompletedAt).To(Equal(originalCompletedAt))
		})
	})

	Context("with an unknown chore", func() {
		It("succeeds without changing any rows", func() {
			Expect(persister.MarkComplete(99, completedAt)).To(Succeed())

			var count int
			Expect(db.QueryRow(`select count(*) from chores where is_complete = true`).Scan(&count)).To(Succeed())
			Expect(count).To(Equal(0))
		})
	})

	Context("when the database update fails", func() {
		It("returns an error", func() {
			Expect(db.Close()).To(Succeed())

			err := persister.MarkComplete(1, completedAt)

			Expect(err).To(MatchError(ContainSubstring("update chores exec")))
		})
	})
})
