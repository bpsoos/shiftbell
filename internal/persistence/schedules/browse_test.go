package schedules_test

import (
	"time"

	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/bpsoos/shiftbell/internal/persistence/schedules"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

var _ = Describe("Browse", func() {
	var (
		db        *sqlx.DB
		persister *schedules.Persister
	)
	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = schedules.NewPersister(&schedules.PersisterDeps{Db: db})
	})

	It("returns an empty page from a migrated database", func(ctx SpecContext) {
		page, err := persister.Browse(
			ctx,
			&models.BrowseSchedulesParams{Filter: models.ScheduleFilterActive, Limit: 20},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(page).To(Equal(&models.SchedulePage{Schedules: []models.Schedule{}}))
	})

	When("schedules exist", func() {
		BeforeEach(func() {
			_, err := db.Exec(
				`insert into chore_templates (id, name) values (1, 'Laundry')`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.Exec(
				`insert into schedules (id, name, chore_template_id, interval_days, deactivated_at) values
				(1, 'Weekly laundry', 1, 7, null), (2, 'Monthly laundry', 1, 30, null), (3, 'Old laundry', 1, 2, '2026-08-01 00:00:00')`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.Exec(
				`insert into chores (id, name, is_complete, deadline, schedule_id) values
				(1, 'Laundry', 0, '2026-08-10', 2), (2, 'Laundry', 1, '2026-07-10', 2)`,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It(
			"returns active schedules newest first with only their active chore",
			func(ctx SpecContext) {
				page, err := persister.Browse(
					ctx,
					&models.BrowseSchedulesParams{
						Filter: models.ScheduleFilterActive,
						Limit:  1,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				choreID := 1
				deadline := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
				Expect(
					page,
				).To(Equal(&models.SchedulePage{More: true, Schedules: []models.Schedule{
					{
						Id:                  2,
						Name:                "Monthly laundry",
						ChoreTemplateId:     1,
						ChoreTemplateName:   "Laundry",
						IntervalDays:        30,
						ActiveChoreId:       &choreID,
						ActiveChoreDeadline: &deadline,
					},
				}}))
			},
		)

		It(
			"returns the final page including schedules without an active chore",
			func(ctx SpecContext) {
				page, err := persister.Browse(
					ctx,
					&models.BrowseSchedulesParams{
						Filter: models.ScheduleFilterActive,
						Limit:  1,
						Offset: 1,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(page).To(Equal(&models.SchedulePage{Schedules: []models.Schedule{
					{
						Id:                1,
						Name:              "Weekly laundry",
						ChoreTemplateId:   1,
						ChoreTemplateName: "Laundry",
						IntervalDays:      7,
					},
				}}))
			},
		)

		It("filters by name without case sensitivity", func(ctx SpecContext) {
			page, err := persister.Browse(
				ctx,
				&models.BrowseSchedulesParams{
					Filter: models.ScheduleFilterActive,
					Limit:  20,
					Search: "WEEKLY",
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(page.Schedules).To(HaveLen(1))
			Expect(page.Schedules[0].Id).To(Equal(1))
		})

		It("returns deactivated schedules", func(ctx SpecContext) {
			page, err := persister.Browse(
				ctx,
				&models.BrowseSchedulesParams{
					Filter: models.ScheduleFilterDeactivated,
					Limit:  20,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(page.Schedules).To(HaveLen(1))
			Expect(page.Schedules[0].Id).To(Equal(3))
			Expect(
				page.Schedules[0].DeactivatedAt,
			).To(gstruct.PointTo(Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))))
		})

		DescribeTable(
			"paginates within the selected state",
			func(ctx SpecContext, filter models.ScheduleFilter, firstID, secondID int) {
				_, err := db.ExecContext(
					ctx,
					`insert into schedules (id, name, chore_template_id, interval_days, deactivated_at)
				values (4, 'Another old laundry', 1, 14, '2026-08-02 00:00:00')`,
				)
				Expect(err).NotTo(HaveOccurred())

				first, err := persister.Browse(
					ctx,
					&models.BrowseSchedulesParams{Filter: filter, Limit: 1},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(first.More).To(BeTrue())
				Expect(first.Schedules).To(HaveLen(1))
				Expect(first.Schedules[0].Id).To(Equal(firstID))
				second, err := persister.Browse(
					ctx,
					&models.BrowseSchedulesParams{Filter: filter, Limit: 1, Offset: 1},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(second.More).To(BeFalse())
				Expect(second.Schedules).To(HaveLen(1))
				Expect(second.Schedules[0].Id).To(Equal(secondID))
			},
			Entry("active", models.ScheduleFilterActive, 2, 1),
			Entry("deactivated", models.ScheduleFilterDeactivated, 4, 3),
		)
	})
})
