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

var _ = Describe("Edit", func() {
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

	It("persists and returns the edited active template", func(ctx SpecContext) {
		_, err := db.ExecContext(
			ctx,
			`insert into chore_templates (id, name, description) values (?, ?, ?)`,
			42,
			"Original",
			"Original description",
		)
		Expect(err).NotTo(HaveOccurred())

		edited, err := persister.Edit(ctx, &models.EditChoreTemplateParams{
			Id:          42,
			Name:        "Edited template",
			Description: "Edited description",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(edited).To(Equal(&models.ChoreTemplate{
			Id:          42,
			Name:        "Edited template",
			Description: "Edited description",
		}))
		var name, description string
		Expect(db.QueryRowContext(
			ctx,
			`select name, description from chore_templates where id = ?`,
			42,
		).Scan(&name, &description)).To(Succeed())
		Expect(name).To(Equal("Edited template"))
		Expect(description).To(Equal("Edited description"))
	})

	It(
		"returns the existing active template when its name conflicts",
		func(ctx SpecContext) {
			_, err := db.ExecContext(
				ctx,
				`
					insert into chore_templates (id, name, description)
					values (?, ?, ?), (?, ?, ?)
				`,
				7,
				"Existing",
				"Existing description",
				42,
				"Target",
				"Target description",
			)
			Expect(err).NotTo(HaveOccurred())

			edited, err := persister.Edit(ctx, &models.EditChoreTemplateParams{
				Id:          42,
				Name:        "EXISTING",
				Description: "Conflicting edit",
			})

			Expect(edited).To(BeNil())
			Expect(err).To(MatchError(models.ErrNameConflict))
			var name, description string
			Expect(db.QueryRowContext(
				ctx,
				`select name, description from chore_templates where id = ?`,
				42,
			).Scan(&name, &description)).To(Succeed())
			Expect(name).To(Equal("Target"))
			Expect(description).To(Equal("Target description"))
		},
	)

	It("rejects editing a deactivated template", func(ctx SpecContext) {
		deactivatedAt := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
		_, err := db.ExecContext(
			ctx,
			`insert into chore_templates (id, name, deactivated_at) values (?, ?, ?)`,
			42,
			"Inactive",
			deactivatedAt,
		)
		Expect(err).NotTo(HaveOccurred())

		edited, err := persister.Edit(ctx, &models.EditChoreTemplateParams{
			Id:   42,
			Name: "Edited",
		})

		Expect(edited).To(BeNil())
		Expect(err).To(MatchError(models.ErrInactive))
	})

	It("rejects editing a missing template", func(ctx SpecContext) {
		edited, err := persister.Edit(ctx, &models.EditChoreTemplateParams{
			Id:   42,
			Name: "Edited",
		})

		Expect(edited).To(BeNil())
		Expect(err).To(MatchError(models.ErrNotFound))
	})
})
