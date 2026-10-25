package schedules_test

import (
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/bpsoos/shiftbell/internal/persistence/schedules"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get", func() {
	var (
		db        *sqlx.DB
		persister *schedules.Persister
	)
	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = schedules.NewPersister(&schedules.PersisterDeps{Db: db})
	})

	It("returns not found for an unknown schedule", func(ctx SpecContext) {
		result, err := persister.Get(ctx, 42)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(models.ErrNotFound))
	})

	It("returns database failures", func(ctx SpecContext) {
		Expect(db.Close()).To(Succeed())

		result, err := persister.Get(ctx, 42)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("select schedule")))
		Expect(err).NotTo(MatchError(models.ErrNotFound))
	})

	When("a schedule exists", func() {
		BeforeEach(func(ctx SpecContext) {
			_, err := db.ExecContext(
				ctx,
				`insert into chore_templates (id, name) values (7, 'Wash clothes')`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`insert into schedules (id, name, chore_template_id, interval_days)
				values (42, 'Weekly laundry', 7, 7)`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`insert into chores (id, name, is_complete, completed_on, deadline, schedule_id)
				values (10, 'Old snapshot', true, '2020-03-26', '2020-02-03', 42),
					(11, 'Current snapshot', false, null, '2020-04-02', 42)`,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It(
			"returns schedule details with its template and only its active chore",
			func(ctx SpecContext) {
				result, err := persister.Get(ctx, 42)

				Expect(err).NotTo(HaveOccurred())
				choreID := 11
				deadline := time.Date(2020, 4, 2, 0, 0, 0, 0, time.UTC)
				Expect(result).To(Equal(&models.ScheduleDetails{
					Id:                  42,
					Name:                "Weekly laundry",
					ChoreTemplateId:     7,
					ChoreTemplateName:   "Wash clothes",
					IntervalDays:        7,
					ActiveChoreId:       &choreID,
					ActiveChoreDeadline: &deadline,
				}))
			},
		)

		It(
			"returns deactivated schedules without treating completed history as active",
			func(ctx SpecContext) {
				_, err := db.ExecContext(ctx, `delete from chores where id = 11`)
				Expect(err).NotTo(HaveOccurred())
				deactivatedAt := time.Date(2020, 4, 3, 10, 0, 0, 0, time.UTC)
				_, err = db.ExecContext(
					ctx,
					`update schedules set deactivated_at = ? where id = 42`,
					deactivatedAt,
				)
				Expect(err).NotTo(HaveOccurred())
				_, err = db.ExecContext(
					ctx,
					`update chore_templates set deactivated_at = ? where id = 7`,
					deactivatedAt,
				)
				Expect(err).NotTo(HaveOccurred())

				result, err := persister.Get(ctx, 42)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(&models.ScheduleDetails{
					Id:                42,
					Name:              "Weekly laundry",
					ChoreTemplateId:   7,
					ChoreTemplateName: "Wash clothes",
					IntervalDays:      7,
					DeactivatedAt:     &deactivatedAt,
				}))
			},
		)

		It("returns schedules without an active chore", func(ctx SpecContext) {
			_, err := db.ExecContext(ctx, `delete from chores where id = 11`)
			Expect(err).NotTo(HaveOccurred())

			result, err := persister.Get(ctx, 42)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.ActiveChoreId).To(BeNil())
			Expect(result.ActiveChoreDeadline).To(BeNil())
		})
	})
})
