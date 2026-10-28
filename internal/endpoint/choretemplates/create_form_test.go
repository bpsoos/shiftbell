package choretemplates_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("New template form", func() {
	DescribeTable(
		"renders an empty form without creating anything",
		func(ctx SpecContext, hx string, fullPage bool) {
			view := NewMockView(GinkgoT())
			view.EXPECT().
				CreateForm(viewmodels.CreateForm{ActionHref: "/chore-templates", CancelHref: "/chore-templates"}, fullPage).
				Return(templ.Raw("new template")).
				Once()
			handler := choretemplates.NewHandler(
				&choretemplates.HandlerDeps{
					Service: NewMockService(GinkgoT()),
					View:    view,
				},
			)
			app := echo.New()
			app.GET("/chore-templates/new", handler.CreateForm)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chore-templates/new",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hx)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("new template"))
		},
		Entry("full page", "", true),
		Entry("HTMX", "true", false),
	)

	DescribeTable("rejects non-HTML requests", func(ctx SpecContext, accept string) {
		handler := choretemplates.NewHandler(
			&choretemplates.HandlerDeps{Service: NewMockService(GinkgoT())},
		)
		app := echo.New()
		app.GET("/chore-templates/new", handler.CreateForm)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates/new",
			nil,
		)
		request.Header.Set("Accept", accept)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
	}, Entry("JSON", hypermedia.MediaType), Entry("unsupported", "image/png"))
})
