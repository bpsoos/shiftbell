package profiles_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete", func() {
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
		Entry("HTML", "text/html"),
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
