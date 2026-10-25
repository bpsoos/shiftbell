package choretemplates_test

import (
	"errors"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	choretemplatespersistence "github.com/bpsoos/shiftbell/internal/persistence/choretemplates"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deactivate", func() {
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

	It("persists and returns the deactivated template", func(ctx SpecContext) {
		_, err := db.ExecContext(
			ctx,
			`insert into chore_templates (id, name, description) values (?, ?, ?)`,
			42,
			"Laundry",
			"Wash and fold",
		)
		Expect(err).NotTo(HaveOccurred())
		deactivatedAt := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)

		deactivated, err := persister.Deactivate(
			ctx,
			&models.DeactivateChoreTemplateParams{Id: 42, DeactivatedAt: deactivatedAt},
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(deactivated).To(Equal(&models.ChoreTemplate{
			Id:            42,
			Name:          "Laundry",
			Description:   "Wash and fold",
			DeactivatedAt: &deactivatedAt,
		}))
	})

	It("rejects deactivating an already deactivated template", func(ctx SpecContext) {
		deactivatedAt := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
		_, err := db.ExecContext(
			ctx,
			`insert into chore_templates (id, name, deactivated_at) values (?, ?, ?)`,
			42,
			"Inactive",
			deactivatedAt,
		)
		Expect(err).NotTo(HaveOccurred())

		deactivated, err := persister.Deactivate(
			ctx,
			&models.DeactivateChoreTemplateParams{Id: 42, DeactivatedAt: deactivatedAt},
		)

		Expect(deactivated).To(BeNil())
		Expect(err).To(MatchError(models.ErrInactive))
	})

	It("rejects deactivating a missing template", func(ctx SpecContext) {
		deactivated, err := persister.Deactivate(
			ctx,
			&models.DeactivateChoreTemplateParams{
				Id:            42,
				DeactivatedAt: time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC),
			},
		)

		Expect(deactivated).To(BeNil())
		Expect(err).To(MatchError(models.ErrNotFound))
	})

	When("schedules reference templates", func() {
		var deactivatedAt time.Time

		BeforeEach(func(ctx SpecContext) {
			deactivatedAt = time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
			_, err := db.ExecContext(
				ctx,
				`insert into chore_templates (id, name, description)
				values (42, 'Laundry', 'Wash and fold'), (43, 'Kitchen', 'Wash dishes')`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`insert into schedules (id, name, chore_template_id, interval_days, deactivated_at)
				values (1, 'Archived laundry', 42, 30, ?), (2, 'Weekly kitchen', 43, 7, null)`,
				deactivatedAt,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It(
			"reports all active references and leaves the template and dependent records unchanged",
			func(ctx SpecContext) {
				_, err := db.ExecContext(
					ctx,
					`insert into schedules (id, name, chore_template_id, interval_days)
				values (3, 'Weekly laundry', 42, 7), (4, 'Fortnightly laundry', 42, 14)`,
				)
				Expect(err).NotTo(HaveOccurred())
				_, err = db.ExecContext(
					ctx,
					`insert into chores (id, name, is_complete, deadline, schedule_id)
				values (1, 'Laundry', false, '2026-08-10', 3), (2, 'Laundry', false, '2026-08-17', 4)`,
				)
				Expect(err).NotTo(HaveOccurred())

				deactivated, err := persister.Deactivate(
					ctx,
					&models.DeactivateChoreTemplateParams{
						Id:            42,
						DeactivatedAt: deactivatedAt,
					},
				)

				Expect(deactivated).To(BeNil())
				var blocked *models.ActiveScheduleReferencesError
				Expect(errors.As(err, &blocked)).To(BeTrue())
				Expect(blocked.Schedules).To(Equal([]models.ActiveScheduleReference{
					{Id: 4, Name: "Fortnightly laundry"}, {Id: 3, Name: "Weekly laundry"},
				}))
				details, err := persister.Get(ctx, 42)
				Expect(err).NotTo(HaveOccurred())
				Expect(
					details.ChoreTemplate,
				).To(Equal(models.ChoreTemplate{Id: 42, Name: "Laundry", Description: "Wash and fold"}))
				var activeSchedules, activeChores int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from schedules where chore_template_id = 42 and deactivated_at is null`).
						Scan(&activeSchedules),
				).To(Succeed())
				Expect(
					db.QueryRowContext(ctx, `select count(*) from chores where schedule_id in (3, 4) and is_complete = false`).
						Scan(&activeChores),
				).To(Succeed())
				Expect(activeSchedules).To(Equal(2))
				Expect(activeChores).To(Equal(2))
			},
		)

		It(
			"blocks an active schedule even when it has no active chore",
			func(ctx SpecContext) {
				deactivated, err := persister.Deactivate(
					ctx,
					&models.DeactivateChoreTemplateParams{
						Id:            43,
						DeactivatedAt: deactivatedAt,
					},
				)

				Expect(deactivated).To(BeNil())
				var blocked *models.ActiveScheduleReferencesError
				Expect(errors.As(err, &blocked)).To(BeTrue())
				Expect(
					blocked.Schedules,
				).To(ConsistOf(models.ActiveScheduleReference{Id: 2, Name: "Weekly kitchen"}))
			},
		)

		It(
			"allows deactivation when only deactivated schedules reference the template",
			func(ctx SpecContext) {
				deactivated, err := persister.Deactivate(
					ctx,
					&models.DeactivateChoreTemplateParams{
						Id:            42,
						DeactivatedAt: deactivatedAt,
					},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(deactivated).To(Equal(&models.ChoreTemplate{
					Id:            42,
					Name:          "Laundry",
					Description:   "Wash and fold",
					DeactivatedAt: &deactivatedAt,
				}))
				details, err := persister.Get(ctx, 42)
				Expect(err).NotTo(HaveOccurred())
				Expect(details.ChoreTemplate).To(Equal(*deactivated))
				var scheduleCount int
				Expect(
					db.QueryRowContext(ctx, `select count(*) from schedules where id = 1 and deactivated_at is not null`).
						Scan(&scheduleCount),
				).To(Succeed())
				Expect(scheduleCount).To(Equal(1))
			},
		)
	})
})
