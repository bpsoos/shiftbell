package chores_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChoresEndpoint(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chores Endpoint Suite")
}
