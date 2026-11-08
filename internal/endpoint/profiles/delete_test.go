package profiles_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete", func() {
	DescribeTable(
		"redirects HTML clients after deletion",
		func(ctx SpecContext, hx string, status int, header string, failure error) {
			handler, service, _ := newHandler()
			service.On("Delete", ctx, 7).Return(failure).Once()
			app := echo.New()
			app.DELETE("/profiles/:id", handler.Delete)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodDelete,
				"/profiles/7",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hx)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
			Expect(response.Header().Get(header)).To(Equal("/profiles"))
		},
		Entry("browser", "", http.StatusSeeOther, "Location", nil),
		Entry("HTMX", "true", http.StatusOK, "HX-Redirect", nil),
		Entry(
			"already deleted",
			"true",
			http.StatusOK,
			"HX-Redirect",
			models.ErrNotFound,
		),
	)

	It("renders deletion failures for HTMX", func(ctx SpecContext) {
		handler, service, view := newHandler()
		service.On("Delete", ctx, 7).Return(errors.New("storage failed")).Once()
		view.On("Error", viewmodels.Error{
			Message: "Something went wrong. Try again.", CollectionHref: "/profiles",
		}, false).Return(templ.Raw("deletion failed")).Once()
		app := echo.New()
		app.DELETE("/profiles/:id", handler.Delete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodDelete,
			"/profiles/7",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusInternalServerError))
		Expect(response.Body.String()).To(Equal("deletion failed"))
		Expect(response.Header().Get("HX-Redirect")).To(BeEmpty())
	})

	It("renders invalid profile errors for HTMX", func(ctx SpecContext) {
		handler, _, view := newHandler()
		view.On("Error", viewmodels.Error{
			Message:        "Invalid profile. Return to profiles and try again.",
			CollectionHref: "/profiles",
		}, false).Return(templ.Raw("invalid profile")).Once()
		app := echo.New()
		app.DELETE("/profiles/:id", handler.Delete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodDelete,
			"/profiles/invalid",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(Equal("invalid profile"))
	})

	DescribeTable(
		"returns no content after deletion",
		func(ctx SpecContext, failure error) {
			handler, service, _ := newHandler()
			service.On("Delete", ctx, 7).Return(failure).Once()
			app := echo.New()
			app.DELETE("/profiles/:id", handler.Delete)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodDelete,
				"/profiles/7",
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusNoContent))
			Expect(response.Body.String()).To(BeEmpty())
		},
		Entry("existing profile", nil),
		Entry("already deleted profile", fmt.Errorf("delete: %w", models.ErrNotFound)),
	)

	DescribeTable("rejects invalid profile IDs", func(ctx SpecContext, id string) {
		handler, _, _ := newHandler()
		app := echo.New()
		app.DELETE("/profiles/:id", handler.Delete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodDelete,
			"/profiles/"+id,
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
	},
		Entry("non-numeric", "invalid"),
		Entry("zero", "0"),
		Entry("negative", "-1"),
		Entry("overflow", "99999999999999999999"),
	)

	DescribeTable(
		"rejects unsupported representations",
		func(ctx SpecContext, accept string) {
			handler, _, _ := newHandler()
			app := echo.New()
			app.DELETE("/profiles/:id", handler.Delete)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodDelete,
				"/profiles/7",
				nil,
			)
			request.Header.Set("Accept", accept)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		},
		Entry("plain text", "text/plain"),
		Entry("XML", "application/xml"),
	)

	It("reports deletion failures", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.On("Delete", ctx, 7).Return(errors.New("storage failed")).Once()
		app := echo.New()
		app.DELETE("/profiles/:id", handler.Delete)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodDelete,
			"/profiles/7",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusInternalServerError))
		Expect(
			response.Body.String(),
		).To(ContainSubstring(`"error":"something went wrong"`))
	})
})
