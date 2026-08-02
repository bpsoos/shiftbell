package home_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHomeEndpoint(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Home Endpoint Suite")
}
