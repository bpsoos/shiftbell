package chores_test

import (
	"time"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Complete chore", func() {
	It(
		"marks a one-off chore completed and returns its updated state",
		func(ctx SpecContext) {
			db := sqlitetest.NewMigratedDB()
			persister := chorespersistence.NewPersister(
				&chorespersistence.PersisterDeps{Db: db},
			)
			deadline := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
			completedOn := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			_, err := db.ExecContext(ctx, `
			insert into chores (id, name, description, is_complete, completed_on, deadline)
			values (42, 'Kitchen', 'Clean counters', false, null, ?)
		`, deadline)
			Expect(err).NotTo(HaveOccurred())

			result, err := persister.Complete(ctx, &choremodels.CompleteChoreParams{
				Id:          42,
				CompletedOn: completedOn,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(&choremodels.CompleteChoreResult{
				Chore: &choremodels.Chore{
					Id:          42,
					Status:      choremodels.ChoreStatusCompleted,
					Name:        "Kitchen",
					Description: "Clean counters",
					Deadline:    deadline,
					CompletedOn: completedOn,
				},
			}))
		},
	)

	It("preserves the original completion date when repeated", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)
		deadline := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
		originalDate := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
		differentDate := time.Date(2020, time.February, 4, 0, 0, 0, 0, time.UTC)
		_, err := db.ExecContext(ctx, `
			insert into chores (id, name, description, is_complete, completed_on, deadline)
			values (42, 'Kitchen', 'Clean counters', false, null, ?)
		`, deadline)
		Expect(err).NotTo(HaveOccurred())

		first, err := persister.Complete(ctx, &choremodels.CompleteChoreParams{
			Id:          42,
			CompletedOn: originalDate,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(first.Chore.CompletedOn).To(Equal(originalDate))

		second, err := persister.Complete(ctx, &choremodels.CompleteChoreParams{
			Id:          42,
			CompletedOn: differentDate,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(second).To(Equal(first))

		persisted, err := persister.Get(ctx, 42)
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted).To(Equal(first.Chore))
	})
})
