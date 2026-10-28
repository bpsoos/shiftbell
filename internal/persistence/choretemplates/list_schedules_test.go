package choretemplates_test

import (
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	persistence "github.com/bpsoos/shiftbell/internal/persistence/choretemplates"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List template schedules", func() {
	var db *sqlx.DB
	var persister *persistence.Persister
	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = persistence.NewChoreTemplatePersister(
			&persistence.PersisterDeps{Db: db},
		)
		_, err := db.Exec(
			`insert into chore_templates (id, name) values (1, 'Laundry'), (2, 'Kitchen'), (3, 'Unused')`,
		)
		Expect(err).NotTo(HaveOccurred())
		_, err = db.Exec(
			`insert into schedules (id, name, chore_template_id, interval_days, deactivated_at) values
			(1, 'Zulu laundry', 1, 7, null), (2, 'alpha laundry', 1, 14, null),
			(3, 'Archived laundry', 1, 30, '2026-10-01 00:00:00'), (4, 'Other template', 2, 7, null)`,
		)
		Expect(err).NotTo(HaveOccurred())
	})
	DescribeTable(
		"filters by template and state in name order",
		func(ctx SpecContext, templateID int, state schedulemodels.ScheduleFilter, expected []models.ScheduleReference) {
			result, err := persister.ListSchedules(ctx, templateID, state)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		},
		Entry(
			"active",
			1,
			schedulemodels.ScheduleFilterActive,
			[]models.ScheduleReference{
				{Id: 2, Name: "alpha laundry"},
				{Id: 1, Name: "Zulu laundry"},
			},
		),
		Entry(
			"deactivated",
			1,
			schedulemodels.ScheduleFilterDeactivated,
			[]models.ScheduleReference{{Id: 3, Name: "Archived laundry"}},
		),
		Entry(
			"empty",
			3,
			schedulemodels.ScheduleFilterActive,
			[]models.ScheduleReference{},
		),
	)
})
