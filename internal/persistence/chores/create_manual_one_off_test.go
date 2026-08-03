package chores_test

import (
	"database/sql"
	"time"

	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
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

	It(
		"atomically persists the chore and a new template when requested",
		func(ctx SpecContext) {
			deadline := time.Date(2020, time.February, 4, 0, 0, 0, 0, time.UTC)

			result, err := persister.CreateManualOneOff(
				ctx,
				&choremodels.CreateManualOneOffParams{
					Name:                "Kitchen",
					Description:         "Reusable kitchen steps.",
					Deadline:            deadline,
					SaveAsChoreTemplate: true,
				},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(&choremodels.CreateChoreResult{
				Chore: &choremodels.Chore{
					Id:          1,
					Status:      choremodels.ChoreStatusActive,
					Name:        "Kitchen",
					Description: "Reusable kitchen steps.",
					Deadline:    deadline,
				},
				ChoreTemplate: &choretemplatemodels.ChoreTemplate{
					Id:          1,
					Name:        "Kitchen",
					Description: "Reusable kitchen steps.",
				},
			}))
			var choreTemplateId sql.NullInt64
			Expect(db.QueryRowContext(
				ctx,
				`select chore_template_id from chores where id = ?`,
				1,
			).Scan(&choreTemplateId)).To(Succeed())
			Expect(choreTemplateId.Valid).To(BeFalse())
			var name, description string
			Expect(db.QueryRowContext(
				ctx,
				`select name, description from chore_templates where id = ?`,
				1,
			).Scan(&name, &description)).To(Succeed())
			Expect(name).To(Equal("Kitchen"))
			Expect(description).To(Equal("Reusable kitchen steps."))
		},
	)

	It(
		"rolls back the chore when an active template name conflicts",
		func(ctx SpecContext) {
			_, err := db.ExecContext(
				ctx,
				`insert into chore_templates (id, name, description) values (?, ?, ?)`,
				7,
				"KITCHEN",
				"Existing template",
			)
			Expect(err).NotTo(HaveOccurred())
			deadline := time.Date(2020, time.February, 4, 0, 0, 0, 0, time.UTC)

			result, err := persister.CreateManualOneOff(
				ctx,
				&choremodels.CreateManualOneOffParams{
					Name:                "Kitchen",
					Description:         "Reusable kitchen steps.",
					Deadline:            deadline,
					SaveAsChoreTemplate: true,
				},
			)

			Expect(result).To(BeNil())
			Expect(err).To(MatchError(choretemplatemodels.ErrNameConflict))
			var choreCount, templateCount int
			Expect(
				db.QueryRowContext(ctx, `select count(*) from chores`).Scan(&choreCount),
			).To(Succeed())
			Expect(
				db.QueryRowContext(
					ctx,
					`select count(*) from chore_templates`,
				).Scan(&templateCount),
			).To(Succeed())
			Expect(choreCount).To(BeZero())
			Expect(templateCount).To(Equal(1))
		},
	)
})
