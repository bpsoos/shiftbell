package chores_test

import (
	"database/sql"
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

var _ = Describe("Create a manual scheduled chore", func() {
	var (
		db        *sqlx.DB
		persister *chores.Persister
		params    *models.CreateManualScheduledParams
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = chores.NewPersister(&chores.PersisterDeps{Db: db})
		params = &models.CreateManualScheduledParams{
			Name: "Kitchen", Description: "Wash and fold.",
			ScheduleName: "Weekly kitchen", IntervalDays: 7,
			Deadline: time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC),
		}
	})

	It(
		"persists a template, schedule, and first active snapshot with their associations",
		func(ctx SpecContext) {
			result, err := persister.CreateManualScheduled(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.ChoreTemplate).To(Equal(&templatemodels.ChoreTemplate{
				Id: 1, Name: params.Name, Description: params.Description,
			}))
			scheduleId := 1
			Expect(result.Chore).To(Equal(&models.Chore{
				Id:          1,
				ScheduleId:  &scheduleId,
				Name:        params.Name,
				Description: params.Description,
				Status:      models.ChoreStatusActive,
				Deadline:    params.Deadline,
			}))
			stored, err := persister.Get(ctx, result.Chore.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(stored).To(Equal(result.Chore))
			var templateName, templateDescription, scheduleName string
			var templateId, interval int
			var deactivated sql.NullTime
			Expect(db.QueryRowContext(ctx, `
			select t.id, t.name, t.description, s.name, s.interval_days, s.deactivated_at
			from schedules s join chore_templates t on t.id = s.chore_template_id
			where s.id = ?
		`, scheduleId).Scan(&templateId, &templateName, &templateDescription, &scheduleName, &interval, &deactivated)).To(Succeed())
			Expect(templateId).To(Equal(result.ChoreTemplate.Id))
			Expect(templateName).To(Equal(params.Name))
			Expect(templateDescription).To(Equal(params.Description))
			Expect(scheduleName).To(Equal(params.ScheduleName))
			Expect(interval).To(Equal(params.IntervalDays))
			Expect(deactivated.Valid).To(BeFalse())

			_, err = db.ExecContext(
				ctx,
				`update chore_templates set name = 'Changed', description = 'Changed' where id = ?`,
				templateId,
			)
			Expect(err).NotTo(HaveOccurred())
			stored, err = persister.Get(ctx, result.Chore.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(stored).To(Equal(result.Chore))
		},
	)

	It("stores an absent description as null in both snapshots", func(ctx SpecContext) {
		params.Description = ""

		result, err := persister.CreateManualScheduled(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Chore.Description).To(BeEmpty())
		Expect(result.ChoreTemplate.Description).To(BeEmpty())
		var choreDescription, templateDescription sql.NullString
		Expect(db.QueryRowContext(ctx, `
			select c.description, t.description from chores c
			join schedules s on s.id = c.schedule_id
			join chore_templates t on t.id = s.chore_template_id where c.id = ?
		`, result.Chore.Id).Scan(&choreDescription, &templateDescription)).To(Succeed())
		Expect(choreDescription.Valid).To(BeFalse())
		Expect(templateDescription.Valid).To(BeFalse())
	})

	DescribeTable(
		"rejects active template name conflicts without partial records",
		func(ctx SpecContext, name string) {
			_, err := db.ExecContext(
				ctx,
				`insert into chore_templates (name, description) values ('Kitchen', 'Original')`,
			)
			Expect(err).NotTo(HaveOccurred())
			params.Name = name

			result, err := persister.CreateManualScheduled(ctx, params)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(templatemodels.ErrNameConflict))
			var counts [3]int
			Expect(db.QueryRowContext(ctx, `select (select count(*) from chore_templates),
			(select count(*) from schedules), (select count(*) from chores)`).
				Scan(&counts[0], &counts[1], &counts[2])).To(Succeed())
			Expect(counts).To(Equal([3]int{1, 0, 0}))
			var description string
			Expect(
				db.QueryRowContext(ctx, `select description from chore_templates`).
					Scan(&description),
			).To(Succeed())
			Expect(description).To(Equal("Original"))
		},
		Entry("identical name", "Kitchen"),
		Entry("different case", "KITCHEN"),
	)

	DescribeTable(
		"rolls back the new template on active schedule name conflicts",
		func(ctx SpecContext, name string) {
			original, err := persister.CreateManualScheduled(ctx, params)
			Expect(err).NotTo(HaveOccurred())
			params.Name = "Laundry"
			params.ScheduleName = name

			result, err := persister.CreateManualScheduled(ctx, params)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(schedulemodels.ErrNameConflict))
			var counts [3]int
			Expect(db.QueryRowContext(ctx, `select (select count(*) from chore_templates),
			(select count(*) from schedules), (select count(*) from chores)`).
				Scan(&counts[0], &counts[1], &counts[2])).To(Succeed())
			Expect(counts).To(Equal([3]int{1, 1, 1}))
			stored, err := persister.Get(ctx, original.Chore.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(stored).To(Equal(original.Chore))
		},
		Entry("identical name", "Weekly kitchen"),
		Entry("different case", "WEEKLY KITCHEN"),
	)

	It("allows names held by deactivated templates and schedules", func(ctx SpecContext) {
		_, err := db.ExecContext(
			ctx,
			`insert into chore_templates (name, deactivated_at) values ('KITCHEN', '2020-01-01')`,
		)
		Expect(err).NotTo(HaveOccurred())
		_, err = db.ExecContext(
			ctx,
			`insert into schedules (name, chore_template_id, interval_days, deactivated_at)
			values ('WEEKLY KITCHEN', 1, 14, '2020-01-01')`,
		)
		Expect(err).NotTo(HaveOccurred())

		result, err := persister.CreateManualScheduled(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.ChoreTemplate.Id).To(Equal(2))
		Expect(*result.Chore.ScheduleId).To(Equal(2))
	})

	DescribeTable(
		"rolls back all new records when an insert fails",
		func(ctx SpecContext, table string) {
			_, err := db.ExecContext(
				ctx,
				`create trigger reject_insert before insert on `+table+`
			begin select raise(abort, 'storage failed'); end`,
			)
			Expect(err).NotTo(HaveOccurred())

			result, err := persister.CreateManualScheduled(ctx, params)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("storage failed")))
			Expect(errors.Is(err, templatemodels.ErrNameConflict)).To(BeFalse())
			Expect(errors.Is(err, schedulemodels.ErrNameConflict)).To(BeFalse())
			var counts [3]int
			Expect(db.QueryRowContext(ctx, `select (select count(*) from chore_templates),
			(select count(*) from schedules), (select count(*) from chores)`).
				Scan(&counts[0], &counts[1], &counts[2])).To(Succeed())
			Expect(counts).To(Equal([3]int{}))
		},
		Entry("template insert", "chore_templates"),
		Entry("schedule insert", "schedules"),
		Entry("chore insert", "chores"),
	)
})
