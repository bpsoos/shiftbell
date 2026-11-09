package accounts_test

import (
	"context"

	models "github.com/bpsoos/shiftbell/internal/models/accounts"
	"github.com/bpsoos/shiftbell/internal/persistence/accounts"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {
	var db *sqlx.DB
	var persister *accounts.Persister
	var values *models.CreateValues

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = accounts.NewPersister(&accounts.Deps{Db: db})
		values = &models.CreateValues{
			Username:     "O'Connor",
			PasswordHash: "$argon2id$v=19$m=65536,t=3,p=1$c2FsdA$aGFzaA",
		}
	})

	It("persists the username and complete encoded hash verbatim", func(ctx SpecContext) {
		created, err := persister.Create(ctx, values)
		Expect(err).NotTo(HaveOccurred())
		Expect(created.Id).To(BeNumerically(">", 0))
		Expect(created.Username).To(Equal(values.Username))

		var stored struct {
			Id           int    `db:"id"`
			Username     string `db:"username"`
			PasswordHash string `db:"password_hash"`
		}
		Expect(db.GetContext(
			ctx,
			&stored,
			`SELECT id, username, password_hash FROM accounts WHERE id = ?`,
			created.Id,
		)).To(Succeed())
		Expect(stored.Id).To(Equal(created.Id))
		Expect(stored.Username).To(Equal(values.Username))
		Expect(stored.PasswordHash).To(Equal(values.PasswordHash))
	})

	DescribeTable(
		"rejects a second account without changing existing credentials",
		func(ctx SpecContext, username string) {
			first, err := persister.Create(ctx, values)
			Expect(err).NotTo(HaveOccurred())

			created, err := persister.Create(ctx, &models.CreateValues{
				Username:     username,
				PasswordHash: "replacement-hash",
			})
			Expect(created).To(BeNil())
			Expect(err).To(MatchError(models.ErrAlreadyExists))
			var stored []models.Account
			Expect(
				db.SelectContext(ctx, &stored, `SELECT id, username FROM accounts`),
			).To(Succeed())
			Expect(stored).To(Equal([]models.Account{*first}))
			var hash string
			Expect(
				db.GetContext(ctx, &hash, `SELECT password_hash FROM accounts`),
			).To(Succeed())
			Expect(hash).To(Equal(values.PasswordHash))
		},
		Entry("same username", "O'Connor"),
		Entry("different username", "Another owner"),
	)

	It("gives a replacement account a new identity", func(ctx SpecContext) {
		first, err := persister.Create(ctx, values)
		Expect(err).NotTo(HaveOccurred())
		_, err = db.ExecContext(ctx, `DELETE FROM accounts WHERE id = ?`, first.Id)
		Expect(err).NotTo(HaveOccurred())

		replacement, err := persister.Create(ctx, values)
		Expect(err).NotTo(HaveOccurred())
		Expect(replacement.Id).To(BeNumerically(">", first.Id))
	})

	It("preserves cancellation without inserting an account", func(ctx SpecContext) {
		cancelled, cancel := context.WithCancel(ctx)
		cancel()

		created, err := persister.Create(cancelled, values)
		Expect(created).To(BeNil())
		Expect(err).To(MatchError(context.Canceled))
		var count int
		Expect(db.GetContext(ctx, &count, `SELECT count(*) FROM accounts`)).To(Succeed())
		Expect(count).To(BeZero())
	})

	It(
		"propagates database failures without reporting an account conflict",
		func(ctx SpecContext) {
			_, err := db.ExecContext(ctx, `DROP TABLE accounts`)
			Expect(err).NotTo(HaveOccurred())

			created, err := persister.Create(ctx, values)
			Expect(created).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err).NotTo(MatchError(models.ErrAlreadyExists))
		},
	)
})
