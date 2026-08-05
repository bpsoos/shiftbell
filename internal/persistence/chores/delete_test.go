package chores_test

import (
	"time"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete chore", func() {
	DescribeTable(
		"deletes a one-off chore",
		func(ctx SpecContext, isComplete bool, completedOn any) {
			db := sqlitetest.NewMigratedDB()
			persister := chorespersistence.NewPersister(
				&chorespersistence.PersisterDeps{Db: db},
			)
			_, err := db.ExecContext(ctx, `
				insert into chores (id, name, is_complete, completed_on, deadline)
				values (42, 'Kitchen', ?, ?, '2020-02-01')
			`, isComplete, completedOn)
			Expect(err).NotTo(HaveOccurred())

			err = persister.Delete(ctx, 42)

			Expect(err).NotTo(HaveOccurred())
			_, err = persister.Get(ctx, 42)
			Expect(err).To(MatchError(choremodels.ErrNotFound))
		},
		Entry("when active", false, nil),
		Entry(
			"when completed",
			true,
			time.Date(2020, time.February, 2, 0, 0, 0, 0, time.UTC),
		),
	)

	It("returns not found when the chore does not exist", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)

		err := persister.Delete(ctx, 42)

		Expect(err).To(MatchError(choremodels.ErrNotFound))
	})
})
