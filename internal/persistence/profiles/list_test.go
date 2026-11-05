package profiles_test

import (
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	"github.com/bpsoos/shiftbell/internal/persistence/profiles"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List", func() {
	It("returns an empty non-nil list", func(ctx SpecContext) {
		persister := profiles.NewPersister(&profiles.Deps{Db: sqlitetest.NewMigratedDB()})
		items, err := persister.List(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]models.Profile{}))
	})

	It("lists profiles in creation order", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		_, err := db.ExecContext(
			ctx,
			`INSERT INTO profiles (id, name, name_key) VALUES (8, 'Alex', 'alex'), (3, 'Zoe', 'zoe')`,
		)
		Expect(err).NotTo(HaveOccurred())
		persister := profiles.NewPersister(&profiles.Deps{Db: db})
		items, err := persister.List(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(
			items,
		).To(Equal([]models.Profile{{Id: 3, Name: "Zoe"}, {Id: 8, Name: "Alex"}}))
	})

	It("propagates database failures", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		_, err := db.ExecContext(ctx, `DROP TABLE profiles`)
		Expect(err).NotTo(HaveOccurred())
		persister := profiles.NewPersister(&profiles.Deps{Db: db})
		items, err := persister.List(ctx)
		Expect(items).To(BeNil())
		Expect(err).To(HaveOccurred())
	})
})
