package passwordhash_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPasswordHash(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Password Hash Suite")
}
