package chores_test

import (
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get", func() {
	var (
		db        *sqlx.DB
		persister *chorespersistence.Persister
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)
	})

	Context("with an active chore without a description", func() {
		var deadline time.Time

		BeforeEach(func() {
			deadline = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
			_, err := db.Exec(
				`insert into chores (id, name, description, is_complete, deadline) values (?, ?, ?, ?, ?)`,
				1,
				"First chore",
				nil,
				false,
				deadline,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the active chore", func() {
			result, err := persister.Get(1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(&models.Chore{
				Id:          1,
				Name:        "First chore",
				Status:      models.ChoreStatusActive,
				Description: "",
				Deadline:    deadline,
			}))
		})
	})

	Context("with a completed chore", func() {
		var (
			completedOn time.Time
			deadline    time.Time
		)

		BeforeEach(func() {
			completedOn = time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)
			deadline = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
			_, err := db.Exec(
				`insert into chores (id, name, description, is_complete, completed_on, deadline) values (?, ?, ?, ?, ?, ?)`,
				1,
				"First chore",
				"First description",
				true,
				completedOn,
				deadline,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the completed chore", func() {
			result, err := persister.Get(1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(&models.Chore{
				Id:          1,
				Name:        "First chore",
				Status:      models.ChoreStatusCompleted,
				Description: "First description",
				Deadline:    deadline,
				CompletedOn: completedOn,
			}))
		})
	})
})
