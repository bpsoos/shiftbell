package chores_test

import (
	"log/slog"
	"testing"

	"github.com/bpsoos/shiftbell/internal/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChores(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chores Persistence Suite")
}

var _ = BeforeSuite(func() {
	logging.Configure(logging.Config{
		Level:   slog.LevelInfo,
		Handler: logging.HandlerDiscard,
	})
})
