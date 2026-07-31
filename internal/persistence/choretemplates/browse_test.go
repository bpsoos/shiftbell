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

var _ = Describe("Browse", func() {
	var (
		db            *sqlx.DB
		persister     *choretemplatespersistence.Persister
		deactivatedAt time.Time
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = choretemplatespersistence.NewChoreTemplatePersister(
			&choretemplatespersistence.PersisterDeps{Db: db},
		)
		deactivatedAt = time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)

		_, err := db.Exec(
			`insert into chore_templates (id, name, description, deactivated_at) values
				(?, ?, ?, ?),
				(?, ?, ?, ?),
				(?, ?, ?, ?),
				(?, ?, ?, ?)`,
			1, "Laundry", "Wash and fold clothes", nil,
			2, "Dishes", "Load the dishwasher", nil,
			3, "Floors", nil, nil,
			4, "Archived", "No longer used", deactivatedAt,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns active chore templates newest first and reports another page", func(ctx SpecContext) {
		page, err := persister.Browse(ctx, &models.BrowseChoreTemplatesParams{
			Filter: models.ChoreTemplateFilterActive,
			Limit:  2,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(page).To(Equal(&models.ChoreTemplatePage{
			ChoreTemplates: []models.ChoreTemplate{
				{Id: 3, Name: "Floors"},
				{Id: 2, Name: "Dishes", Description: "Load the dishwasher"},
			},
			More: true,
		}))
	})

	It("applies the offset and reports the final active page", func(ctx SpecContext) {
		page, err := persister.Browse(ctx, &models.BrowseChoreTemplatesParams{
			Filter: models.ChoreTemplateFilterActive,
			Offset: 2,
			Limit:  2,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(page).To(Equal(&models.ChoreTemplatePage{
			ChoreTemplates: []models.ChoreTemplate{
				{Id: 1, Name: "Laundry", Description: "Wash and fold clothes"},
			},
			More: false,
		}))
	})

	It("returns deactivated chore templates with their deactivation time", func(ctx SpecContext) {
		page, err := persister.Browse(ctx, &models.BrowseChoreTemplatesParams{
			Filter: models.ChoreTemplateFilterDeactivated,
			Limit:  10,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(page).To(Equal(&models.ChoreTemplatePage{
			ChoreTemplates: []models.ChoreTemplate{
				{
					Id:            4,
					Name:          "Archived",
					Description:   "No longer used",
					DeactivatedAt: &deactivatedAt,
				},
			},
			More: false,
		}))
	})

	It("searches chore template names case-insensitively", func(ctx SpecContext) {
		page, err := persister.Browse(ctx, &models.BrowseChoreTemplatesParams{
			Filter: models.ChoreTemplateFilterActive,
			Search: "dIsH",
			Limit:  10,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(page).To(Equal(&models.ChoreTemplatePage{
			ChoreTemplates: []models.ChoreTemplate{
				{Id: 2, Name: "Dishes", Description: "Load the dishwasher"},
			},
			More: false,
		}))
	})

	It("searches chore template descriptions case-insensitively", func(ctx SpecContext) {
		page, err := persister.Browse(ctx, &models.BrowseChoreTemplatesParams{
			Filter: models.ChoreTemplateFilterActive,
			Search: "FoLd ClOtHeS",
			Limit:  10,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(page).To(Equal(&models.ChoreTemplatePage{
			ChoreTemplates: []models.ChoreTemplate{
				{Id: 1, Name: "Laundry", Description: "Wash and fold clothes"},
			},
			More: false,
		}))
	})
})
