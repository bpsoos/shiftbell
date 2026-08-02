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
		completedOn time.Time
		deadline    time.Time
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)
		completedOn = time.Date(2026, time.July, 19, 0, 0, 0, 0, time.UTC)
		deadline = time.Date(2026, time.July, 20, 0, 0, 0, 0, time.UTC)
	})

	Context("with an active chore", func() {
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

		It("persists the completed status and completion date", func() {
			Expect(persister.MarkComplete(1, completedOn)).To(Succeed())

			var isCompleted bool
			var persistedCompletedOn time.Time
			Expect(db.QueryRow(
				`select is_complete, completed_on from chores where id = ?`,
				1,
			).Scan(&isCompleted, &persistedCompletedOn)).To(Succeed())
			Expect(isCompleted).To(BeTrue())
			Expect(persistedCompletedOn).To(Equal(completedOn))
		})
	})

	Context("with an already completed chore", func() {
		var originalCompletedOn time.Time

		BeforeEach(func() {
			originalCompletedOn = time.Date(2026, time.July, 18, 0, 0, 0, 0, time.UTC)
			_, err := db.Exec(
				`insert into chores (id, name, description, is_complete, completed_on, deadline) values (?, ?, ?, ?, ?, ?)`,
				1,
				"Laundry",
				"Wash and fold clothes",
				true,
				originalCompletedOn,
				deadline,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("keeps the original completion date", func() {
			Expect(persister.MarkComplete(1, completedOn)).To(Succeed())

			var persistedCompletedOn time.Time
			Expect(db.QueryRow(
				`select completed_on from chores where id = ?`,
				1,
			).Scan(&persistedCompletedOn)).To(Succeed())
			Expect(persistedCompletedOn).To(Equal(originalCompletedOn))
		})
	})

	Context("with an unknown chore", func() {
		It("succeeds without changing any rows", func() {
			Expect(persister.MarkComplete(99, completedOn)).To(Succeed())

			var count int
			Expect(
				db.QueryRow(`select count(*) from chores where is_complete = true`).
					Scan(&count),
			).To(Succeed())
			Expect(count).To(Equal(0))
		})
	})

	Context("when the database update fails", func() {
		It("returns an error", func() {
			Expect(db.Close()).To(Succeed())

			err := persister.MarkComplete(1, completedOn)

			Expect(err).To(MatchError(ContainSubstring("update chores exec")))
		})
	})
})
