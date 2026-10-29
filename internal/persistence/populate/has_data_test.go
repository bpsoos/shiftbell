package populate_test

import (
	"context"
	"testing"

	"github.com/bpsoos/shiftbell/internal/persistence/populate"
	"github.com/bpsoos/shiftbell/internal/testsupport/sqlitetest"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPopulatePersistence(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Populate Persistence Suite")
}

var _ = Describe("HasData", func() {
	var db *sqlx.DB
	var persister *populate.Persister

	BeforeEach(func() {
		db = sqlitetest.NewMigratedDB()
		persister = populate.NewPersister(&populate.PersisterDeps{Db: db})
	})

	It("ignores migration metadata in an empty database", func() {
		populated, err := persister.HasData(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(populated).To(BeFalse())
	})

	DescribeTable(
		"finds existing application data",
		func(setup string) {
			_, err := db.Exec(setup)
			Expect(err).NotTo(HaveOccurred())
			populated, err := persister.HasData(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(populated).To(BeTrue())
		},
		Entry(
			"an active chore",
			"INSERT INTO chores (name, is_complete) VALUES ('Chore', 0)",
		),
		Entry(
			"a completed chore",
			"INSERT INTO chores (name, is_complete) VALUES ('Chore', 1)",
		),
		Entry("a template", "INSERT INTO chore_templates (name) VALUES ('Template')"),
		Entry(
			"an inactive template",
			"INSERT INTO chore_templates (name, deactivated_at) VALUES ('Template', '2026-01-01')",
		),
		Entry(
			"a schedule",
			`INSERT INTO chore_templates (id, name) VALUES (1, 'Template');
			INSERT INTO schedules (name, chore_template_id, interval_days) VALUES ('Schedule', 1, 7)`,
		),
	)

	It("propagates cancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := persister.HasData(ctx)
		Expect(err).To(MatchError(context.Canceled))
	})
})
