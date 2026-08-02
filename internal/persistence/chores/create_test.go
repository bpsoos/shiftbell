package chores_test

import (
	"database/sql"
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {
	var (
		db        *sqlx.DB
		persister *chorespersistence.Persister
		deadline  time.Time
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)
		deadline = time.Date(2026, time.July, 20, 0, 0, 0, 0, time.UTC)
	})

	Context("with no description", func() {
		It("persists a chore with a null description", func() {
			result, err := persister.Create(&models.CreateOneOffChoreParams{
				Name:     "Laundry",
				Deadline: deadline,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(&models.Chore{
				Id:          1,
				Name:        "Laundry",
				Status:      models.ChoreStatusActive,
				Description: "",
				Deadline:    deadline,
			}))

			var name string
			var description sql.NullString
			var isCompleted bool
			var persistedDeadline time.Time
			Expect(db.QueryRow(
				`select name, description, is_complete, deadline from chores where id = ?`,
				result.Id,
			).Scan(&name, &description, &isCompleted, &persistedDeadline)).To(Succeed())
			Expect(name).To(Equal("Laundry"))
			Expect(description.Valid).To(BeFalse())
			Expect(isCompleted).To(BeFalse())
			Expect(persistedDeadline).To(Equal(deadline))
		})
	})

	Context("with one description", func() {
		It("persists the description", func() {
			result, err := persister.Create(&models.CreateOneOffChoreParams{
				Name:        "Laundry",
				Description: "Wash and fold clothes",
				Deadline:    deadline,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Description).To(Equal("Wash and fold clothes"))

			var description string
			Expect(
				db.QueryRow(`select description from chores where id = ?`, result.Id).
					Scan(&description),
			).To(Succeed())
			Expect(description).To(Equal("Wash and fold clothes"))
		})
	})

	Context("with many chores", func() {
		It("persists every chore", func() {
			first, err := persister.Create(&models.CreateOneOffChoreParams{
				Name:        "Laundry",
				Description: "Wash and fold clothes",
				Deadline:    deadline,
			})
			Expect(err).NotTo(HaveOccurred())

			second, err := persister.Create(&models.CreateOneOffChoreParams{
				Name:        "Dishes",
				Description: "Load the dishwasher",
				Deadline:    deadline.Add(24 * time.Hour),
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(first.Id).To(Equal(1))
			Expect(second.Id).To(Equal(2))
			var count int
			Expect(db.QueryRow(`select count(*) from chores`).Scan(&count)).To(Succeed())
			Expect(count).To(Equal(2))
		})
	})

	Context("when the database write fails", func() {
		It("returns an error", func() {
			Expect(db.Close()).To(Succeed())

			result, err := persister.Create(&models.CreateOneOffChoreParams{
				Name:     "Laundry",
				Deadline: deadline,
			})

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("db exec inserting chore")))
		})
	})
})
