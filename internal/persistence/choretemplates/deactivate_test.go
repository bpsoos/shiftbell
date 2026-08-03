package choretemplates_test

import (
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
})
