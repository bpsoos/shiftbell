package schedules_test

import (
	"database/sql"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/bpsoos/shiftbell/internal/persistence/schedules"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edit", func() {
	var (
		db        *sqlx.DB
		persister *schedules.Persister
		params    *models.EditScheduleParams
	)
	BeforeEach(func(ctx SpecContext) {
		db = sqlitetest.NewMigratedDB()
		persister = schedules.NewPersister(&schedules.PersisterDeps{Db: db})
		params = &models.EditScheduleParams{
			Id:           42,
			Name:         "Biweekly laundry",
			IntervalDays: 14,
		}
		_, err := db.ExecContext(ctx, `insert into chore_templates (id, name, description)
			values (7, 'Wash clothes', 'Shared steps')`)
		Expect(err).NotTo(HaveOccurred())
		_, err = db.ExecContext(
			ctx,
			`insert into schedules (id, name, chore_template_id, interval_days)
			values (42, 'Weekly laundry', 7, 7), (43, 'Monthly laundry', 7, 30)`,
		)
		Expect(err).NotTo(HaveOccurred())
		_, err = db.ExecContext(
			ctx,
			`insert into chores (id, name, description, is_complete, completed_on, deadline, schedule_id, predecessor_id)
			values (10, 'Old snapshot', 'Historical steps', true, '2020-03-26', '2020-02-03', 42, null),
				(11, 'Current snapshot', 'Locally edited steps', false, null, '2020-04-02', 42, 10)`,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("updates only the schedule name and interval", func(ctx SpecContext) {
		type choreSnapshot struct {
			Id              int
			ChoreTemplateID sql.NullInt64 `db:"chore_template_id"`
			Name            string
			Description     string
			IsComplete      bool         `db:"is_complete"`
			CompletedOn     sql.NullTime `db:"completed_on"`
			Deadline        time.Time
			ScheduleID      int           `db:"schedule_id"`
			PredecessorID   sql.NullInt64 `db:"predecessor_id"`
		}
		var originalChores []choreSnapshot
		Expect(
			db.SelectContext(ctx, &originalChores, `select * from chores order by id`),
		).To(Succeed())
		other, err := persister.Get(ctx, 43)
		Expect(err).NotTo(HaveOccurred())

		result, err := persister.Edit(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		choreID := 11
		deadline := time.Date(2020, 4, 2, 0, 0, 0, 0, time.UTC)
		Expect(result).To(Equal(&models.Schedule{
			Id:                  42,
			Name:                "Biweekly laundry",
			ChoreTemplateId:     7,
			ChoreTemplateName:   "Wash clothes",
			IntervalDays:        14,
			ActiveChoreId:       &choreID,
			ActiveChoreDeadline: &deadline,
		}))
		persisted, err := persister.Get(ctx, 42)
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted).To(Equal(result))
		unchangedOther, err := persister.Get(ctx, 43)
		Expect(err).NotTo(HaveOccurred())
		Expect(unchangedOther).To(Equal(other))
		var unchangedChores []choreSnapshot
		Expect(
			db.SelectContext(ctx, &unchangedChores, `select * from chores order by id`),
		).To(Succeed())
		Expect(unchangedChores).To(Equal(originalChores))
		var templateName, templateDescription string
		Expect(
			db.QueryRowContext(ctx, `select name, description from chore_templates where id = 7`).
				Scan(&templateName, &templateDescription),
		).To(Succeed())
		Expect(templateName).To(Equal("Wash clothes"))
		Expect(templateDescription).To(Equal("Shared steps"))
	})

	It("allows retaining the current name with different case", func(ctx SpecContext) {
		params.Name = "WEEKLY LAUNDRY"

		result, err := persister.Edit(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Name).To(Equal(params.Name))
		Expect(result.IntervalDays).To(Equal(14))
	})

	It(
		"rejects a conflicting active name without changing either schedule",
		func(ctx SpecContext) {
			original, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			other, err := persister.Get(ctx, 43)
			Expect(err).NotTo(HaveOccurred())
			params.Name = "MONTHLY LAUNDRY"

			result, err := persister.Edit(ctx, params)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(models.ErrNameConflict))
			unchanged, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(unchanged).To(Equal(original))
			unchangedOther, err := persister.Get(ctx, 43)
			Expect(err).NotTo(HaveOccurred())
			Expect(unchangedOther).To(Equal(other))
		},
	)

	It("allows reusing a deactivated schedule name", func(ctx SpecContext) {
		_, err := db.ExecContext(
			ctx,
			`update schedules set deactivated_at = '2020-03-27' where id = 43`,
		)
		Expect(err).NotTo(HaveOccurred())
		params.Name = "MONTHLY LAUNDRY"

		result, err := persister.Edit(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Name).To(Equal(params.Name))
	})

	It(
		"rejects editing a deactivated schedule without changing it",
		func(ctx SpecContext) {
			_, err := db.ExecContext(ctx, `delete from chores where id = 11`)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`update schedules set deactivated_at = '2020-03-27' where id = 42`,
			)
			Expect(err).NotTo(HaveOccurred())
			original, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())

			result, err := persister.Edit(ctx, params)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(models.ErrInactive))
			unchanged, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(unchanged).To(Equal(original))
		},
	)

	It("returns not found for an unknown schedule", func(ctx SpecContext) {
		params.Id = 99

		result, err := persister.Edit(ctx, params)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(models.ErrNotFound))
	})

	It(
		"does not classify an update failure as a name conflict and allows retrying",
		func(ctx SpecContext) {
			original, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`create trigger reject_schedule_edit before update on schedules
			begin select raise(abort, 'schedule edit failed'); end`,
			)
			Expect(err).NotTo(HaveOccurred())

			result, err := persister.Edit(ctx, params)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("schedule edit failed")))
			Expect(err).NotTo(MatchError(models.ErrNameConflict))
			unchanged, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(unchanged).To(Equal(original))
			_, err = db.ExecContext(ctx, `drop trigger reject_schedule_edit`)
			Expect(err).NotTo(HaveOccurred())
			retry, err := persister.Edit(ctx, params)
			Expect(err).NotTo(HaveOccurred())
			Expect(retry.Name).To(Equal(params.Name))
			Expect(retry.IntervalDays).To(Equal(params.IntervalDays))
		},
	)
})
