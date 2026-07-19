package appcfg

import (
	"path/filepath"
	"testing"

	"github.com/bpsoos/shiftbell/internal/logging"
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
		GinkgoT().Setenv("LOG_FORMAT", "")
	})

	It("loads the database filepath", func() {
		cfg, err := Load()

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.DatabaseFilepath).To(Equal(databaseFilepath))
		Expect(cfg.LogHandler).To(Equal(logging.HandlerJSON))
	})

	Context("with text log format", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("LOG_FORMAT", "text")
		})

		It("loads the text handler", func() {
			cfg, err := Load()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.LogHandler).To(Equal(logging.HandlerConsole))
		})
	})

	Context("with JSON log format", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("LOG_FORMAT", "json")
		})

		It("loads the JSON handler", func() {
			cfg, err := Load()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.LogHandler).To(Equal(logging.HandlerJSON))
		})
	})

	Context("with an invalid log format", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("LOG_FORMAT", "xml")
		})

		It("returns an error", func() {
			cfg, err := Load()

			Expect(cfg).To(BeNil())
			Expect(err).To(MatchError(`invalid LOG_FORMAT "xml": expected text or json`))
		})
	})
})
