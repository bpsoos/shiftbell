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

var _ = Describe("Create a template-based one-off chore", func() {
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

	It("persists an independent snapshot of the active template", func(ctx SpecContext) {
		_, err := db.ExecContext(ctx, `
			insert into chore_templates (id, name, description)
			values (?, ?, ?)
		`, 42, "Kitchen", "Reusable template steps.")
		Expect(err).NotTo(HaveOccurred())
		deadline := time.Date(2020, time.February, 5, 0, 0, 0, 0, time.UTC)

		result, err := persister.CreateTemplateOneOff(
			ctx,
			&choremodels.CreateTemplateOneOffParams{
				ChoreTemplateId: 42,
				Deadline:        deadline,
			},
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(&choremodels.CreateChoreResult{
			Chore: &choremodels.Chore{
				Id:          1,
				Status:      choremodels.ChoreStatusActive,
				Name:        "Kitchen",
				Description: "Reusable template steps.",
				Deadline:    deadline,
			},
		}))
		var choreTemplateId sql.NullInt64
		Expect(db.QueryRowContext(
			ctx,
			`select chore_template_id from chores where id = ?`,
			1,
		).Scan(&choreTemplateId)).To(Succeed())
		Expect(choreTemplateId.Valid).To(BeFalse())
	})

	It("rejects a missing template without creating a chore", func(ctx SpecContext) {
		result, err := persister.CreateTemplateOneOff(
			ctx,
			&choremodels.CreateTemplateOneOffParams{
				ChoreTemplateId: 42,
				Deadline:        time.Date(2020, time.February, 5, 0, 0, 0, 0, time.UTC),
			},
		)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(choretemplatemodels.ErrNotFound))
		var count int
		Expect(
			db.QueryRowContext(ctx, `select count(*) from chores`).Scan(&count),
		).To(Succeed())
		Expect(count).To(BeZero())
	})

	It("rejects an inactive template without creating a chore", func(ctx SpecContext) {
		deactivatedAt := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
		_, err := db.ExecContext(ctx, `
			insert into chore_templates (id, name, description, deactivated_at)
			values (?, ?, ?, ?)
		`, 42, "Kitchen", "Reusable template steps.", deactivatedAt)
		Expect(err).NotTo(HaveOccurred())

		result, err := persister.CreateTemplateOneOff(
			ctx,
			&choremodels.CreateTemplateOneOffParams{
				ChoreTemplateId: 42,
				Deadline:        time.Date(2020, time.February, 5, 0, 0, 0, 0, time.UTC),
			},
		)

		Expect(result).To(BeNil())
		Expect(err).To(MatchError(choretemplatemodels.ErrInactive))
		var count int
		Expect(
			db.QueryRowContext(ctx, `select count(*) from chores`).Scan(&count),
		).To(Succeed())
		Expect(count).To(BeZero())
	})
})
