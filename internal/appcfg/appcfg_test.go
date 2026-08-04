package appcfg_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bpsoos/shiftbell/internal/appcfg"
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
		GinkgoT().Setenv("APP_TIMEZONE", "")
	})

	It("loads the database filepath", func() {
		cfg, err := appcfg.Load()

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.DatabaseFilepath).To(Equal(databaseFilepath))
		Expect(cfg.LogHandler).To(Equal(logging.HandlerJSON))
	})

	Context("without an app timezone", func() {
		BeforeEach(func() {
			Expect(os.Unsetenv("APP_TIMEZONE")).To(Succeed())
		})

		It("defaults to UTC", func() {
			cfg, err := appcfg.Load()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.AppTimezone).To(Equal(time.UTC))
		})
	})

	Context("with an empty app timezone", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("APP_TIMEZONE", "")
		})

		It("defaults to UTC", func() {
			cfg, err := appcfg.Load()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.AppTimezone).To(Equal(time.UTC))
		})
	})

	Context("with an app timezone", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("APP_TIMEZONE", "Europe/Budapest")
		})

		It("loads the timezone", func() {
			cfg, err := appcfg.Load()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.AppTimezone).To(Equal(mustLoadLocation("Europe/Budapest")))
		})
	})

	Context("with an invalid app timezone", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("APP_TIMEZONE", "invalid")
		})

		It("returns an error", func() {
			cfg, err := appcfg.Load()

			Expect(cfg).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Context("with text log format", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("LOG_FORMAT", "text")
		})

		It("loads the text handler", func() {
			cfg, err := appcfg.Load()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.LogHandler).To(Equal(logging.HandlerConsole))
		})
	})

	Context("with JSON log format", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("LOG_FORMAT", "json")
		})

		It("loads the JSON handler", func() {
			cfg, err := appcfg.Load()

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.LogHandler).To(Equal(logging.HandlerJSON))
		})
	})

	Context("with an invalid log format", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("LOG_FORMAT", "xml")
		})

		It("returns an error", func() {
			cfg, err := appcfg.Load()

			Expect(cfg).To(BeNil())
			Expect(err).To(MatchError(`invalid LOG_FORMAT "xml": expected text or json`))
		})
	})
})

func mustLoadLocation(name string) *time.Location {
	GinkgoHelper()

	location, err := time.LoadLocation(name)
	Expect(err).NotTo(HaveOccurred())

	return location
}
