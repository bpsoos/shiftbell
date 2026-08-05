package chores_test

import (
	"time"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	chorespersistence "github.com/bpsoos/shiftbell/internal/persistence/chores"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edit one-off chore", func() {
	It("updates the active chore snapshot and returns it", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)
		originalDeadline := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
		newDeadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
		_, err := db.ExecContext(ctx, `
			insert into chores (id, name, description, is_complete, completed_on, deadline)
			values (42, 'Kitchen', 'Clean counters', false, null, ?)
		`, originalDeadline)
		Expect(err).NotTo(HaveOccurred())

		result, err := persister.EditOneOff(ctx, &choremodels.EditOneOffChoreParams{
			Id:          42,
			Name:        "Kitchen edited",
			Description: "Clean all counters",
			Deadline:    newDeadline,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(&choremodels.ChoreDetails{
			Id:          42,
			Status:      choremodels.ChoreStatusActive,
			Name:        "Kitchen edited",
			Description: "Clean all counters",
			Deadline:    newDeadline,
		}))
		persisted, err := persister.Get(ctx, 42)
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted).To(Equal(result))
	})

	It("returns not found for a missing chore", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)

		result, err := persister.EditOneOff(ctx, &choremodels.EditOneOffChoreParams{
			Id:          42,
			Name:        "Kitchen edited",
			Description: "Clean all counters",
			Deadline:    time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC),
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(choremodels.ErrNotFound))
	})

	It("returns not found without changing a completed chore", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := chorespersistence.NewPersister(
			&chorespersistence.PersisterDeps{Db: db},
		)
		deadline := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
		completedOn := time.Date(2020, time.February, 2, 0, 0, 0, 0, time.UTC)
		_, err := db.ExecContext(ctx, `
			insert into chores (id, name, description, is_complete, completed_on, deadline)
			values (42, 'Kitchen', 'Clean counters', true, ?, ?)
		`, completedOn, deadline)
		Expect(err).NotTo(HaveOccurred())

		result, err := persister.EditOneOff(ctx, &choremodels.EditOneOffChoreParams{
			Id:          42,
			Name:        "Kitchen edited",
			Description: "Clean all counters",
			Deadline:    time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC),
		})

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(choremodels.ErrNotFound))
		persisted, err := persister.Get(ctx, 42)
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted).To(Equal(&choremodels.ChoreDetails{
			Id:          42,
			Status:      choremodels.ChoreStatusCompleted,
			Name:        "Kitchen",
			Description: "Clean counters",
			Deadline:    deadline,
			CompletedOn: completedOn,
		}))
	})
})
