package profiles_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/profiles"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Profile selection", func() {
	It(
		"renders the picker with the current choice and return location",
		func(ctx SpecContext) {
			handler, service, view := newHandlerWithSetupRequired(false)
			available := []models.Profile{{Id: 7, Name: "Alex"}, {Id: 8, Name: "Sam"}}
			service.EXPECT().List(ctx).Return(available, nil).Once()
			view.EXPECT().
				Selection(viewmodels.Selection{Profiles: available, SelectedProfileId: 7, ReturnTo: "/chores?status=completed"}, true).
				Return(templ.Raw("picker sentinel")).
				Once()
			app := echo.New()
			app.GET("/profiles/selection", handler.Selection)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/profiles/selection?return_to=%2Fchores%3Fstatus%3Dcompleted",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.AddCookie(&http.Cookie{Name: "shiftbell_profile", Value: "7"})
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("picker sentinel"))
		},
	)

	It("skips the picker for a single profile", func(ctx SpecContext) {
		handler, service, _ := newHandlerWithSetupRequired(false)
		service.EXPECT().
			List(ctx).
			Return([]models.Profile{{Id: 7, Name: "Alex"}}, nil).
			Once()
		app := echo.New()
		app.GET("/profiles/selection", handler.Selection)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/profiles/selection?return_to=%2Fschedules",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusSeeOther))
		Expect(response.Header().Get("Location")).To(Equal("/schedules"))
		Expect(response.Result().Cookies()).To(BeEmpty())
	})

	DescribeTable(
		"remembers the new choice and returns to the page",
		func(ctx SpecContext, htmx string, status int, header string) {
			handler, service, _ := newHandlerWithSetupRequired(false)
			service.EXPECT().
				List(ctx).
				Return([]models.Profile{{Id: 7, Name: "Alex"}, {Id: 8, Name: "Sam"}}, nil).
				Once()
			app := echo.New()
			app.POST("/profiles/selection", handler.Select)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"https://example.com/profiles/selection",
				strings.NewReader("profile_id=8&return_to=%2Fchores%3Fsearch%3DKitchen"),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("HX-Request", htmx)
			request.Header.Set("Referer", "https://example.com/profiles/selection")
			request.AddCookie(&http.Cookie{Name: "shiftbell_profile", Value: "7"})
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
			Expect(response.Header().Get(header)).To(Equal("/chores?search=Kitchen"))
			cookies := response.Result().Cookies()
			Expect(cookies).To(HaveLen(1))
			Expect(cookies[0].Name).To(Equal("shiftbell_profile"))
			Expect(cookies[0].Value).To(Equal("8"))
			Expect(cookies[0].Path).To(Equal("/"))
			Expect(cookies[0].MaxAge).To(BeNumerically(">", 0))
			Expect(cookies[0].HttpOnly).To(BeTrue())
			Expect(cookies[0].Secure).To(BeTrue())
			Expect(cookies[0].SameSite).To(Equal(http.SameSiteLaxMode))
		},
		Entry("browser", "", 303, "Location"),
		Entry("HTMX forces a full navigation refresh", "true", 200, "HX-Redirect"),
	)

	DescribeTable(
		"returns dropdown selections to a local referring page",
		func(ctx SpecContext, referer, destination string) {
			handler, service, _ := newHandlerWithSetupRequired(false)
			service.EXPECT().
				List(ctx).
				Return([]models.Profile{{Id: 7}, {Id: 8}}, nil).
				Once()
			app := echo.New()
			app.POST("/profiles/selection", handler.Select)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"https://example.com/profiles/selection",
				strings.NewReader("profile_id=8"),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Referer", referer)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusSeeOther))
			Expect(response.Header().Get("Location")).To(Equal(destination))
			Expect(response.Result().Cookies()).To(HaveLen(1))
			Expect(response.Result().Cookies()[0].Value).To(Equal("8"))
		},
		Entry(
			"preserves path and query",
			"https://example.com/chores?status=completed&search=Kitchen",
			"/chores?status=completed&search=Kitchen",
		),
		Entry("external referrer", "https://other.example/schedules", "/chores"),
		Entry("missing referrer", "", "/chores"),
		Entry("picker referrer", "https://example.com/profiles/selection", "/chores"),
	)

	It("rejects a profile that is not available", func(ctx SpecContext) {
		handler, service, view := newHandlerWithSetupRequired(false)
		available := []models.Profile{{Id: 7, Name: "Alex"}, {Id: 8, Name: "Sam"}}
		service.EXPECT().List(ctx).Return(available, nil).Once()
		view.EXPECT().
			Selection(viewmodels.Selection{Profiles: available, ReturnTo: "/chores", Error: "Choose an available profile."}, true).
			Return(templ.Raw("selection error")).
			Once()
		app := echo.New()
		app.POST("/profiles/selection", handler.Select)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPost,
			"/profiles/selection",
			strings.NewReader("profile_id=99"),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(Equal("selection error"))
		Expect(response.Result().Cookies()).To(BeEmpty())
	})

	DescribeTable(
		"restricts return locations to local pages",
		func(ctx SpecContext, returnTo string) {
			handler, service, _ := newHandlerWithSetupRequired(false)
			service.EXPECT().
				List(ctx).
				Return([]models.Profile{{Id: 7}, {Id: 8}}, nil).
				Once()
			app := echo.New()
			app.POST("/profiles/selection", handler.Select)
			body := url.Values{"profile_id": {"7"}, "return_to": {returnTo}}.Encode()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/profiles/selection",
				strings.NewReader(body),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusSeeOther))
			Expect(response.Header().Get("Location")).To(Equal("/chores"))
		},
		Entry("external URL", "https://example.org/"),
		Entry("protocol-relative URL", "//example.org/"),
		Entry("browser backslash URL", `/\example.org/`),
		Entry("picker loop", "/profiles/selection"),
	)

	DescribeTable(
		"hides selection routes when the flag is disabled",
		func(ctx SpecContext, method string) {
			handler := profiles.NewHandler(
				&profiles.HandlerDeps{Service: NewMockService(GinkgoT())},
				&profiles.HandlerConfig{},
			)
			app := echo.New()
			app.GET("/profiles/selection", handler.Selection)
			app.POST("/profiles/selection", handler.Select)
			request := httptest.NewRequestWithContext(
				ctx,
				method,
				"/profiles/selection",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusNotFound))
		},
		Entry("picker", http.MethodGet),
		Entry("selection", http.MethodPost),
	)
})
