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

var _ = Describe("Set chore assignment", func() {
	var db *sqlx.DB
	var persister *persistence.Persister

	BeforeEach(func(ctx SpecContext) {
		db = sqlitetest.NewMigratedDB()
		persister = persistence.NewPersister(&persistence.PersisterDeps{Db: db})
		_, err := db.ExecContext(
			ctx,
			`insert into profiles (id, name, name_key) values (1, 'Alex', 'alex'), (2, 'Sam', 'sam')`,
		)
		Expect(err).NotTo(HaveOccurred())
		_, err = db.ExecContext(
			ctx,
			`insert into chores (id, name, deadline, is_complete) values (42, 'Kitchen', ?, false)`,
			time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC),
		)
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable(
		"replaces ownership without changing chore details",
		func(ctx SpecContext, completed bool, previous *int, owner int, ownerName string) {
			_, err := db.ExecContext(
				ctx,
				`update chores set is_complete = ?, completed_on = ?, owner_profile_id = ? where id = 42`,
				completed,
				time.Date(2020, 2, 4, 0, 0, 0, 0, time.UTC),
				previous,
			)
			Expect(err).NotTo(HaveOccurred())
			original, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			updated, err := persister.SetAssignment(
				ctx,
				&models.SetAssignmentParams{Id: 42, OwnerProfileId: owner},
			)
			Expect(err).NotTo(HaveOccurred())
			expected := *original
			expected.OwnerProfileId = &owner
			expected.OwnerProfileName = ownerName
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
		Entry("assigns active", false, nil, 1, "Alex"),
		Entry("reassigns active", false, new(1), 2, "Sam"),
		Entry("retains active assignment", false, new(1), 1, "Alex"),
		Entry("assigns completed", true, nil, 1, "Alex"),
		Entry("reassigns completed", true, new(1), 2, "Sam"),
		Entry("retains completed assignment", true, new(1), 1, "Alex"),
	)

	It(
		"rejects unknown profiles without losing the existing assignment",
		func(ctx SpecContext) {
			original, err := persister.SetAssignment(
				ctx,
				&models.SetAssignmentParams{Id: 42, OwnerProfileId: 1},
			)
			Expect(err).NotTo(HaveOccurred())
			result, err := persister.SetAssignment(
				ctx,
				&models.SetAssignmentParams{Id: 42, OwnerProfileId: 999},
			)
			Expect(err).To(MatchError(models.ErrInvalidOwnerProfile))
			Expect(result).To(BeNil())
			persisted, err := persister.Get(ctx, 42)
			Expect(err).NotTo(HaveOccurred())
			Expect(persisted).To(Equal(original))
		},
	)

	It("returns not found for an absent chore", func(ctx SpecContext) {
		result, err := persister.SetAssignment(
			ctx,
			&models.SetAssignmentParams{Id: 999, OwnerProfileId: 1},
		)
		Expect(result).To(BeNil())
		Expect(err).To(MatchError(models.ErrNotFound))
	})

	It("enforces profile references in SQLite", func(ctx SpecContext) {
		_, err := db.ExecContext(
			ctx,
			`update chores set owner_profile_id = 999 where id = 42`,
		)
		Expect(err).To(HaveOccurred())
	})
})
