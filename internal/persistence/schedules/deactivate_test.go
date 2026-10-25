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

var _ = Describe("Deactivate", func() {
	var (
		db        *sqlx.DB
		persister *schedules.Persister
		params    *models.DeactivateScheduleParams
	)
	BeforeEach(func(ctx SpecContext) {
		db = sqlitetest.NewMigratedDB()
		persister = schedules.NewPersister(&schedules.PersisterDeps{Db: db})
		params = &models.DeactivateScheduleParams{
			Id: 42, DeactivatedAt: time.Date(2020, 4, 3, 10, 0, 0, 0, time.UTC),
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
			values (9, 'Older snapshot', 'Older steps', true, '2020-03-19', '2020-02-01', 42, null),
				(10, 'Old snapshot', 'Historical steps', true, '2020-03-26', '2020-03-26', 42, 9),
				(11, 'Current snapshot', 'Local steps', false, null, '2020-04-02', 42, 10),
				(12, 'Other schedule', 'Other steps', false, null, '2020-04-25', 43, null),
				(13, 'One-off chore', 'One-off steps', false, null, '2020-04-01', null, null)`,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It(
		"deactivates the schedule and removes only its active chore",
		func(ctx SpecContext) {
			type choreSnapshot struct {
				Id            int
				Name          string
				Description   string
				IsComplete    bool         `db:"is_complete"`
				CompletedOn   sql.NullTime `db:"completed_on"`
				Deadline      time.Time
				ScheduleID    sql.NullInt64 `db:"schedule_id"`
				PredecessorID sql.NullInt64 `db:"predecessor_id"`
			}
			var originalChores []choreSnapshot
			query := `select id, name, description, is_complete, completed_on, deadline, schedule_id, predecessor_id
			from chores where id != 11 order by id`
			Expect(db.SelectContext(ctx, &originalChores, query)).To(Succeed())
			other, err := persister.Get(ctx, 43)
			Expect(err).NotTo(HaveOccurred())

			result, err := persister.Deactivate(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(&models.Schedule{
				Id:                42,
				Name:              "Weekly laundry",
				ChoreTemplateId:   7,
				ChoreTemplateName: "Wash clothes",
				IntervalDays:      7,
				DeactivatedAt:     &params.DeactivatedAt,
			}))
			persisted, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(persisted).To(Equal(result))
			var activeCount int
			Expect(
				db.QueryRowContext(ctx, `select count(*) from chores where id = 11`).
					Scan(&activeCount),
			).To(Succeed())
			Expect(activeCount).To(BeZero())
			var remainingChores []choreSnapshot
			Expect(db.SelectContext(ctx, &remainingChores, query)).To(Succeed())
			Expect(remainingChores).To(Equal(originalChores))
			unchangedOther, err := persister.Get(ctx, 43)
			Expect(err).NotTo(HaveOccurred())
			Expect(unchangedOther).To(Equal(other))
			var templateName, templateDescription string
			var templateDeactivatedAt sql.NullTime
			Expect(
				db.QueryRowContext(ctx, `select name, description, deactivated_at from chore_templates where id = 7`).
					Scan(&templateName, &templateDescription, &templateDeactivatedAt),
			).To(Succeed())
			Expect(templateName).To(Equal("Wash clothes"))
			Expect(templateDescription).To(Equal("Shared steps"))
			Expect(templateDeactivatedAt.Valid).To(BeFalse())
		},
	)

	It(
		"keeps repeated requests stable and preserves the original deactivation time",
		func(ctx SpecContext) {
			first, err := persister.Deactivate(ctx, params)
			Expect(err).NotTo(HaveOccurred())
			params = &models.DeactivateScheduleParams{
				Id:            42,
				DeactivatedAt: params.DeactivatedAt.Add(24 * time.Hour),
			}

			repeated, err := persister.Deactivate(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			Expect(repeated).To(Equal(first))
			persisted, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(persisted).To(Equal(first))
			var count int
			Expect(
				db.QueryRowContext(ctx, `select count(*) from chores where schedule_id = 42`).
					Scan(&count),
			).To(Succeed())
			Expect(count).To(Equal(2))
		},
	)

	It("deactivates a schedule without an active chore", func(ctx SpecContext) {
		_, err := db.ExecContext(ctx, `delete from chores where id = 11`)
		Expect(err).NotTo(HaveOccurred())

		result, err := persister.Deactivate(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.DeactivatedAt).To(Equal(&params.DeactivatedAt))
		Expect(result.ActiveChoreId).To(BeNil())
		Expect(result.ActiveChoreDeadline).To(BeNil())
	})

	It("returns not found for an unknown schedule", func(ctx SpecContext) {
		params.Id = 99

		result, err := persister.Deactivate(ctx, params)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(models.ErrNotFound))
	})

	It(
		"rolls back deactivation when removing the active chore fails and allows retrying",
		func(ctx SpecContext) {
			original, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`create trigger reject_active_chore_removal before delete on chores
			when old.id = 11 begin select raise(abort, 'active chore removal failed'); end`,
			)
			Expect(err).NotTo(HaveOccurred())

			result, err := persister.Deactivate(ctx, params)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("active chore removal failed")))
			unchanged, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(unchanged).To(Equal(original))
			var count int
			Expect(
				db.QueryRowContext(ctx, `select count(*) from chores where schedule_id = 42`).
					Scan(&count),
			).To(Succeed())
			Expect(count).To(Equal(3))
			_, err = db.ExecContext(ctx, `drop trigger reject_active_chore_removal`)
			Expect(err).NotTo(HaveOccurred())
			retry, err := persister.Deactivate(ctx, params)
			Expect(err).NotTo(HaveOccurred())
			Expect(retry.DeactivatedAt).To(Equal(&params.DeactivatedAt))
			Expect(retry.ActiveChoreId).To(BeNil())
			Expect(
				db.QueryRowContext(ctx, `select count(*) from chores where schedule_id = 42`).
					Scan(&count),
			).To(Succeed())
			Expect(count).To(Equal(2))
		},
	)

	It(
		"leaves the active chore untouched when recording deactivation fails",
		func(ctx SpecContext) {
			original, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`create trigger reject_schedule_deactivation before update of deactivated_at on schedules
			when old.id = 42 begin select raise(abort, 'schedule deactivation failed'); end`,
			)
			Expect(err).NotTo(HaveOccurred())

			result, err := persister.Deactivate(ctx, params)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("schedule deactivation failed")))
			unchanged, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(unchanged).To(Equal(original))
		},
	)
})
