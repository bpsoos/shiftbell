package profiles_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/profiles"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RequireSetup", func() {
	It(
		"allows browser navigation before setup when the UI is disabled",
		func(ctx SpecContext) {
			handler := profiles.NewHandler(
				&profiles.HandlerDeps{Service: &MockService{}},
				&profiles.HandlerConfig{SetupRequired: true},
			)
			e := echo.New()
			e.GET("/chores", handler.RequireSetup(func(c *echo.Context) error {
				return c.String(http.StatusOK, "chores")
			}))
			request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("chores"))
		},
	)
	DescribeTable(
		"redirects HTML navigation when setup is required",
		func(ctx SpecContext, hx string, status int, header string) {
			handler, _, _ := newHandler()
			next := handler.RequireSetup(func(*echo.Context) error {
				Fail("the protected page must not run before setup")
				return nil
			})
			e := echo.New()
			e.GET("/chores", next)
			request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hx)
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
			Expect(response.Header().Get(header)).To(Equal("/setup"))
		},
		Entry("browser", "", http.StatusSeeOther, "Location"),
		Entry("HTMX", "true", http.StatusOK, "HX-Redirect"),
	)

	It(
		"allows HTML navigation when startup found an existing profile",
		func(ctx SpecContext) {
			handler, _, _ := newHandlerWithSetupRequired(false)
			next := handler.RequireSetup(func(c *echo.Context) error {
				return c.String(http.StatusOK, "chores")
			})
			e := echo.New()
			e.GET("/chores", next)
			request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("chores"))
		},
	)

	It("passes JSON API requests through independently of setup", func(ctx SpecContext) {
		handler, _, _ := newHandler()
		next := handler.RequireSetup(func(c *echo.Context) error {
			return hypermedia.JSON(c, http.StatusOK, struct{}{})
		})
		e := echo.New()
		e.GET("/chores", next)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		e.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
	})
})
