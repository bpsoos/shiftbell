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

var _ = Describe("Correct chore completion", func() {
	It("updates a completed one-off and returns it", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)
		deadline := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
		originalDate := time.Date(2020, time.February, 2, 0, 0, 0, 0, time.UTC)
		correctedDate := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
		_, err := db.ExecContext(ctx, `
			insert into chores (id, name, description, is_complete, completed_on, deadline)
			values (42, 'Kitchen', 'Clean counters', true, ?, ?)
		`, originalDate, deadline)
		Expect(err).NotTo(HaveOccurred())

		result, err := persister.CorrectCompletion(
			ctx,
			&choremodels.CorrectCompletionParams{
				Id:          42,
				CompletedOn: correctedDate,
			},
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(&choremodels.CorrectCompletionResult{
			Chore: &choremodels.Chore{
				Id:          42,
				Status:      choremodels.ChoreStatusCompleted,
				Name:        "Kitchen",
				Description: "Clean counters",
				Deadline:    deadline,
				CompletedOn: correctedDate,
			},
		}))
		persisted, err := persister.Get(ctx, 42)
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted).To(Equal(result.Chore))
	})

	It("returns not found for a missing chore", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)

		result, err := persister.CorrectCompletion(
			ctx,
			&choremodels.CorrectCompletionParams{
				Id:          42,
				CompletedOn: time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC),
			},
		)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(choremodels.ErrNotFound))
	})

	It("returns not found without changing an active chore", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)
		deadline := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
		_, err := db.ExecContext(ctx, `
			insert into chores (id, name, description, is_complete, completed_on, deadline)
			values (42, 'Kitchen', 'Clean counters', false, null, ?)
		`, deadline)
		Expect(err).NotTo(HaveOccurred())

		result, err := persister.CorrectCompletion(
			ctx,
			&choremodels.CorrectCompletionParams{
				Id:          42,
				CompletedOn: time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC),
			},
		)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(choremodels.ErrNotFound))
		persisted, err := persister.Get(ctx, 42)
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted).To(Equal(&choremodels.ChoreDetails{
			Id:          42,
			Status:      choremodels.ChoreStatusActive,
			Name:        "Kitchen",
			Description: "Clean counters",
			Deadline:    deadline,
		}))
	})

	When("correcting a scheduled chore", func() {
		var (
			db        *sqlx.DB
			persister *chorespersistence.Persister
			params    *choremodels.CorrectCompletionParams
		)

		BeforeEach(func(ctx SpecContext) {
			db = sqlitetest.NewMigratedDB()
			persister = chorespersistence.NewPersister(
				&chorespersistence.PersisterDeps{Db: db},
			)
			params = &choremodels.CorrectCompletionParams{
				Id: 42, CompletedOn: time.Date(2020, 3, 28, 0, 0, 0, 0, time.UTC),
			}
			_, err := db.ExecContext(
				ctx,
				`insert into chore_templates (id, name, description)
				values (1, 'Shared template', 'Current template steps.')`,
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
				`insert into chores (id, name, description, is_complete, completed_on, deadline, schedule_id, predecessor_id)
				values (42, 'Original snapshot', 'Original steps.', true, '2020-03-26', '2020-02-03', 7, null),
					(43, 'Successor snapshot', 'Locally edited steps.', false, null, '2020-04-02', 7, 42)`,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It(
			"corrects the predecessor and updates only the active successor's deadline",
			func(ctx SpecContext) {
				original, err := persister.Get(ctx, 42)
				Expect(err).NotTo(HaveOccurred())
				successor, err := persister.Get(ctx, 43)
				Expect(err).NotTo(HaveOccurred())

				result, err := persister.CorrectCompletion(ctx, params)

				Expect(err).NotTo(HaveOccurred())
				original.CompletedOn = params.CompletedOn
				successor.Deadline = time.Date(2020, 4, 4, 0, 0, 0, 0, time.UTC)
				Expect(
					result,
				).To(Equal(&choremodels.CorrectCompletionResult{Chore: original, Successor: successor}))
				persistedOriginal, err := persister.Get(ctx, 42)
				Expect(err).NotTo(HaveOccurred())
				Expect(persistedOriginal).To(Equal(original))
				persistedSuccessor, err := persister.Get(ctx, 43)
				Expect(err).NotTo(HaveOccurred())
				Expect(persistedSuccessor).To(Equal(successor))
				var count int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from chores where schedule_id = 7`).
						Scan(&count),
				).To(Succeed())
				Expect(count).To(Equal(2))
			},
		)

		DescribeTable(
			"uses calendar-day arithmetic and the schedule interval",
			func(ctx SpecContext, correctedDate, expectedDeadline string, interval int) {
				date, err := time.Parse(time.DateOnly, correctedDate)
				Expect(err).NotTo(HaveOccurred())
				params.CompletedOn = date
				_, err = db.ExecContext(
					ctx,
					`update schedules set interval_days = ? where id = 7`,
					interval,
				)
				Expect(err).NotTo(HaveOccurred())

				result, err := persister.CorrectCompletion(ctx, params)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Chore.CompletedOn).To(Equal(date))
				Expect(result.Successor.Id).To(Equal(43))
				Expect(
					result.Successor.Deadline.Format(time.DateOnly),
				).To(Equal(expectedDeadline))
			},
			Entry("one day across leap day", "2020-02-28", "2020-02-29", 1),
			Entry("seven days across a month", "2020-02-26", "2020-03-04", 7),
			Entry("fourteen days across a year", "2020-12-29", "2021-01-12", 14),
			Entry("maximum interval", "2010-01-01", "2019-12-30", 3650),
		)

		It(
			"keeps repeated corrections stable without replacing either chore",
			func(ctx SpecContext) {
				first, err := persister.CorrectCompletion(ctx, params)
				Expect(err).NotTo(HaveOccurred())

				repeated, err := persister.CorrectCompletion(ctx, params)

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
			"rolls back the corrected completion date when updating the successor fails",
			func(ctx SpecContext) {
				original, err := persister.Get(ctx, 42)
				Expect(err).NotTo(HaveOccurred())
				successor, err := persister.Get(ctx, 43)
				Expect(err).NotTo(HaveOccurred())
				_, err = db.ExecContext(
					ctx,
					`create trigger reject_deadline_correction before update of deadline on chores
				when old.id = 43 begin select raise(abort, 'deadline correction failed'); end`,
				)
				Expect(err).NotTo(HaveOccurred())

				result, err := persister.CorrectCompletion(ctx, params)

				Expect(result).To(BeNil())
				Expect(err).To(MatchError(ContainSubstring("deadline correction failed")))
				unchangedOriginal, err := persister.Get(ctx, 42)
				Expect(err).NotTo(HaveOccurred())
				Expect(unchangedOriginal).To(Equal(original))
				unchangedSuccessor, err := persister.Get(ctx, 43)
				Expect(err).NotTo(HaveOccurred())
				Expect(unchangedSuccessor).To(Equal(successor))
				_, err = db.ExecContext(ctx, `drop trigger reject_deadline_correction`)
				Expect(err).NotTo(HaveOccurred())
				retry, err := persister.CorrectCompletion(ctx, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(retry.Chore.CompletedOn).To(Equal(params.CompletedOn))
				Expect(
					retry.Successor.Deadline,
				).To(Equal(time.Date(2020, 4, 4, 0, 0, 0, 0, time.UTC)))
			},
		)

		When("the immediate successor has completed", func() {
			BeforeEach(func(ctx SpecContext) {
				_, err := db.ExecContext(
					ctx,
					`update chores set is_complete = true, completed_on = '2020-04-03' where id = 43`,
				)
				Expect(err).NotTo(HaveOccurred())
				_, err = db.ExecContext(
					ctx,
					`insert into chores (id, name, is_complete, deadline, schedule_id, predecessor_id)
					values (44, 'Later snapshot', false, '2020-04-10', 7, 43)`,
				)
				Expect(err).NotTo(HaveOccurred())
			})

			It(
				"does not change the completed successor or cascade to the later active chore",
				func(ctx SpecContext) {
					successor, err := persister.Get(ctx, 43)
					Expect(err).NotTo(HaveOccurred())
					later, err := persister.Get(ctx, 44)
					Expect(err).NotTo(HaveOccurred())

					result, err := persister.CorrectCompletion(ctx, params)

					Expect(err).NotTo(HaveOccurred())
					Expect(result.Chore.CompletedOn).To(Equal(params.CompletedOn))
					Expect(result.Successor).To(Equal(successor))
					unchangedSuccessor, err := persister.Get(ctx, 43)
					Expect(err).NotTo(HaveOccurred())
					Expect(unchangedSuccessor).To(Equal(successor))
					unchangedLater, err := persister.Get(ctx, 44)
					Expect(err).NotTo(HaveOccurred())
					Expect(unchangedLater).To(Equal(later))
				},
			)

			It(
				"does not affect a later active chore when the immediate successor has been deleted",
				func(ctx SpecContext) {
					_, err := db.ExecContext(ctx, `delete from chores where id = 43`)
					Expect(err).NotTo(HaveOccurred())
					later, err := persister.Get(ctx, 44)
					Expect(err).NotTo(HaveOccurred())

					result, err := persister.CorrectCompletion(ctx, params)

					Expect(err).NotTo(HaveOccurred())
					Expect(result.Chore.CompletedOn).To(Equal(params.CompletedOn))
					Expect(result.Successor).To(BeNil())
					unchangedLater, err := persister.Get(ctx, 44)
					Expect(err).NotTo(HaveOccurred())
					Expect(unchangedLater).To(Equal(later))
					var count int
					Expect(
						db.QueryRowContext(ctx, `select count(*) from chores where schedule_id = 7`).
							Scan(&count),
					).To(Succeed())
					Expect(count).To(Equal(2))
				},
			)
		})

		It(
			"corrects history for a deactivated schedule without recreating its removed active chore",
			func(ctx SpecContext) {
				_, err := db.ExecContext(ctx, `delete from chores where id = 43`)
				Expect(err).NotTo(HaveOccurred())
				_, err = db.ExecContext(
					ctx,
					`update schedules set deactivated_at = '2020-03-27' where id = 7`,
				)
				Expect(err).NotTo(HaveOccurred())

				result, err := persister.CorrectCompletion(ctx, params)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Chore.CompletedOn).To(Equal(params.CompletedOn))
				Expect(result.Successor).To(BeNil())
				var count int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from chores where schedule_id = 7`).
						Scan(&count),
				).To(Succeed())
				Expect(count).To(Equal(1))
			},
		)
	})
})
