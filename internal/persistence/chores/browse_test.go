package chores_test

import (
	"time"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Browse service chores", func() {
	var (
		db        *sqlx.DB
		persister *chorespersistence.Persister
	)

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)
	})

	It("returns active chores ordered by deadline and ID", func(ctx SpecContext) {
		firstDeadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
		secondDeadline := time.Date(2020, time.February, 4, 0, 0, 0, 0, time.UTC)
		_, err := db.ExecContext(ctx, `
			insert into chores (id, name, description, is_complete, completed_on, deadline)
			values
				(1, 'Second', 'Second description', false, null, ?),
				(2, 'First', 'First description', false, null, ?),
				(3, 'Completed', null, true, ?, ?)
		`, secondDeadline, firstDeadline, secondDeadline, firstDeadline)
		Expect(err).NotTo(HaveOccurred())

		page, err := persister.Browse(ctx, &choremodels.BrowseChoresParams{
			Status: choremodels.ChoreStatusActive,
			Offset: 0,
			Limit:  20,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(page).To(Equal(&choremodels.ChorePage{
			Chores: []choremodels.Chore{
				{
					Id:          2,
					Status:      choremodels.ChoreStatusActive,
					Name:        "First",
					Description: "First description",
					Deadline:    firstDeadline,
				},
				{
					Id:          1,
					Status:      choremodels.ChoreStatusActive,
					Name:        "Second",
					Description: "Second description",
					Deadline:    secondDeadline,
				},
			},
		}))
	})
})
