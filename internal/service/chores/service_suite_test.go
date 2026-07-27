package chores

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChoresService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chores Service Suite")
}
