package appcfg

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAppCfg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AppCfg Suite")
}

var _ = Describe("Load", func() {
	var databaseFilepath string

	BeforeEach(func() {
		databaseFilepath = filepath.Join(GinkgoT().TempDir(), "shiftbell.db")
		GinkgoT().Setenv("DATABASE_FILEPATH", databaseFilepath)
	})

	It("loads the database filepath", func() {
		cfg, err := Load(context.Background())

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.DatabaseFilepath).To(Equal(databaseFilepath))
	})
})
