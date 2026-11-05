package profiles_test

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RequireSetup", func() {
	DescribeTable(
		"redirects HTML navigation when setup is required",
		func(ctx SpecContext, hx string, status int, header string) {
			handler, service, _ := newHandler()
			service.On("SetupRequired", ctx).Return(true, nil).Once()
			next := handler.RequireSetup(func(*echo.Context) error {
				Fail("the protected page must not run before setup")
				return nil
			})
			response := perform(
				ctx,
				next,
				http.MethodGet,
				"/chores",
				"text/html",
				"",
				"",
				hx,
			)
			Expect(response.Code).To(Equal(status))
			Expect(response.Header().Get(header)).To(Equal("/setup"))
		},
		Entry("browser", "", http.StatusSeeOther, "Location"),
		Entry("HTMX", "true", http.StatusOK, "HX-Redirect"),
	)

	It(
		"allows HTML navigation once any browser has created a profile",
		func(ctx SpecContext) {
			handler, service, _ := newHandler()
			service.On("SetupRequired", ctx).Return(false, nil).Once()
			next := handler.RequireSetup(func(c *echo.Context) error {
				return c.String(http.StatusOK, "chores")
			})
			response := perform(
				ctx,
				next,
				http.MethodGet,
				"/chores",
				"text/html",
				"",
				"",
				"",
			)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("chores"))
		},
	)

	It("passes JSON API requests through independently of setup", func(ctx SpecContext) {
		handler, _, _ := newHandler()
		next := handler.RequireSetup(func(c *echo.Context) error {
			return hypermedia.JSON(c, http.StatusOK, struct{}{})
		})
		response := perform(
			ctx,
			next,
			http.MethodGet,
			"/chores",
			hypermedia.MediaType,
			"",
			"",
			"",
		)
		Expect(response.Code).To(Equal(http.StatusOK))
	})

	It("renders lookup failures instead of redirecting", func(ctx SpecContext) {
		handler, service, view := newHandler()
		service.On("SetupRequired", ctx).Return(false, errors.New("secret")).Once()
		view.On("Error", viewmodels.Error{Message: "Something went wrong. Try again.", CollectionHref: "/profiles"}, true).
			Return(templ.Raw("error")).
			Once()
		next := handler.RequireSetup(func(*echo.Context) error {
			Fail("the protected page must not run after a lookup failure")
			return nil
		})
		response := perform(ctx, next, http.MethodGet, "/chores", "text/html", "", "", "")
		Expect(response.Code).To(Equal(http.StatusInternalServerError))
		Expect(response.Body.String()).To(Equal("error"))
	})
})
