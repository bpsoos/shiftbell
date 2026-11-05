package profiles_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bpsoos/shiftbell/internal/endpoint/profiles"
	"github.com/labstack/echo/v5"
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
	), service, view
}

func perform(
	ctx context.Context,
	handler echo.HandlerFunc,
	method, path, accept, contentType, body, hx string,
) *httptest.ResponseRecorder {
	GinkgoHelper()
	app := echo.New()
	app.Add(method, path, handler)
	request := httptest.NewRequestWithContext(ctx, method, path, strings.NewReader(body))
	request.Header.Set("Accept", accept)
	request.Header.Set("Content-Type", contentType)
	request.Header.Set("HX-Request", hx)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	return response
}
