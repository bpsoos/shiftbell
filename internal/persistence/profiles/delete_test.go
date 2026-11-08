package profiles_test

import (
	"database/sql"

	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	"github.com/bpsoos/shiftbell/internal/persistence/profiles"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete", func() {
	It(
		"unlinks active and completed chores while preserving chores and other profiles",
		func(ctx SpecContext) {
			db := sqlitetest.NewMigratedDB()
			_, err := db.ExecContext(
				ctx,
				`INSERT INTO profiles (id, name, name_key) VALUES
			(7, 'Alex', 'alex'), (8, 'Sam', 'sam')`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`INSERT INTO chores (id, name, deadline, is_complete, completed_on, owner_profile_id) VALUES
			(1, 'Active', '2026-01-01', false, NULL, 7),
			(2, 'Completed', '2026-01-02', true, '2026-01-03', 7),
			(3, 'Other owner', '2026-01-04', false, NULL, 8),
			(4, 'Unassigned', '2026-01-05', false, NULL, NULL)`,
			)
			Expect(err).NotTo(HaveOccurred())
			persister := profiles.NewPersister(&profiles.Deps{Db: db})

			Expect(persister.Delete(ctx, 7)).To(Succeed())

			remaining, err := persister.List(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(remaining).To(Equal([]models.Profile{{Id: 8, Name: "Sam"}}))
			type chore struct {
				Id          int            `db:"id"`
				Name        string         `db:"name"`
				Deadline    string         `db:"deadline"`
				Complete    bool           `db:"is_complete"`
				CompletedOn sql.NullString `db:"completed_on"`
				Owner       sql.NullInt64  `db:"owner_profile_id"`
			}
			var chores []chore
			Expect(
				db.SelectContext(
					ctx,
					&chores,
					`SELECT id, name, date(deadline) AS deadline, is_complete, date(completed_on) AS completed_on, owner_profile_id FROM chores ORDER BY id`,
				),
			).To(Succeed())
			Expect(chores).To(Equal([]chore{
				{Id: 1, Name: "Active", Deadline: "2026-01-01"},
				{
					Id:          2,
					Name:        "Completed",
					Deadline:    "2026-01-02",
					Complete:    true,
					CompletedOn: sql.NullString{String: "2026-01-03", Valid: true},
				},
				{
					Id:       3,
					Name:     "Other owner",
					Deadline: "2026-01-04",
					Owner:    sql.NullInt64{Int64: 8, Valid: true},
				},
				{Id: 4, Name: "Unassigned", Deadline: "2026-01-05"},
			}))
		},
	)

	It("allows deleting the last profile", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := profiles.NewPersister(&profiles.Deps{Db: db})
		profile, err := persister.Create(ctx, &models.CreateParams{Name: "Alex"})
		Expect(err).NotTo(HaveOccurred())
		Expect(persister.Delete(ctx, profile.Id)).To(Succeed())
		exists, err := persister.HasProfiles(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeFalse())
	})

	It("reports a missing profile", func(ctx SpecContext) {
		persister := profiles.NewPersister(&profiles.Deps{Db: sqlitetest.NewMigratedDB()})
		Expect(persister.Delete(ctx, 7)).To(MatchError(models.ErrNotFound))
	})

	DescribeTable(
		"rolls back ownership changes on failure",
		func(ctx SpecContext, trigger string) {
			db := sqlitetest.NewMigratedDB()
			_, err := db.ExecContext(
				ctx,
				`INSERT INTO profiles (id, name, name_key) VALUES (7, 'Alex', 'alex')`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(
				ctx,
				`INSERT INTO chores (id, name, deadline, is_complete, owner_profile_id) VALUES (1, 'Kitchen', '2026-01-01', false, 7)`,
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.ExecContext(ctx, trigger)
			Expect(err).NotTo(HaveOccurred())
			persister := profiles.NewPersister(&profiles.Deps{Db: db})

			Expect(persister.Delete(ctx, 7)).To(HaveOccurred())

			remaining, err := persister.List(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(remaining).To(Equal([]models.Profile{{Id: 7, Name: "Alex"}}))
			var owner int
			Expect(
				db.GetContext(
					ctx,
					&owner,
					`SELECT owner_profile_id FROM chores WHERE id = 1`,
				),
			).To(Succeed())
			Expect(owner).To(Equal(7))
		},
		Entry(
			"profile deletion fails",
			`CREATE TRIGGER reject_deletion BEFORE DELETE ON profiles BEGIN SELECT RAISE(ABORT, 'delete failed'); END`,
		),
		Entry(
			"unlinking fails",
			`CREATE TRIGGER reject_unlink BEFORE UPDATE OF owner_profile_id ON chores BEGIN SELECT RAISE(ABORT, 'unlink failed'); END`,
		),
	)
})
