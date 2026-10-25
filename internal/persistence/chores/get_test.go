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

		It("returns the active chore", func(ctx SpecContext) {
			result, err := persister.Get(ctx, 1)

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

		It("returns the completed chore", func(ctx SpecContext) {
			result, err := persister.Get(ctx, 1)

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

	It("returns the schedule association of a scheduled chore", func(ctx SpecContext) {
		_, err := db.ExecContext(
			ctx,
			`insert into chore_templates (id, name) values (1, 'Laundry')`,
		)
		Expect(err).NotTo(HaveOccurred())
		_, err = db.ExecContext(
			ctx,
			`insert into schedules (id, name, chore_template_id, interval_days) values (1, 'Weekly laundry', 1, 7)`,
		)
		Expect(err).NotTo(HaveOccurred())
		deadline := time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC)
		_, err = db.ExecContext(
			ctx,
			`insert into chores (id, name, is_complete, deadline, schedule_id) values (1, 'Laundry', false, ?, 1)`,
			deadline,
		)
		Expect(err).NotTo(HaveOccurred())
		chore, err := persister.Get(ctx, 1)
		Expect(err).NotTo(HaveOccurred())
		scheduleID := 1
		Expect(chore).To(Equal(&models.Chore{
			Id:         1,
			ScheduleId: &scheduleID,
			Name:       "Laundry",
			Status:     models.ChoreStatusActive,
			Deadline:   deadline,
		}))
	})
})
