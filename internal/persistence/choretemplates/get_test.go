package choretemplates_test

import (
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	choretemplatespersistence "github.com/bpsoos/shiftbell/internal/persistence/choretemplates"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get", func() {
	var (
		db        *sqlx.DB
		persister *choretemplatespersistence.Persister
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = choretemplatespersistence.NewChoreTemplatePersister(
			&choretemplatespersistence.PersisterDeps{Db: db},
		)
	})

	Context("with no chore templates", func() {
		It("returns an error", func(ctx SpecContext) {
			result, err := persister.Get(ctx, 1)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(models.ErrNotFound))
		})
	})

	Context("with one chore template", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chore_templates (id, name, description) values (?, ?, ?)`,
				1,
				"Laundry",
				"Wash and fold clothes",
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the chore template", func(ctx SpecContext) {
			result, err := persister.Get(ctx, 1)

			Expect(err).NotTo(HaveOccurred())
			Expect(
				result,
			).To(Equal(&models.ChoreTemplateDetails{Id: 1, Name: "Laundry", Description: "Wash and fold clothes"}))
		})
	})

	Context("with a chore template without a description", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chore_templates (id, name, description) values (?, ?, ?)`,
				1,
				"Laundry",
				nil,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an empty description", func(ctx SpecContext) {
			result, err := persister.Get(ctx, 1)

			Expect(err).NotTo(HaveOccurred())
			Expect(
				result,
			).To(Equal(&models.ChoreTemplateDetails{Id: 1, Name: "Laundry"}))
		})
	})

	Context("with many chore templates", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chore_templates (id, name, description) values
					(?, ?, ?),
					(?, ?, ?)`,
				1, "Laundry", "Wash and fold clothes",
				2, "Dishes", "Load the dishwasher",
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the requested chore template", func(ctx SpecContext) {
			result, err := persister.Get(ctx, 2)

			Expect(err).NotTo(HaveOccurred())
			Expect(
				result,
			).To(Equal(&models.ChoreTemplateDetails{Id: 2, Name: "Dishes", Description: "Load the dishwasher"}))
		})
	})

	Context("with schedules linked to different templates", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chore_templates (id, name) values (1, 'Laundry'), (2, 'Kitchen')`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.Exec(
				`insert into schedules (id, name, chore_template_id, interval_days, deactivated_at) values
				(1, 'Weekly laundry', 1, 7, null),
				(2, 'Monthly laundry', 1, 30, null),
				(3, 'Previous laundry', 1, 14, '2026-10-01 00:00:00'),
				(4, 'Weekly kitchen', 2, 7, null)`,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable(
			"counts only schedules linked to the requested template",
			func(ctx SpecContext, id, active, deactivated int) {
				result, err := persister.Get(ctx, id)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ActiveScheduleCount).To(Equal(active))
				Expect(result.DeactivatedScheduleCount).To(Equal(deactivated))
			},
			Entry("mixed states", 1, 2, 1),
			Entry("another template", 2, 1, 0),
		)

		It("reflects schedule deactivation in the counts", func(ctx SpecContext) {
			_, err := db.ExecContext(
				ctx,
				`update schedules set deactivated_at = '2026-10-02 00:00:00' where id = 1`,
			)
			Expect(err).NotTo(HaveOccurred())
			result, err := persister.Get(ctx, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.ActiveScheduleCount).To(Equal(1))
			Expect(result.DeactivatedScheduleCount).To(Equal(2))
		})

		It("keeps historical counts for a deactivated template", func(ctx SpecContext) {
			_, err := db.ExecContext(
				ctx,
				`update schedules set deactivated_at = '2026-10-02 00:00:00' where chore_template_id = 1`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`update chore_templates set deactivated_at = '2026-10-02 00:00:00' where id = 1`,
			)
			Expect(err).NotTo(HaveOccurred())
			result, err := persister.Get(ctx, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.DeactivatedAt).NotTo(BeNil())
			Expect(result.ActiveScheduleCount).To(BeZero())
			Expect(result.DeactivatedScheduleCount).To(Equal(3))
		})
	})
})
