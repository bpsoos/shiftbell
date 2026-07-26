package normalization_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNormalization(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Normalization Suite")
}
