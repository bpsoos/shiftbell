package home_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/bpsoos/shiftbell/internal/endpoint/home"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get home", func() {
	It("advertises the chore and chore-template collections", func(ctx SpecContext) {
		handler := home.NewHandler()
		e := echo.New()
		e.GET("/", handler.Get)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
		Expect(response.Header().Get("Vary")).To(Equal("Accept"))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"_links": [
				{"rel": "self", "href": "/"},
				{"rel": "chores", "href": "/chores"},
				{"rel": "chore_templates", "href": "/chore-templates"}
			]
		}`))
	})

	It("redirects browser navigation to the chore collection", func(ctx SpecContext) {
		handler := home.NewHandler()
		e := echo.New()
		e.GET("/", handler.Get)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusSeeOther))
		Expect(response.Header().Get("Location")).To(Equal("/chores"))
		Expect(response.Header().Get("Vary")).To(Equal("Accept, HX-Request"))
	})

	It("rejects unsupported representations", func(ctx SpecContext) {
		handler := home.NewHandler()
		e := echo.New()
		e.GET("/", handler.Get)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
		request.Header.Set("Accept", "application/json")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Header().Get("Vary")).To(Equal("Accept"))
		Expect(response.Body.String()).To(BeEmpty())
	})
})
