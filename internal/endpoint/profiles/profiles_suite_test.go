package profiles_test

import (
	"testing"

	"github.com/bpsoos/shiftbell/internal/endpoint/profiles"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProfiles(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Profile Endpoint Suite")
}

func newHandler() (*profiles.Handler, *MockService, *MockView) {
	GinkgoHelper()
	service := &MockService{}
	view := &MockView{}
	DeferCleanup(func() {
		service.AssertExpectations(GinkgoT())
		view.AssertExpectations(GinkgoT())
	})
	return profiles.NewHandler(
		&profiles.HandlerDeps{Service: service, View: view},
		&profiles.HandlerConfig{UIEnabled: true},
	), service, view
}
