package chores_test

import (
	"time"

	"github.com/bpsoos/shiftbell/internal/database"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Complete chore", func() {
	It(
		"marks a one-off chore completed and returns its updated state",
		func(ctx SpecContext) {
			db := sqlitetest.NewMigratedDB()
			persister := chorespersistence.NewPersister(
				&chorespersistence.PersisterDeps{Db: db},
			)
			deadline := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
			completedOn := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			_, err := db.ExecContext(ctx, `
			insert into chores (id, name, description, is_complete, completed_on, deadline)
			values (42, 'Kitchen', 'Clean counters', false, null, ?)
		`, deadline)
			Expect(err).NotTo(HaveOccurred())

			result, err := persister.Complete(ctx, &choremodels.CompleteChoreParams{
				Id:          42,
				CompletedOn: completedOn,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(&choremodels.CompleteChoreResult{
				Chore: &choremodels.Chore{
					Id:          42,
					Status:      choremodels.ChoreStatusCompleted,
					Name:        "Kitchen",
					Description: "Clean counters",
					Deadline:    deadline,
					CompletedOn: completedOn,
				},
			}))
		},
	)

	It("preserves the original completion date when repeated", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)
		deadline := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
		originalDate := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
		differentDate := time.Date(2020, time.February, 4, 0, 0, 0, 0, time.UTC)
		_, err := db.ExecContext(ctx, `
			insert into chores (id, name, description, is_complete, completed_on, deadline)
			values (42, 'Kitchen', 'Clean counters', false, null, ?)
		`, deadline)
		Expect(err).NotTo(HaveOccurred())

		first, err := persister.Complete(ctx, &choremodels.CompleteChoreParams{
			Id:          42,
			CompletedOn: originalDate,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(first.Chore.CompletedOn).To(Equal(originalDate))

		second, err := persister.Complete(ctx, &choremodels.CompleteChoreParams{
			Id:          42,
			CompletedOn: differentDate,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(second).To(Equal(first))

		persisted, err := persister.Get(ctx, 42)
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted).To(Equal(first.Chore))
	})

	It("returns not found when the chore does not exist", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)
		result, err := persister.Complete(ctx, &choremodels.CompleteChoreParams{
			Id: 42, CompletedOn: time.Date(2020, 3, 26, 0, 0, 0, 0, time.UTC),
		})
		Expect(result).To(BeNil())
		Expect(err).To(MatchError(choremodels.ErrNotFound))
	})

	When("completing a scheduled chore", func() {
		var (
			db        *sqlx.DB
			persister *chorespersistence.Persister
			params    *choremodels.CompleteChoreParams
		)
		BeforeEach(func(ctx SpecContext) {
			db = sqlitetest.NewMigratedDB()
			persister = chorespersistence.NewPersister(
				&chorespersistence.PersisterDeps{Db: db},
			)
			params = &choremodels.CompleteChoreParams{
				Id: 42, CompletedOn: time.Date(2020, 3, 26, 0, 0, 0, 0, time.UTC),
			}
			_, err := db.ExecContext(
				ctx,
				`insert into chore_templates (id, name, description)
				values (1, 'Current template', 'Current steps.')`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`insert into schedules (id, name, chore_template_id, interval_days)
				values (7, 'Weekly', 1, 7)`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`insert into chores (id, name, description, is_complete, deadline, schedule_id)
				values (42, 'Old snapshot', 'Old steps.', false, '2020-02-01', 7)`,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It(
			"atomically completes the original and persists a linked current-template snapshot",
			func(ctx SpecContext) {
				result, err := persister.Complete(ctx, params)
				Expect(err).NotTo(HaveOccurred())
				scheduleID, predecessorID := 7, 42
				Expect(result).To(Equal(&choremodels.CompleteChoreResult{
					Chore: &choremodels.Chore{
						Id:          42,
						ScheduleId:  &scheduleID,
						Name:        "Old snapshot",
						Description: "Old steps.",
						Status:      choremodels.ChoreStatusCompleted,
						Deadline: time.Date(
							2020,
							2,
							1,
							0,
							0,
							0,
							0,
							time.UTC,
						),
						CompletedOn: params.CompletedOn,
					},
					Successor: &choremodels.Chore{
						Id:            43,
						ScheduleId:    &scheduleID,
						PredecessorId: &predecessorID,
						Name:          "Current template",
						Description:   "Current steps.",
						Status:        choremodels.ChoreStatusActive,
						Deadline:      time.Date(2020, 4, 2, 0, 0, 0, 0, time.UTC),
					},
				}))
				original, err := persister.Get(ctx, params.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(original).To(Equal(result.Chore))
				successor, err := persister.Get(ctx, result.Successor.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(successor).To(Equal(result.Successor))
			},
		)

		DescribeTable(
			"adds calendar days across date boundaries",
			func(ctx SpecContext, completedOn, expectedDeadline string, intervalDays int) {
				date, err := time.Parse(time.DateOnly, completedOn)
				Expect(err).NotTo(HaveOccurred())
				params.CompletedOn = date
				_, err = db.ExecContext(
					ctx,
					`update schedules set interval_days = ? where id = 7`,
					intervalDays,
				)
				Expect(err).NotTo(HaveOccurred())
				result, err := persister.Complete(ctx, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(
					result.Successor.Deadline.Format(time.DateOnly),
				).To(Equal(expectedDeadline))
			},
			Entry("leap day", "2020-02-26", "2020-03-04", 7),
			Entry("year boundary", "2020-12-29", "2021-01-05", 7),
			Entry("one day", "2020-02-28", "2020-02-29", 1),
			Entry("fortnight", "2020-12-29", "2021-01-12", 14),
			Entry("maximum interval", "2010-01-01", "2019-12-30", 3650),
		)

		It(
			"returns the existing successor on repeated completion without changing either chore",
			func(ctx SpecContext) {
				first, err := persister.Complete(ctx, params)
				Expect(err).NotTo(HaveOccurred())
				params.CompletedOn = params.CompletedOn.AddDate(0, 0, 1)
				_, err = db.ExecContext(
					ctx,
					`update chore_templates set name = 'Changed again', description = null where id = 1`,
				)
				Expect(err).NotTo(HaveOccurred())

				repeated, err := persister.Complete(ctx, params)

				Expect(err).NotTo(HaveOccurred())
				Expect(repeated).To(Equal(first))
				var count int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from chores where schedule_id = 7`).
						Scan(&count),
				).To(Succeed())
				Expect(count).To(Equal(2))
			},
		)

		It(
			"keeps older completion retries idempotent after its successor is completed",
			func(ctx SpecContext) {
				first, err := persister.Complete(ctx, params)
				Expect(err).NotTo(HaveOccurred())
				second, err := persister.Complete(ctx, &choremodels.CompleteChoreParams{
					Id:          first.Successor.Id,
					CompletedOn: time.Date(2020, 4, 4, 0, 0, 0, 0, time.UTC),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(second.Successor.PredecessorId).To(Equal(&first.Successor.Id))
				Expect(
					second.Successor.Deadline,
				).To(Equal(time.Date(2020, 4, 11, 0, 0, 0, 0, time.UTC)))

				repeated, err := persister.Complete(ctx, params)

				Expect(err).NotTo(HaveOccurred())
				Expect(repeated.Chore).To(Equal(first.Chore))
				Expect(repeated.Successor).To(Equal(second.Chore))
				var total, active int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from chores where schedule_id = 7`).
						Scan(&total),
				).To(Succeed())
				Expect(
					db.QueryRowContext(ctx, `select count(*) from chores where schedule_id = 7 and is_complete = false`).
						Scan(&active),
				).To(Succeed())
				Expect(total).To(Equal(3))
				Expect(active).To(Equal(1))
			},
		)

		It(
			"rolls back completion if inserting the successor fails and permits a clean retry",
			func(ctx SpecContext) {
				original, err := persister.Get(ctx, params.Id)
				Expect(err).NotTo(HaveOccurred())
				_, err = db.ExecContext(
					ctx,
					`create trigger reject_successor before insert on chores
				when new.predecessor_id is not null begin select raise(abort, 'successor storage failed'); end`,
				)
				Expect(err).NotTo(HaveOccurred())

				result, err := persister.Complete(ctx, params)

				Expect(result).To(BeNil())
				Expect(err).To(MatchError(ContainSubstring("successor storage failed")))
				unchanged, err := persister.Get(ctx, params.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(unchanged).To(Equal(original))
				var count int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from chores where schedule_id = 7`).
						Scan(&count),
				).To(Succeed())
				Expect(count).To(Equal(1))
				_, err = db.ExecContext(ctx, `drop trigger reject_successor`)
				Expect(err).NotTo(HaveOccurred())
				retry, err := persister.Complete(ctx, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(retry.Chore.Status).To(Equal(choremodels.ChoreStatusCompleted))
				Expect(retry.Successor.PredecessorId).To(Equal(&original.Id))
			},
		)

		It(
			"does not create a successor for a deactivated schedule",
			func(ctx SpecContext) {
				_, err := db.ExecContext(
					ctx,
					`update schedules set deactivated_at = '2020-03-01' where id = 7`,
				)
				Expect(err).NotTo(HaveOccurred())
				result, err := persister.Complete(ctx, params)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError(schedulemodels.ErrInactive))
				original, err := persister.Get(ctx, params.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(original.Status).To(Equal(choremodels.ChoreStatusActive))
			},
		)

		It("enforces one successor per predecessor in SQLite", func(ctx SpecContext) {
			_, err := persister.Complete(ctx, params)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`insert into chores (name, is_complete, completed_on, deadline, schedule_id, predecessor_id)
				values ('Duplicate successor', true, '2020-04-03', '2020-04-02', 7, 42)`,
			)
			Expect(database.IsUniqueConstraintError(err)).To(BeTrue())
		})
	})
})
