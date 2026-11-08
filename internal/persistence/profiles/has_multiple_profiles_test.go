package profiles_test

import (
	"github.com/bpsoos/shiftbell/internal/persistence/profiles"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HasMultipleProfiles", func() {
	It(
		"detects the second profile and reflects subsequent removal",
		func(ctx SpecContext) {
			db := sqlitetest.NewMigratedDB()
			persister := profiles.NewPersister(
				&profiles.Deps{Db: db},
				&profiles.Config{MaxProfiles: 2},
			)
			multiple, err := persister.HasMultipleProfiles(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(multiple).To(BeFalse())

			_, err = db.ExecContext(
				ctx,
				`INSERT INTO profiles (id, name, name_key) VALUES (7, 'Alex', 'alex')`,
			)
			Expect(err).NotTo(HaveOccurred())
			multiple, err = persister.HasMultipleProfiles(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(multiple).To(BeFalse())

			_, err = db.ExecContext(
				ctx,
				`INSERT INTO profiles (id, name, name_key) VALUES (42, 'Sam', 'sam')`,
			)
			Expect(err).NotTo(HaveOccurred())
			multiple, err = persister.HasMultipleProfiles(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(multiple).To(BeTrue())

			_, err = db.ExecContext(ctx, `DELETE FROM profiles WHERE id = 7`)
			Expect(err).NotTo(HaveOccurred())
			multiple, err = persister.HasMultipleProfiles(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(multiple).To(BeFalse())
		},
	)

	It("propagates database failures", func(ctx SpecContext) {
		db := sqlitetest.NewMigratedDB()
		_, err := db.ExecContext(ctx, `DROP TABLE profiles`)
		Expect(err).NotTo(HaveOccurred())
		persister := profiles.NewPersister(
			&profiles.Deps{Db: db},
			&profiles.Config{MaxProfiles: 2},
		)
		_, err = persister.HasMultipleProfiles(ctx)
		Expect(err).To(HaveOccurred())
	})
})
