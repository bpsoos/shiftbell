package profiles_test

import (
	"errors"
	"sync"

	"github.com/bpsoos/shiftbell/internal/database"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	"github.com/bpsoos/shiftbell/internal/persistence/profiles"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateInitial", func() {
	It("creates exactly one initial profile across retries", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := profiles.NewPersister(
			&profiles.Deps{Db: db},
			&profiles.Config{MaxProfiles: 2},
		)
		created, err := persister.CreateInitial(ctx, &models.CreateParams{Name: "Alex"})
		Expect(err).NotTo(HaveOccurred())
		Expect(created).To(Equal(&models.Profile{Id: 1, Name: "Alex"}))
		second, err := persister.CreateInitial(ctx, &models.CreateParams{Name: "Sam"})
		Expect(second).To(BeNil())
		Expect(err).To(MatchError(models.ErrAlreadyInitialized))
		items, err := persister.List(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]models.Profile{*created}))
	})

	It("does not overwrite profiles created through management", func(ctx SpecContext) {
		persister := profiles.NewPersister(
			&profiles.Deps{Db: sqlitetest.NewMigratedDB()},
			&profiles.Config{MaxProfiles: 2},
		)
		original, err := persister.Create(ctx, &models.CreateParams{Name: "Sam"})
		Expect(err).NotTo(HaveOccurred())
		created, err := persister.CreateInitial(ctx, &models.CreateParams{Name: "Alex"})
		Expect(created).To(BeNil())
		Expect(err).To(MatchError(models.ErrAlreadyInitialized))
		items, err := persister.List(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]models.Profile{*original}))
	})

	It("folds the initial name for subsequent uniqueness checks", func(ctx SpecContext) {
		persister := profiles.NewPersister(
			&profiles.Deps{Db: sqlitetest.NewMigratedDB()},
			&profiles.Config{MaxProfiles: 2},
		)
		_, err := persister.CreateInitial(ctx, &models.CreateParams{Name: "Álex"})
		Expect(err).NotTo(HaveOccurred())
		_, err = persister.Create(ctx, &models.CreateParams{Name: "ÁLEX"})
		Expect(err).To(MatchError(models.ErrNameConflict))
	})

	It(
		"serializes competing setup requests across database connections",
		func(ctx SpecContext) {
			db := sqlitetest.NewMigratedDB()
			var filename string
			Expect(
				db.GetContext(
					ctx,
					&filename,
					`SELECT file FROM pragma_database_list WHERE name = 'main'`,
				),
			).To(Succeed())
			other, err := sqlx.Connect(
				"sqlite",
				database.SQLiteDSN(filename)+"&_pragma=busy_timeout(5000)",
			)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(other.Close)
			_, err = db.ExecContext(ctx, `PRAGMA busy_timeout = 5000`)
			Expect(err).NotTo(HaveOccurred())
			persisters := []*profiles.Persister{
				profiles.NewPersister(
					&profiles.Deps{Db: db},
					&profiles.Config{MaxProfiles: 2},
				),
				profiles.NewPersister(
					&profiles.Deps{Db: other},
					&profiles.Config{MaxProfiles: 2},
				),
			}
			results := make([]error, len(persisters))
			start := make(chan struct{})
			var workers sync.WaitGroup
			for i, persister := range persisters {
				workers.Go(func() {
					<-start
					_, results[i] = persister.CreateInitial(
						ctx,
						&models.CreateParams{Name: "Alex"},
					)
				})
			}
			close(start)
			workers.Wait()
			Expect(results).To(ConsistOf(BeNil(), WithTransform(func(err error) bool {
				return errors.Is(err, models.ErrAlreadyInitialized)
			}, BeTrue())))
			var count int
			Expect(
				db.GetContext(ctx, &count, `SELECT count(*) FROM profiles`),
			).To(Succeed())
			Expect(count).To(Equal(1))
		},
	)

	It("propagates database failures without inserting a profile", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		_, err := db.ExecContext(
			ctx,
			`CREATE TRIGGER reject_profiles BEFORE INSERT ON profiles BEGIN SELECT RAISE(ABORT, 'unavailable'); END`,
		)
		Expect(err).NotTo(HaveOccurred())
		persister := profiles.NewPersister(
			&profiles.Deps{Db: db},
			&profiles.Config{MaxProfiles: 2},
		)
		created, err := persister.CreateInitial(ctx, &models.CreateParams{Name: "Alex"})
		Expect(created).To(BeNil())
		Expect(err).To(HaveOccurred())
		var count int
		Expect(db.GetContext(ctx, &count, `SELECT count(*) FROM profiles`)).To(Succeed())
		Expect(count).To(BeZero())
	})
})
