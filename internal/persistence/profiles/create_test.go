package profiles_test

import (
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	"github.com/bpsoos/shiftbell/internal/persistence/profiles"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {
	var db *sqlx.DB
	var persister *profiles.Persister
	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = profiles.NewPersister(&profiles.Deps{Db: db})
	})

	It("persists the display name and a stable identity", func(ctx SpecContext) {
		created, err := persister.Create(ctx, &models.CreateParams{Name: "Álex"})
		Expect(err).NotTo(HaveOccurred())
		Expect(created.Id).To(BeNumerically(">", 0))
		Expect(created.Name).To(Equal("Álex"))
		var name string
		Expect(
			db.GetContext(
				ctx,
				&name,
				`SELECT name FROM profiles WHERE id = ?`,
				created.Id,
			),
		).To(Succeed())
		Expect(name).To(Equal("Álex"))
	})

	DescribeTable(
		"enforces Unicode caseless uniqueness",
		func(ctx SpecContext, first, second string) {
			_, err := persister.Create(ctx, &models.CreateParams{Name: first})
			Expect(err).NotTo(HaveOccurred())
			created, err := persister.Create(ctx, &models.CreateParams{Name: second})
			Expect(created).To(BeNil())
			Expect(err).To(MatchError(models.ErrNameConflict))
			var count int
			Expect(
				db.GetContext(ctx, &count, `SELECT count(*) FROM profiles`),
			).To(Succeed())
			Expect(count).To(Equal(1))
		},
		Entry("ASCII", "Alex", "alex"),
		Entry("accents", "ÁLEX", "álex"),
		Entry("canonical equivalents", "Álex", "A\u0301lex"),
		Entry("case expansion", "Straße", "STRASSE"),
		Entry("Greek sigma", "ΟΣ", "ος"),
	)

	It("enforces the key constraint in SQLite", func(ctx SpecContext) {
		_, err := persister.Create(ctx, &models.CreateParams{Name: "Alex"})
		Expect(err).NotTo(HaveOccurred())
		_, err = db.ExecContext(
			ctx,
			`INSERT INTO profiles (name, name_key) VALUES ('ALEX', 'alex')`,
		)
		Expect(err).To(HaveOccurred())
	})

	It("propagates database failures", func(ctx SpecContext) {
		_, err := db.ExecContext(ctx, `DROP TABLE profiles`)
		Expect(err).NotTo(HaveOccurred())
		created, err := persister.Create(ctx, &models.CreateParams{Name: "Alex"})
		Expect(created).To(BeNil())
		Expect(err).To(HaveOccurred())
	})
})
