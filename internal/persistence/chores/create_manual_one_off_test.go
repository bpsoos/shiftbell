package chores_test

import (
	"database/sql"
	"time"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create a manual one-off chore", func() {
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

	It("persists and returns an active chore", func(ctx SpecContext) {
		deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)

		result, err := persister.CreateManualOneOff(
			ctx,
			&choremodels.CreateManualOneOffParams{
				Name:        "Kitchen",
				Description: "Wash and fold.",
				Deadline:    deadline,
			},
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(&choremodels.CreateChoreResult{
			Chore: &choremodels.Chore{
				Id:          1,
				Status:      choremodels.ChoreStatusActive,
				Name:        "Kitchen",
				Description: "Wash and fold.",
				Deadline:    deadline,
			},
		}))
		var (
			name        string
			description sql.NullString
			isComplete  bool
			persisted   time.Time
		)
		Expect(db.QueryRowContext(
			ctx,
			`select name, description, is_complete, deadline from chores where id = ?`,
			1,
		).Scan(&name, &description, &isComplete, &persisted)).To(Succeed())
		Expect(name).To(Equal("Kitchen"))
		Expect(
			description,
		).To(Equal(sql.NullString{String: "Wash and fold.", Valid: true}))
		Expect(isComplete).To(BeFalse())
		Expect(persisted).To(Equal(deadline))
	})
})
