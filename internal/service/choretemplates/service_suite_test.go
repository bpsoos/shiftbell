package choretemplates_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChoreTemplatesService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chore Templates Service Suite")
}
