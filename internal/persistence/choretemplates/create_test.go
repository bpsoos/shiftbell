package choretemplates_test

import (
	"database/sql"

	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	choretemplatespersistence "github.com/bpsoos/shiftbell/internal/persistence/choretemplates"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {
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

	Context("with no description", func() {
		It("persists a null description", func(ctx SpecContext) {
			created, err := persister.Create(
				ctx,
				&models.CreateChoreTemplateParams{Name: "Laundry"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(created).To(Equal(&models.ChoreTemplate{Id: 1, Name: "Laundry"}))

			var name string
			var description sql.NullString
			Expect(
				db.QueryRow(`select name, description from chore_templates`).
					Scan(&name, &description),
			).To(Succeed())
			Expect(name).To(Equal("Laundry"))
			Expect(description.Valid).To(BeFalse())
		})
	})

	Context("with one description", func() {
		It("persists the name and description", func(ctx SpecContext) {
			created, err := persister.Create(ctx, &models.CreateChoreTemplateParams{
				Name:        "Laundry",
				Description: "Wash and fold clothes",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(
				created,
			).To(Equal(&models.ChoreTemplate{Id: 1, Name: "Laundry", Description: "Wash and fold clothes"}))

			var name string
			var description string
			Expect(
				db.QueryRow(`select name, description from chore_templates`).
					Scan(&name, &description),
			).To(Succeed())
			Expect(name).To(Equal("Laundry"))
			Expect(description).To(Equal("Wash and fold clothes"))
		})
	})

	Context("with many chore templates", func() {
		It("persists every chore template", func(ctx SpecContext) {
			_, err := persister.Create(
				ctx,
				&models.CreateChoreTemplateParams{
					Name:        "Laundry",
					Description: "Wash and fold clothes",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = persister.Create(
				ctx,
				&models.CreateChoreTemplateParams{
					Name:        "Dishes",
					Description: "Load the dishwasher",
				},
			)
			Expect(err).NotTo(HaveOccurred())

			var count int
			Expect(
				db.QueryRow(`select count(*) from chore_templates where name in (?, ?)`, "Laundry", "Dishes").
					Scan(&count),
			).To(Succeed())
			Expect(count).To(Equal(2))
		})
	})
})
