package profiles_test

import (
	"github.com/bpsoos/shiftbell/internal/persistence/profiles"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HasProfiles", func() {
	It("distinguishes empty and populated databases", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		persister := profiles.NewPersister(
			&profiles.Deps{Db: db},
			&profiles.Config{MaxProfiles: 2},
		)
		exists, err := persister.HasProfiles(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeFalse())
		_, err = db.ExecContext(
			ctx,
			`INSERT INTO profiles (name, name_key) VALUES ('Alex', 'alex')`,
		)
		Expect(err).NotTo(HaveOccurred())
		exists, err = persister.HasProfiles(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeTrue())
	})

	It("propagates database failures", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		_, err := db.ExecContext(ctx, `DROP TABLE profiles`)
		Expect(err).NotTo(HaveOccurred())
		persister := profiles.NewPersister(
			&profiles.Deps{Db: db},
			&profiles.Config{MaxProfiles: 2},
		)
		_, err = persister.HasProfiles(ctx)
		Expect(err).To(HaveOccurred())
	})
})
