package chores_test

import (
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	templatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
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
	})
})
