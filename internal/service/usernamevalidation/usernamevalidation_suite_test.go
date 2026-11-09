package usernamevalidation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUsernameValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Username Validation Suite")
}
