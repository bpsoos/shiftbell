package profiles_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	validation "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupStatus", func() {
	It("reflects the last profile being removed between requests", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.On("SetupRequired", ctx).Return(false, nil).Once()
		service.On("SetupRequired", ctx).Return(true, nil).Once()
		app := echo.New()
		app.GET("/setup", handler.SetupStatus)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/setup", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(ContainSubstring(`"required":false`))
		response = httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(ContainSubstring(`"required":true`))
		Expect(
			response.Body.String(),
		).To(ContainSubstring(`"rel":"create","href":"/setup"`))
	})

	It("reports setup lookup failures", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.On("SetupRequired", ctx).
			Return(false, errors.New("storage failed")).
			Once()
		app := echo.New()
		app.GET("/setup", handler.SetupStatus)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/setup", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusInternalServerError))
	})

	DescribeTable(
		"reports current profile state and available actions",
		func(ctx SpecContext, required bool, expected string) {
			handler, service, _ := newHandler()
			service.On("SetupRequired", ctx).Return(required, nil).Once()
			e := echo.New()
			e.GET("/setup", handler.SetupStatus)
			request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/setup", nil)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.Bytes()).To(MatchJSON(expected))
		},
		Entry(
			"required",
			true,
			`{"required":true,"_links":[{"rel":"self","href":"/setup"},{"rel":"profiles","href":"/profiles"},{"rel":"chores","href":"/chores"}],"_actions":[{"rel":"create","href":"/setup"}]}`,
		),
		Entry(
			"complete",
			false,
			`{"required":false,"_links":[{"rel":"self","href":"/setup"},{"rel":"profiles","href":"/profiles"},{"rel":"chores","href":"/chores"}],"_actions":[]}`,
		),
	)

	It("renders the setup form", func(ctx SpecContext) {
		handler, service, view := newHandler()
		service.On("SetupRequired", ctx).Return(true, nil).Once()
		view.On("Form", viewmodels.Form{ActionHref: "/setup", Setup: true}, true).
			Return(templ.Raw("setup form")).Once()
		e := echo.New()
		e.GET("/setup", handler.SetupStatus)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/setup", nil)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("setup form"))
	})

	It(
		"redirects to chores when a profile exists",
		func(ctx SpecContext) {
			handler, service, _ := newHandler()
			service.On("SetupRequired", ctx).Return(false, nil).Once()
			e := echo.New()
			e.GET("/setup", handler.SetupStatus)
			request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/setup", nil)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusSeeOther))
			Expect(response.Header().Get("Location")).To(Equal("/chores"))
		},
	)
})

var _ = Describe("Setup", func() {
	It(
		"creates the first profile and redirects subsequent visits to chores",
		func(ctx SpecContext) {
			handler, service, _ := newHandler()
			service.On("SetupRequired", ctx).Return(false, nil).Once()
			service.On("Setup", ctx, &models.CreateParams{Name: "Alex"}).
				Return(&models.Profile{Id: 1, Name: "Alex"}, nil).Once()
			e := echo.New()
			e.POST("/setup", handler.Setup)
			e.GET("/setup", handler.SetupStatus)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/setup",
				strings.NewReader("name=Alex"),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("HX-Redirect")).To(Equal("/chores"))

			request = httptest.NewRequestWithContext(ctx, http.MethodGet, "/setup", nil)
			request.Header.Set("Accept", "text/html")
			response = httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusSeeOther))
			Expect(response.Header().Get("Location")).To(Equal("/chores"))
		},
	)

	It("keeps setup required after a storage failure", func(ctx SpecContext) {
		handler, service, _ := newHandler()
		service.On("SetupRequired", ctx).Return(true, nil).Once()
		service.On("Setup", ctx, &models.CreateParams{Name: "Alex"}).
			Return(nil, errors.New("storage failed")).Once()
		e := echo.New()
		e.POST("/setup", handler.Setup)
		e.GET("/setup", handler.SetupStatus)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPost,
			"/setup",
			strings.NewReader(`{"name":"Alex"}`),
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		e.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusInternalServerError))

		request = httptest.NewRequestWithContext(ctx, http.MethodGet, "/setup", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response = httptest.NewRecorder()
		e.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(ContainSubstring(`"required":true`))
	})

	It(
		"offers continuation when setup is already complete",
		func(ctx SpecContext) {
			handler, service, view := newHandler()
			service.On("SetupRequired", ctx).Return(false, nil).Once()
			service.On("Setup", ctx, &models.CreateParams{Name: "Alex"}).
				Return(nil, models.ErrAlreadyInitialized).Once()
			view.On("Form", viewmodels.Form{
				ActionHref: "/setup", CancelHref: "/chores", Setup: true,
				Name:         viewmodels.Field{Value: "Alex"},
				SummaryError: "Setup is already complete. Continue to chores.",
			}, true).Return(templ.Raw("continue")).Once()
			e := echo.New()
			e.POST("/setup", handler.Setup)
			e.GET("/setup", handler.SetupStatus)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/setup",
				strings.NewReader("name=Alex"),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusConflict))
			Expect(response.Body.String()).To(Equal("continue"))

			request = httptest.NewRequestWithContext(ctx, http.MethodGet, "/setup", nil)
			request.Header.Set("Accept", "text/html")
			response = httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusSeeOther))
			Expect(response.Header().Get("Location")).To(Equal("/chores"))
		},
	)

	It("preserves the setup form after invalid input", func(ctx SpecContext) {
		handler, service, view := newHandler()
		service.On("SetupRequired", ctx).Return(true, nil).Once()
		service.On("Setup", ctx, &models.CreateParams{Name: "Alex Smith"}).
			Return(nil, errors.Join(validation.ErrInvalidName, models.ErrNameContainsSpaces)).
			Once()
		view.On("Form", viewmodels.Form{
			ActionHref: "/setup", Setup: true,
			Name: viewmodels.Field{Value: "Alex Smith", Error: "Spaces aren’t allowed."},
		}, true).Return(templ.Raw("invalid name")).Once()
		e := echo.New()
		e.POST("/setup", handler.Setup)
		e.GET("/setup", handler.SetupStatus)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPost,
			"/setup",
			strings.NewReader("name=Alex+Smith"),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		e.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(Equal("invalid name"))

		request = httptest.NewRequestWithContext(ctx, http.MethodGet, "/setup", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response = httptest.NewRecorder()
		e.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(ContainSubstring(`"required":true`))
	})
})
