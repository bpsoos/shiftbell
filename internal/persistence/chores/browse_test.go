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

	DescribeTable(
		"returns owner names alongside unassigned chores",
		func(ctx SpecContext, completed bool, status choremodels.ChoreStatus) {
			_, err := db.ExecContext(
				ctx,
				`insert into profiles (id, name, name_key) values (7, 'Alex', 'alex')`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`insert into chores (id, name, deadline, is_complete, completed_on, owner_profile_id) values
			(42, 'Kitchen assigned', '2026-01-01', ?, case when ? then '2026-01-02' else null end, 7),
			(43, 'Kitchen unassigned', '2026-01-01', ?, case when ? then '2026-01-02' else null end, null)`,
				completed,
				completed,
				completed,
				completed,
			)
			Expect(err).NotTo(HaveOccurred())
			page, err := persister.Browse(
				ctx,
				&choremodels.BrowseChoresParams{
					Status: status,
					Search: "Kitchen",
					Limit:  20,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(page.Chores).To(HaveLen(2))
			Expect(
				page.Chores,
			).To(ContainElement(And(HaveField("Id", 42), HaveField("OwnerProfileId", new(7)), HaveField("OwnerProfileName", "Alex"))))
			Expect(
				page.Chores,
			).To(ContainElement(And(HaveField("Id", 43), HaveField("OwnerProfileId", BeNil()), HaveField("OwnerProfileName", ""))))
		},
		Entry("active", false, choremodels.ChoreStatusActive),
		Entry("completed", true, choremodels.ChoreStatusCompleted),
	)

	It(
		"includes schedule and predecessor associations when browsing chores",
		func(ctx SpecContext) {
			_, err := db.ExecContext(
				ctx,
				`insert into chore_templates (id, name) values (1, 'Laundry')`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`insert into schedules (id, name, chore_template_id, interval_days) values (1, 'Weekly laundry', 1, 7)`,
			)
			Expect(err).NotTo(HaveOccurred())
			deadline := time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC)
			_, err = db.ExecContext(
				ctx,
				`insert into chores (id, name, is_complete, completed_on, deadline, schedule_id, predecessor_id)
			values (1, 'Laundry', true, '2020-02-02', ?, 1, null), (2, 'Laundry', false, null, ?, 1, 1)`,
				deadline,
				deadline,
			)
			Expect(err).NotTo(HaveOccurred())
			page, err := persister.Browse(
				ctx,
				&choremodels.BrowseChoresParams{
					Status: choremodels.ChoreStatusActive,
					Limit:  20,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			scheduleID, predecessorID := 1, 1
			Expect(page.Chores).To(ConsistOf(choremodels.Chore{
				Id:            2,
				ScheduleId:    &scheduleID,
				PredecessorId: &predecessorID,
				Name:          "Laundry",
				Status:        choremodels.ChoreStatusActive,
				Deadline:      deadline,
			}))
		},
	)
})
