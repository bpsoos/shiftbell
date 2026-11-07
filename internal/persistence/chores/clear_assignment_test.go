package chores_test

import (
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/chores"
	persistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Clear chore assignment", func() {
	var db *sqlx.DB
	var persister *persistence.Persister

	BeforeEach(func(ctx SpecContext) {
		db = sqlitetest.NewMigratedDB()
		persister = persistence.NewPersister(&persistence.PersisterDeps{Db: db})
		_, err := db.ExecContext(
			ctx,
			`insert into profiles (id, name, name_key) values (1, 'Alex', 'alex')`,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable(
		"clears ownership without changing chore details",
		func(ctx SpecContext, completed bool, owner *int) {
			_, err := db.ExecContext(
				ctx,
				`insert into chores (id, name, deadline, is_complete, completed_on, owner_profile_id) values (42, 'Kitchen', ?, ?, ?, ?)`,
				time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC),
				completed,
				time.Date(2020, 2, 4, 0, 0, 0, 0, time.UTC),
				owner,
			)
			Expect(err).NotTo(HaveOccurred())
			original, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			updated, err := persister.ClearAssignment(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			expected := *original
			expected.OwnerProfileId = nil
			Expect(updated).To(Equal(&expected))
			persisted, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(persisted).To(Equal(&expected))
			page, err := persister.Browse(
				ctx,
				&models.BrowseChoresParams{Status: original.Status, Limit: 20},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(page.Chores).To(ConsistOf(expected))
		},
		Entry("assigned active", false, new(1)),
		Entry("unassigned active", false, nil),
		Entry("assigned completed", true, new(1)),
		Entry("unassigned completed", true, nil),
	)

	It("returns not found for an absent chore", func(ctx SpecContext) {
		result, err := persister.ClearAssignment(ctx, 999)
		Expect(result).To(BeNil())
		Expect(err).To(MatchError(models.ErrNotFound))
	})
})
