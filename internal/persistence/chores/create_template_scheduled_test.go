package chores_test

import (
	"errors"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	templatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create a template-based scheduled chore", func() {
	var (
		db        *sqlx.DB
		persister *chores.Persister
		params    *models.CreateTemplateScheduledParams
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = chores.NewPersister(&chores.PersisterDeps{Db: db})
		params = &models.CreateTemplateScheduledParams{
			ChoreTemplateId: 42,
			ScheduleName:    "Weekly kitchen",
			IntervalDays:    7,
			Deadline:        time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC),
		}
	})

	It(
		"rejects a missing template without creating either resource",
		func(ctx SpecContext) {
			result, err := persister.CreateTemplateScheduled(ctx, params)
			Expect(result).To(BeNil())
			Expect(err).To(MatchError(templatemodels.ErrNotFound))
			var scheduleCount, choreCount int
			Expect(
				db.QueryRowContext(ctx, `select count(*) from schedules`).
					Scan(&scheduleCount),
			).To(Succeed())
			Expect(
				db.QueryRowContext(ctx, `select count(*) from chores`).Scan(&choreCount),
			).To(Succeed())
			Expect(scheduleCount).To(BeZero())
			Expect(choreCount).To(BeZero())
		},
	)

	When("the template exists", func() {
		BeforeEach(func(ctx SpecContext) {
			_, err := db.ExecContext(
				ctx,
				`insert into chore_templates (id, name, description) values (42, 'Kitchen', 'Reusable steps.')`,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It(
			"persists the schedule and a linked snapshot of the template",
			func(ctx SpecContext) {
				result, err := persister.CreateTemplateScheduled(ctx, params)
				Expect(err).NotTo(HaveOccurred())
				scheduleID := 1
				Expect(result).To(Equal(&models.CreateChoreResult{Chore: &models.Chore{
					Id:          1,
					ScheduleId:  &scheduleID,
					Name:        "Kitchen",
					Description: "Reusable steps.",
					Status:      models.ChoreStatusActive,
					Deadline:    params.Deadline,
				}}))
				var name string
				var templateID, interval int
				Expect(
					db.QueryRowContext(ctx, `select name, chore_template_id, interval_days from schedules where id = ?`, scheduleID).
						Scan(&name, &templateID, &interval),
				).To(Succeed())
				Expect(name).To(Equal(params.ScheduleName))
				Expect(templateID).To(Equal(params.ChoreTemplateId))
				Expect(interval).To(Equal(params.IntervalDays))
				var choreScheduleID int
				var choreName, description string
				Expect(
					db.QueryRowContext(ctx, `select schedule_id, name, description from chores where id = ?`, result.Chore.Id).
						Scan(&choreScheduleID, &choreName, &description),
				).To(Succeed())
				Expect(choreScheduleID).To(Equal(scheduleID))
				Expect(choreName).To(Equal("Kitchen"))
				Expect(description).To(Equal("Reusable steps."))
				_, err = db.ExecContext(
					ctx,
					`update chore_templates set name = 'Changed', description = 'Changed steps.' where id = 42`,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(
					db.QueryRowContext(ctx, `select name, description from chores where id = ?`, result.Chore.Id).
						Scan(&choreName, &description),
				).To(Succeed())
				Expect(choreName).To(Equal("Kitchen"))
				Expect(description).To(Equal("Reusable steps."))
			},
		)

		It(
			"rejects an inactive template without creating either resource",
			func(ctx SpecContext) {
				_, err := db.ExecContext(
					ctx,
					`update chore_templates set deactivated_at = ? where id = 42`,
					time.Now(),
				)
				Expect(err).NotTo(HaveOccurred())
				result, err := persister.CreateTemplateScheduled(ctx, params)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError(templatemodels.ErrInactive))
				var scheduleCount, choreCount int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from schedules`).
						Scan(&scheduleCount),
				).To(Succeed())
				Expect(
					db.QueryRowContext(ctx, `select count(*) from chores`).
						Scan(&choreCount),
				).To(Succeed())
				Expect(scheduleCount).To(BeZero())
				Expect(choreCount).To(BeZero())
			},
		)

		It("rolls back the schedule if inserting the chore fails", func(ctx SpecContext) {
			_, err := db.ExecContext(
				ctx,
				`create trigger reject_scheduled_chore before insert on chores
				when new.schedule_id is not null begin select raise(abort, 'scheduled chore insert failed'); end`,
			)
			Expect(err).NotTo(HaveOccurred())
			result, err := persister.CreateTemplateScheduled(ctx, params)
			Expect(result).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("scheduled chore insert failed")))
			var scheduleCount, choreCount, templateCount int
			Expect(
				db.QueryRowContext(ctx, `select count(*) from schedules`).
					Scan(&scheduleCount),
			).To(Succeed())
			Expect(
				db.QueryRowContext(ctx, `select count(*) from chores`).Scan(&choreCount),
			).To(Succeed())
			Expect(
				db.QueryRowContext(ctx, `select count(*) from chore_templates`).
					Scan(&templateCount),
			).To(Succeed())
			Expect(scheduleCount).To(BeZero())
			Expect(choreCount).To(BeZero())
			Expect(templateCount).To(Equal(1))
		})

		DescribeTable(
			"rejects an existing active schedule name without persisting partial records",
			func(ctx SpecContext, name string) {
				original, err := persister.CreateTemplateScheduled(ctx, params)
				Expect(err).NotTo(HaveOccurred())
				_, err = db.ExecContext(
					ctx,
					`insert into chore_templates (id, name) values (43, 'Laundry')`,
				)
				Expect(err).NotTo(HaveOccurred())
				duplicate := *params
				duplicate.ChoreTemplateId = 43
				duplicate.ScheduleName = name
				duplicate.IntervalDays = 14
				duplicate.Deadline = params.Deadline.AddDate(0, 0, 7)

				result, err := persister.CreateTemplateScheduled(ctx, &duplicate)

				Expect(result).To(BeNil())
				Expect(err).To(MatchError(schedulemodels.ErrNameConflict))
				chore, err := persister.Get(ctx, original.Chore.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(chore).To(Equal(original.Chore))
				var scheduleCount, choreCount int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from schedules`).
						Scan(&scheduleCount),
				).To(Succeed())
				Expect(
					db.QueryRowContext(ctx, `select count(*) from chores`).
						Scan(&choreCount),
				).To(Succeed())
				Expect(scheduleCount).To(Equal(1))
				Expect(choreCount).To(Equal(1))
				var storedName string
				var templateID, interval int
				Expect(
					db.QueryRowContext(ctx, `select name, chore_template_id, interval_days from schedules where id = ?`, *original.Chore.ScheduleId).
						Scan(&storedName, &templateID, &interval),
				).To(Succeed())
				Expect(storedName).To(Equal(params.ScheduleName))
				Expect(templateID).To(Equal(params.ChoreTemplateId))
				Expect(interval).To(Equal(params.IntervalDays))
			},
			Entry("identical name", "Weekly kitchen"),
			Entry("different case", "WEEKLY KITCHEN"),
		)

		It(
			"allows multiple active schedules with different names for the same template",
			func(ctx SpecContext) {
				first, err := persister.CreateTemplateScheduled(ctx, params)
				Expect(err).NotTo(HaveOccurred())
				params.ScheduleName = "Fortnightly kitchen"
				params.IntervalDays = 14

				second, err := persister.CreateTemplateScheduled(ctx, params)

				Expect(err).NotTo(HaveOccurred())
				Expect(second.Chore.ScheduleId).NotTo(Equal(first.Chore.ScheduleId))
				var count int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from schedules where chore_template_id = 42 and deactivated_at is null`).
						Scan(&count),
				).To(Succeed())
				Expect(count).To(Equal(2))
			},
		)

		It("reuses a name held by deactivated schedules", func(ctx SpecContext) {
			deactivatedAt := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
			_, err := db.ExecContext(
				ctx,
				`insert into schedules (id, name, chore_template_id, interval_days, deactivated_at)
				values (1, 'WEEKLY KITCHEN', 42, 14, ?), (2, 'Weekly kitchen', 42, 30, ?)`,
				deactivatedAt,
				deactivatedAt,
			)
			Expect(err).NotTo(HaveOccurred())

			result, err := persister.CreateTemplateScheduled(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			Expect(*result.Chore.ScheduleId).To(Equal(3))
			var archivedCount, activeCount int
			Expect(
				db.QueryRowContext(ctx, `select count(*) from schedules where deactivated_at is not null`).
					Scan(&archivedCount),
			).To(Succeed())
			Expect(
				db.QueryRowContext(ctx, `select count(*) from schedules where deactivated_at is null`).
					Scan(&activeCount),
			).To(Succeed())
			Expect(archivedCount).To(Equal(2))
			Expect(activeCount).To(Equal(1))
		})

		It(
			"preserves non-unique-constraint schedule insertion errors",
			func(ctx SpecContext) {
				_, err := db.ExecContext(
					ctx,
					`create trigger reject_schedule before insert on schedules
				begin select raise(abort, 'schedule storage failed'); end`,
				)
				Expect(err).NotTo(HaveOccurred())

				result, err := persister.CreateTemplateScheduled(ctx, params)

				Expect(result).To(BeNil())
				Expect(err).To(MatchError(ContainSubstring("schedule storage failed")))
				Expect(errors.Is(err, schedulemodels.ErrNameConflict)).To(BeFalse())
				var scheduleCount, choreCount int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from schedules`).
						Scan(&scheduleCount),
				).To(Succeed())
				Expect(
					db.QueryRowContext(ctx, `select count(*) from chores`).
						Scan(&choreCount),
				).To(Succeed())
				Expect(scheduleCount).To(BeZero())
				Expect(choreCount).To(BeZero())
			},
		)
	})
})
