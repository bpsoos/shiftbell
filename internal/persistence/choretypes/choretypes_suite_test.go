package choretypes_test

import (
	"testing"

	"github.com/bpsoos/shiftbell/internal/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChoretypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Choretypes Suite")
}

var _ = BeforeSuite(func() {
	logging.Configure(logging.Config{
		Handler: logging.HandlerDiscard,
	})
})
