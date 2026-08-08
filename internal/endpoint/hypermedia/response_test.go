package hypermedia_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Representation responses", func() {
	It("encodes JSON with the negotiated response headers", func() {
		e := echo.New()
		e.GET("/", func(ctx *echo.Context) error {
			return hypermedia.JSON(
				ctx,
				http.StatusCreated,
				map[string]string{"href": "/chores"},
			)
		})
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusCreated))
		Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
		Expect(response.Header().Get("Vary")).To(Equal("Accept"))
		Expect(response.Body.Bytes()).To(MatchJSON(`{"href":"/chores"}`))
	})

	It("renders an HTML component with the negotiated response headers", func() {
		e := echo.New()
		e.GET("/", func(ctx *echo.Context) error {
			return hypermedia.HTML(ctx, http.StatusOK, templ.Raw("<main>Chores</main>"))
		})
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(
			response.Header().Get("Content-Type"),
		).To(Equal(echo.MIMETextHTMLCharsetUTF8))
		Expect(response.Header().Get("Vary")).To(Equal("Accept, HX-Request"))
		Expect(response.Body.String()).To(Equal("<main>Chores</main>"))
	})

	It("returns a negotiated not acceptable response", func() {
		e := echo.New()
		e.GET("/", hypermedia.NotAcceptable)
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Header().Get("Vary")).To(Equal("Accept"))
		Expect(response.Body.String()).To(BeEmpty())
	})

	It("redirects an HTML request with the negotiated response headers", func() {
		e := echo.New()
		e.GET("/", func(ctx *echo.Context) error {
			return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, "/chores")
		})
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusSeeOther))
		Expect(response.Header().Get("Location")).To(Equal("/chores"))
		Expect(response.Header().Get("Vary")).To(Equal("Accept, HX-Request"))
	})

	It("turns an HTMX redirect into a client-side redirect instruction", func() {
		e := echo.New()
		e.GET("/", func(ctx *echo.Context) error {
			return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, "/chores")
		})
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("HX-Redirect")).To(Equal("/chores"))
		Expect(response.Header().Get("Vary")).To(Equal("Accept, HX-Request"))
	})

	It("returns an empty negotiated response", func() {
		e := echo.New()
		e.DELETE("/", func(ctx *echo.Context) error {
			return hypermedia.NoContent(ctx, http.StatusNoContent)
		})
		request := httptest.NewRequest(http.MethodDelete, "/", nil)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNoContent))
		Expect(response.Header().Get("Vary")).To(Equal("Accept"))
		Expect(response.Body.String()).To(BeEmpty())
	})
})
