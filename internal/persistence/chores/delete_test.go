package chores_test

import (
	"time"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete chore", func() {
	DescribeTable(
		"deletes a one-off chore",
		func(ctx SpecContext, isComplete bool, completedOn any) {
			db := sqlitetest.NewMigratedDB()
			persister := chorespersistence.NewPersister(
				&chorespersistence.PersisterDeps{Db: db},
			)
			_, err := db.ExecContext(ctx, `
				insert into chores (id, name, is_complete, completed_on, deadline)
				values (42, 'Kitchen', ?, ?, '2020-02-01')
			`, isComplete, completedOn)
			Expect(err).NotTo(HaveOccurred())

			err = persister.Delete(ctx, 42)

			Expect(err).NotTo(HaveOccurred())
			_, err = persister.Get(ctx, 42)
			Expect(err).To(MatchError(choremodels.ErrNotFound))
		},
		Entry("when active", false, nil),
		Entry(
			"when completed",
			true,
			time.Date(2020, time.February, 2, 0, 0, 0, 0, time.UTC),
		),
	)

	It("returns not found when the chore does not exist", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)

		err := persister.Delete(ctx, 42)

		Expect(err).To(MatchError(choremodels.ErrNotFound))
	})

	When("chores belong to a schedule", func() {
		var (
			db        *sqlx.DB
			persister *chorespersistence.Persister
		)

		BeforeEach(func(ctx SpecContext) {
			db = sqlitetest.NewMigratedDB()
			persister = chorespersistence.NewPersister(
				&chorespersistence.PersisterDeps{Db: db},
			)
			_, err := db.ExecContext(
				ctx,
				`insert into chore_templates (id, name) values (1, 'Kitchen')`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`insert into schedules (id, name, chore_template_id, interval_days)
				values (1, 'Weekly kitchen', 1, 7)`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`insert into chores (id, name, is_complete, completed_on, deadline, schedule_id)
				values (41, 'Kitchen', true, '2020-02-02', '2020-02-01', 1),
					(42, 'Kitchen', false, null, '2020-02-09', 1)`,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It(
			"rejects active scheduled deletion and preserves the schedule and its chores",
			func(ctx SpecContext) {
				before, err := persister.Get(ctx, 42)
				Expect(err).NotTo(HaveOccurred())

				err = persister.Delete(ctx, 42)

				Expect(err).To(MatchError(choremodels.ErrActiveScheduledCannotDelete))
				after, err := persister.Get(ctx, 42)
				Expect(err).NotTo(HaveOccurred())
				Expect(after).To(Equal(before))
				var scheduleCount, choreCount int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from schedules where id = 1`).
						Scan(&scheduleCount),
				).To(Succeed())
				Expect(
					db.QueryRowContext(ctx, `select count(*) from chores where schedule_id = 1`).
						Scan(&choreCount),
				).To(Succeed())
				Expect(scheduleCount).To(Equal(1))
				Expect(choreCount).To(Equal(2))
			},
		)

		It(
			"allows completed scheduled deletion without changing the active chore",
			func(ctx SpecContext) {
				active, err := persister.Get(ctx, 42)
				Expect(err).NotTo(HaveOccurred())

				err = persister.Delete(ctx, 41)

				Expect(err).NotTo(HaveOccurred())
				_, err = persister.Get(ctx, 41)
				Expect(err).To(MatchError(choremodels.ErrNotFound))
				remaining, err := persister.Get(ctx, 42)
				Expect(err).NotTo(HaveOccurred())
				Expect(remaining).To(Equal(active))
				var scheduleCount int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from schedules where id = 1`).
						Scan(&scheduleCount),
				).To(Succeed())
				Expect(scheduleCount).To(Equal(1))
			},
		)

		It(
			"keeps the successor when its completed predecessor is deleted",
			func(ctx SpecContext) {
				_, err := db.ExecContext(
					ctx,
					`update chores set predecessor_id = 41 where id = 42`,
				)
				Expect(err).NotTo(HaveOccurred())
				err = persister.Delete(ctx, 41)
				Expect(err).NotTo(HaveOccurred())
				successor, err := persister.Get(ctx, 42)
				Expect(err).NotTo(HaveOccurred())
				Expect(successor.PredecessorId).To(BeNil())
				Expect(successor.Status).To(Equal(choremodels.ChoreStatusActive))
				Expect(*successor.ScheduleId).To(Equal(1))
			},
		)
	})
})
