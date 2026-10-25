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

	When("searching schedule and template names", func() {
		BeforeEach(func(ctx SpecContext) {
			_, err := db.ExecContext(
				ctx,
				`insert into chore_templates (id, name, description) values
				(1, 'Reusable kitchen routine', 'Secret description'),
				(2, 'Bathroom', 'Other description'), (3, 'Percent%_ template', null)`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`insert into schedules (id, name, chore_template_id, interval_days, deactivated_at) values
				(1, 'Morning kitchen', 1, 7, null), (2, 'Kitchen evening', 2, 14, null),
				(3, 'Archived rotation', 1, 30, '2020-03-27'),
				(4, 'Unrelated routine', 2, 7, null), (5, 'Percent%_ schedule', 3, 7, null)`,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable(
			"matches literal substrings within the selected state",
			func(ctx SpecContext, search string, filter models.ScheduleFilter, expectedIDs []int) {
				page, err := persister.Browse(
					ctx,
					&models.BrowseSchedulesParams{
						Search: search,
						Filter: filter,
						Limit:  20,
					},
				)

				Expect(err).NotTo(HaveOccurred())
				ids := make([]int, 0, len(page.Schedules))
				for _, schedule := range page.Schedules {
					ids = append(ids, schedule.Id)
				}
				Expect(ids).To(Equal(expectedIDs))
				Expect(page.More).To(BeFalse())
			},
			Entry(
				"schedule name with different case",
				"CHEN EVE",
				models.ScheduleFilterActive,
				[]int{2},
			),
			Entry(
				"template name with different case",
				"ABLE KIT",
				models.ScheduleFilterActive,
				[]int{1},
			),
			Entry(
				"both names without duplicates",
				"KITCHEN",
				models.ScheduleFilterActive,
				[]int{2, 1},
			),
			Entry(
				"deactivated schedule by template name",
				"ABLE KIT",
				models.ScheduleFilterDeactivated,
				[]int{3},
			),
			Entry(
				"deactivated schedule by schedule name",
				"IVED ROT",
				models.ScheduleFilterDeactivated,
				[]int{3},
			),
			Entry(
				"active matches excluded from deactivated state",
				"EVENING",
				models.ScheduleFilterDeactivated,
				[]int{},
			),
			Entry("empty search", "", models.ScheduleFilterActive, []int{5, 4, 2, 1}),
			Entry("no match", "missing", models.ScheduleFilterActive, []int{}),
			Entry(
				"descriptions excluded",
				"Secret",
				models.ScheduleFilterActive,
				[]int{},
			),
			Entry("percent is literal", "%", models.ScheduleFilterActive, []int{5}),
			Entry("underscore is literal", "_", models.ScheduleFilterActive, []int{5}),
			Entry(
				"SQL syntax is literal",
				"' OR 1=1 --",
				models.ScheduleFilterActive,
				[]int{},
			),
		)

		It("filters before ordering and pagination", func(ctx SpecContext) {
			first, err := persister.Browse(
				ctx,
				&models.BrowseSchedulesParams{
					Search: "KITCHEN",
					Filter: models.ScheduleFilterActive,
					Limit:  1,
				},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(first.Schedules).To(HaveLen(1))
			Expect(first.Schedules[0].Id).To(Equal(2))
			Expect(first.More).To(BeTrue())
			second, err := persister.Browse(
				ctx,
				&models.BrowseSchedulesParams{
					Search: "KITCHEN",
					Filter: models.ScheduleFilterActive,
					Limit:  1,
					Offset: 1,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(second.Schedules).To(HaveLen(1))
			Expect(second.Schedules[0].Id).To(Equal(1))
			Expect(second.More).To(BeFalse())
		})
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
