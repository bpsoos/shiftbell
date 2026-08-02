package choretemplates_test

import (
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	choretemplatespersistence "github.com/bpsoos/shiftbell/internal/persistence/choretemplates"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetBatch", func() {
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
		It("returns an empty batch", func() {
			result, err := persister.GetBatch(0, 1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.ChoreTemplates).To(BeEmpty())
			Expect(result.More).To(BeFalse())
		})
	})

	Context("with one chore template", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chore_templates (id, name, description) values (?, ?, ?)`,
				1,
				"Laundry",
				nil,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the chore template with more false", func() {
			result, err := persister.GetBatch(0, 1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.ChoreTemplates).To(HaveLen(1))
			Expect(
				result.ChoreTemplates[0],
			).To(Equal(models.ChoreTemplate{Id: 1, Name: "Laundry", Description: ""}))
			Expect(result.More).To(BeFalse())
		})
	})

	Context("with many chore templates", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chore_templates (id, name, description) values
					(?, ?, ?),
					(?, ?, ?),
					(?, ?, ?)`,
				1, "Laundry", "Wash and fold clothes",
				2, "Dishes", "Load the dishwasher",
				3, "Floors", "Vacuum the floors",
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the requested page with more true", func() {
			result, err := persister.GetBatch(1, 1)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.ChoreTemplates).To(HaveLen(1))
			Expect(
				result.ChoreTemplates[0],
			).To(Equal(models.ChoreTemplate{Id: 2, Name: "Dishes", Description: "Load the dishwasher"}))
			Expect(result.More).To(BeTrue())
		})

		It("returns the first page with more true", func() {
			result, err := persister.GetBatch(0, 2)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.ChoreTemplates).To(Equal([]models.ChoreTemplate{
				{Id: 3, Name: "Floors", Description: "Vacuum the floors"},
				{Id: 2, Name: "Dishes", Description: "Load the dishwasher"},
			}))
			Expect(result.More).To(BeTrue())
		})

		It("returns the final page with more false", func() {
			result, err := persister.GetBatch(2, 2)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.ChoreTemplates).To(Equal([]models.ChoreTemplate{
				{Id: 1, Name: "Laundry", Description: "Wash and fold clothes"},
			}))
			Expect(result.More).To(BeFalse())
		})

		It("returns an empty page beyond the end", func() {
			result, err := persister.GetBatch(3, 2)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.ChoreTemplates).To(BeEmpty())
			Expect(result.More).To(BeFalse())
		})
	})
})
