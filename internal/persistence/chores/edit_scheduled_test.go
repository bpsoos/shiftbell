package chores_test

import (
	"database/sql"
	"errors"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	templatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edit scheduled chore", func() {
	var (
		db        *sqlx.DB
		persister *chores.Persister
		params    *models.EditScheduledChoreParams
		original  *models.Chore
	)

	BeforeEach(func(ctx SpecContext) {
		db = sqlitetest.NewMigratedDB()
		persister = chores.NewPersister(&chores.PersisterDeps{Db: db})
		_, err := db.ExecContext(
			ctx,
			`insert into chore_templates (id, name, description) values (1, 'Kitchen', 'Reusable steps')`,
		)
		Expect(err).NotTo(HaveOccurred())
		_, err = db.ExecContext(
			ctx,
			`insert into schedules (id, name, chore_template_id, interval_days)
			values (1, 'Weekly kitchen', 1, 7), (2, 'Monthly kitchen', 1, 30)`,
		)
		Expect(err).NotTo(HaveOccurred())
		deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
		_, err = db.ExecContext(
			ctx,
			`insert into chores
			(id, name, description, is_complete, completed_on, deadline, schedule_id, predecessor_id) values
			(41, 'Kitchen', 'Historical steps', true, ?, ?, 1, null),
			(42, 'Kitchen', 'Current steps', false, null, ?, 1, 41),
			(43, 'Kitchen', 'Other schedule steps', false, null, ?, 2, null),
			(44, 'One-off', 'Independent steps', false, null, ?, null, null)`,
			deadline.AddDate(
				0,
				0,
				-7,
			),
			deadline.AddDate(0, 0, -7),
			deadline,
			deadline,
			deadline,
		)
		Expect(err).NotTo(HaveOccurred())
		original, err = persister.Get(ctx, 42)
		Expect(err).NotTo(HaveOccurred())
		params = &models.EditScheduledChoreParams{
			Id:          42,
			Name:        "Kitchen edited",
			Description: "Edited steps",
		}
	})

	It(
		"updates only the active snapshot by default, preserving recurrence data",
		func(ctx SpecContext) {
			result, err := persister.EditScheduled(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			expected := *original
			expected.Name = params.Name
			expected.Description = params.Description
			Expect(result).To(Equal(&expected))
			stored, err := persister.Get(ctx, params.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(stored).To(Equal(result))
			var templateName, templateDescription, scheduleName string
			var templateId, interval int
			Expect(
				db.QueryRowContext(ctx, `select t.name, t.description, s.name, s.chore_template_id, s.interval_days
			from schedules s join chore_templates t on t.id = s.chore_template_id where s.id = 1`).
					Scan(&templateName, &templateDescription, &scheduleName, &templateId, &interval),
			).To(Succeed())
			Expect(templateName).To(Equal("Kitchen"))
			Expect(templateDescription).To(Equal("Reusable steps"))
			Expect(scheduleName).To(Equal("Weekly kitchen"))
			Expect(templateId).To(Equal(1))
			Expect(interval).To(Equal(7))
		},
	)

	It(
		"updates the linked template only when requested, without changing other snapshots",
		func(ctx SpecContext) {
			params.AlsoUpdateChoreTemplate = true
			historical, err := persister.Get(ctx, 41)
			Expect(err).NotTo(HaveOccurred())
			other, err := persister.Get(ctx, 43)
			Expect(err).NotTo(HaveOccurred())

			result, err := persister.EditScheduled(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			expected := *original
			expected.Name = params.Name
			expected.Description = params.Description
			Expect(result).To(Equal(&expected))
			var name, description string
			Expect(
				db.QueryRowContext(ctx, `select name, description from chore_templates where id = 1`).
					Scan(&name, &description),
			).To(Succeed())
			Expect(name).To(Equal(params.Name))
			Expect(description).To(Equal(params.Description))
			storedHistorical, err := persister.Get(ctx, 41)
			Expect(err).NotTo(HaveOccurred())
			Expect(storedHistorical).To(Equal(historical))
			storedOther, err := persister.Get(ctx, 43)
			Expect(err).NotTo(HaveOccurred())
			Expect(storedOther).To(Equal(other))
		},
	)

	DescribeTable(
		"stores an empty description as null in the selected resources",
		func(ctx SpecContext, updateTemplate bool, expectedTemplateDescription sql.NullString) {
			params.Description = ""
			params.AlsoUpdateChoreTemplate = updateTemplate

			result, err := persister.EditScheduled(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Description).To(BeEmpty())
			var choreDescription, templateDescription sql.NullString
			Expect(
				db.QueryRowContext(ctx, `select description from chores where id = 42`).
					Scan(&choreDescription),
			).To(Succeed())
			Expect(choreDescription.Valid).To(BeFalse())
			Expect(
				db.QueryRowContext(ctx, `select description from chore_templates where id = 1`).
					Scan(&templateDescription),
			).To(Succeed())
			Expect(templateDescription).To(Equal(expectedTemplateDescription))
		},
		Entry("chore only", false, sql.NullString{String: "Reusable steps", Valid: true}),
		Entry("chore and template", true, sql.NullString{}),
	)

	DescribeTable(
		"rejects non-active scheduled targets without changing their records",
		func(ctx SpecContext, id int) {
			params.Id = id
			params.AlsoUpdateChoreTemplate = true
			historical, err := persister.Get(ctx, 41)
			Expect(err).NotTo(HaveOccurred())
			oneOff, err := persister.Get(ctx, 44)
			Expect(err).NotTo(HaveOccurred())

			result, err := persister.EditScheduled(ctx, params)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(models.ErrNotFound))
			storedHistorical, err := persister.Get(ctx, 41)
			Expect(err).NotTo(HaveOccurred())
			Expect(storedHistorical).To(Equal(historical))
			storedOneOff, err := persister.Get(ctx, 44)
			Expect(err).NotTo(HaveOccurred())
			Expect(storedOneOff).To(Equal(oneOff))
			var name string
			Expect(
				db.QueryRowContext(ctx, `select name from chore_templates where id = 1`).
					Scan(&name),
			).To(Succeed())
			Expect(name).To(Equal("Kitchen"))
		},
		Entry("missing chore", 999),
		Entry("completed scheduled chore", 41),
		Entry("one-off chore", 44),
	)

	DescribeTable(
		"rolls back the chore update when the active template name conflicts",
		func(ctx SpecContext, name string) {
			_, err := db.ExecContext(
				ctx,
				`insert into chore_templates (name) values ('Laundry')`,
			)
			Expect(err).NotTo(HaveOccurred())
			params.Name = name
			params.AlsoUpdateChoreTemplate = true

			result, err := persister.EditScheduled(ctx, params)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(templatemodels.ErrNameConflict))
			stored, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(stored).To(Equal(original))
			var nameStored, description string
			Expect(
				db.QueryRowContext(ctx, `select name, description from chore_templates where id = 1`).
					Scan(&nameStored, &description),
			).To(Succeed())
			Expect(nameStored).To(Equal("Kitchen"))
			Expect(description).To(Equal("Reusable steps"))
		},
		Entry("identical name", "Laundry"),
		Entry("different case", "LAUNDRY"),
	)

	It(
		"allows a snapshot name that conflicts with a template when template updates are not requested",
		func(ctx SpecContext) {
			_, err := db.ExecContext(
				ctx,
				`insert into chore_templates (name) values ('Laundry')`,
			)
			Expect(err).NotTo(HaveOccurred())
			params.Name = "LAUNDRY"

			result, err := persister.EditScheduled(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Name).To(Equal(params.Name))
		},
	)

	DescribeTable(
		"allows the template to retain its own name",
		func(ctx SpecContext, name string) {
			params.Name = name
			params.AlsoUpdateChoreTemplate = true

			result, err := persister.EditScheduled(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Name).To(Equal(name))
			var templateName string
			Expect(
				db.QueryRowContext(ctx, `select name from chore_templates where id = 1`).
					Scan(&templateName),
			).To(Succeed())
			Expect(templateName).To(Equal(name))
		},
		Entry("same case", "Kitchen"),
		Entry("different case", "KITCHEN"),
	)

	It("allows a template name held by a deactivated template", func(ctx SpecContext) {
		_, err := db.ExecContext(
			ctx,
			`insert into chore_templates (name, deactivated_at) values ('Laundry', '2020-01-01')`,
		)
		Expect(err).NotTo(HaveOccurred())
		params.Name = "LAUNDRY"
		params.AlsoUpdateChoreTemplate = true

		result, err := persister.EditScheduled(ctx, params)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Name).To(Equal(params.Name))
	})

	It(
		"rolls back the chore update when the linked template is inactive",
		func(ctx SpecContext) {
			_, err := db.ExecContext(
				ctx,
				`update chore_templates set deactivated_at = '2020-01-01' where id = 1`,
			)
			Expect(err).NotTo(HaveOccurred())
			params.AlsoUpdateChoreTemplate = true

			result, err := persister.EditScheduled(ctx, params)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(templatemodels.ErrInactive))
			stored, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(stored).To(Equal(original))
		},
	)

	DescribeTable(
		"rolls back all changes when storage fails",
		func(ctx SpecContext, table string) {
			_, err := db.ExecContext(
				ctx,
				`create trigger reject_update before update on `+table+`
			begin select raise(abort, 'storage failed'); end`,
			)
			Expect(err).NotTo(HaveOccurred())
			params.AlsoUpdateChoreTemplate = true

			result, err := persister.EditScheduled(ctx, params)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("storage failed")))
			Expect(errors.Is(err, templatemodels.ErrNameConflict)).To(BeFalse())
			stored, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(stored).To(Equal(original))
			var name, description string
			Expect(
				db.QueryRowContext(ctx, `select name, description from chore_templates where id = 1`).
					Scan(&name, &description),
			).To(Succeed())
			Expect(name).To(Equal("Kitchen"))
			Expect(description).To(Equal("Reusable steps"))
		},
		Entry("chore update fails", "chores"),
		Entry("template update fails", "chore_templates"),
	)
})
