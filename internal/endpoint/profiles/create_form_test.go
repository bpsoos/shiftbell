package profiles_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/profiles"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateForm", func() {
	DescribeTable(
		"renders the creation form",
		func(ctx SpecContext, hx string, fullPage, uiEnabled bool) {
			view := &MockView{}
			DeferCleanup(func() { view.AssertExpectations(GinkgoT()) })
			handler := profiles.NewHandler(
				&profiles.HandlerDeps{Service: &MockService{}, View: view},
				&profiles.HandlerConfig{SetupRequired: true, UIEnabled: uiEnabled},
			)
			view.On("Form", viewmodels.Form{ActionHref: "/profiles", CancelHref: "/profiles"}, fullPage).
				Return(templ.Raw("new profile")).
				Once()
			e := echo.New()
			e.GET("/profiles/new", handler.CreateForm)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/profiles/new",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hx)
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("new profile"))
		},
		Entry("page", "", true, true),
		Entry("fragment", "true", false, true),
		Entry("page with navigation disabled", "", true, false),
		Entry("fragment with navigation disabled", "true", false, false),
	)

	It("rejects JSON for the HTML-only form route", func(ctx SpecContext) {
		handler, _, _ := newHandler()
		e := echo.New()
		e.GET("/profiles/new", handler.CreateForm)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/profiles/new",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		e.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
	})
})
