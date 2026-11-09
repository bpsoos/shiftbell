package passwordvalidation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPasswordValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Password Validation Suite")
}
