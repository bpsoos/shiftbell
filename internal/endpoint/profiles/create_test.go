package profiles_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/profiles"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	validation "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {
	DescribeTable(
		"accepts JSON and form input",
		func(ctx SpecContext, contentType, body string) {
			handler, service, _ := newHandler()
			service.On("SetupRequired", ctx).Return(false, nil).Once()
			service.On("Create", ctx, &models.CreateParams{Name: " Alex "}).
				Return(&models.Profile{Id: 7, Name: "Alex"}, nil).
				Once()
			e := echo.New()
			e.POST("/profiles", handler.Create)
			e.GET("/setup", handler.SetupStatus)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/profiles",
				strings.NewReader(body),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", contentType)
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusCreated))
			Expect(
				response.Body.Bytes(),
			).To(MatchJSON(`{"id":7,"name":"Alex","_links":[{"rel":"collection","href":"/profiles"}],"_actions":[{"rel":"delete","href":"/profiles/7"}]}`))
			request = httptest.NewRequestWithContext(ctx, http.MethodGet, "/setup", nil)
			request.Header.Set("Accept", hypermedia.MediaType)
			response = httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(ContainSubstring(`"required":false`))
		},
		Entry("JSON", "application/json", `{"name":" Alex "}`),
		Entry("vendor JSON", hypermedia.MediaType, `{"name":" Alex "}`),
		Entry("form", "application/x-www-form-urlencoded", "name=+Alex+"),
	)

	DescribeTable(
		"redirects HTML clients after creation",
		func(ctx SpecContext, hx string, status int, header string) {
			handler, service, _ := newHandler()
			service.On("Create", ctx, &models.CreateParams{Name: "Alex"}).
				Return(&models.Profile{Id: 7, Name: "Alex"}, nil).
				Once()
			e := echo.New()
			e.POST("/profiles", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/profiles",
				strings.NewReader("name=Alex"),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("HX-Request", hx)
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
			Expect(response.Header().Get(header)).To(Equal("/profiles"))
		},
		Entry("browser", "", http.StatusSeeOther, "Location"),
		Entry("HTMX", "true", http.StatusOK, "HX-Redirect"),
	)

	DescribeTable(
		"maps service errors to JSON status codes",
		func(ctx SpecContext, failure error, status int, message string) {
			handler, service, _ := newHandler()
			service.On("SetupRequired", ctx).Return(true, nil).Once()
			service.On("Create", ctx, &models.CreateParams{Name: "Alex"}).
				Return(nil, failure).
				Once()
			e := echo.New()
			e.POST("/profiles", handler.Create)
			e.GET("/setup", handler.SetupStatus)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/profiles",
				strings.NewReader(`{"name":"Alex"}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
			Expect(
				response.Body.String(),
			).To(ContainSubstring(`"error":"` + message + `"`))
			request = httptest.NewRequestWithContext(ctx, http.MethodGet, "/setup", nil)
			request.Header.Set("Accept", hypermedia.MediaType)
			response = httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(ContainSubstring(`"required":true`))
		},
		Entry(
			"name",
			validation.ErrInvalidName,
			http.StatusUnprocessableEntity,
			"invalid name",
		),
		Entry(
			"duplicate",
			models.ErrNameConflict,
			http.StatusConflict,
			models.ErrNameConflict.Error(),
		),
		Entry(
			"limit reached",
			models.ErrLimitReached,
			http.StatusConflict,
			models.ErrLimitReached.Error(),
		),
		Entry(
			"storage",
			errors.New("secret"),
			http.StatusInternalServerError,
			"The profile could not be created. Try again.",
		),
	)

	DescribeTable(
		"preserves submitted values and renders HTML feedback",
		func(ctx SpecContext, failure error, status int, fieldError, summary string) {
			handler, service, view := newHandler()
			service.On("Create", ctx, &models.CreateParams{Name: " Alex "}).
				Return(nil, failure).
				Once()
			view.On("Form", viewmodels.Form{
				ActionHref: "/profiles",
				CancelHref: "/profiles",
				Name: viewmodels.Field{
					Value: " Alex ",
					Error: fieldError,
				},
				SummaryError: summary,
			}, false).Return(templ.Raw("form error")).Once()
			e := echo.New()
			e.POST("/profiles", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/profiles",
				strings.NewReader("name=+Alex+"),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
			Expect(response.Body.String()).To(Equal("form error"))
		},
		Entry(
			"name",
			validation.ErrInvalidName,
			http.StatusUnprocessableEntity,
			"Enter a valid name.",
			"",
		),
		Entry(
			"empty name",
			errors.Join(validation.ErrInvalidName, validation.ErrRequired),
			http.StatusUnprocessableEntity,
			"Enter your name.",
			"",
		),
		Entry(
			"long name",
			errors.Join(validation.ErrInvalidName, validation.ErrTooLong),
			http.StatusUnprocessableEntity,
			"Use 200 characters or fewer.",
			"",
		),
		Entry(
			"spaces",
			errors.Join(validation.ErrInvalidName, models.ErrNameContainsSpaces),
			http.StatusUnprocessableEntity,
			"Spaces aren’t allowed.",
			"",
		),
		Entry(
			"unsupported character",
			errors.Join(validation.ErrInvalidName, validation.ErrDisallowedCharacter),
			http.StatusUnprocessableEntity,
			"Your name contains a character we can’t use.",
			"",
		),
		Entry(
			"invalid encoding",
			errors.Join(validation.ErrInvalidName, validation.ErrInvalidUTF8),
			http.StatusUnprocessableEntity,
			"Your name contains a character we can’t use.",
			"",
		),
		Entry(
			"duplicate",
			models.ErrNameConflict,
			http.StatusConflict,
			"A profile already uses this name.",
			"",
		),
		Entry(
			"limit reached",
			models.ErrLimitReached,
			http.StatusConflict,
			"",
			"Profile limit reached. Delete a profile before creating another.",
		),
		Entry(
			"storage",
			errors.New("secret"),
			http.StatusInternalServerError,
			"",
			"The profile could not be created. Try again.",
		),
	)

	DescribeTable(
		"rejects malformed and unsupported requests without mutations",
		func(ctx SpecContext, accept, contentType, body string, status int) {
			handler, _, _ := newHandler()
			e := echo.New()
			e.POST("/profiles", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/profiles",
				strings.NewReader(body),
			)
			request.Header.Set("Accept", accept)
			request.Header.Set("Content-Type", contentType)
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry(
			"unsupported response",
			"application/xml",
			"application/json",
			`{"name":"Alex"}`,
			http.StatusNotAcceptable,
		),
		Entry(
			"unsupported body",
			hypermedia.MediaType,
			"text/plain",
			"Alex",
			http.StatusUnsupportedMediaType,
		),
		Entry(
			"missing content type",
			hypermedia.MediaType,
			"",
			`{"name":"Alex"}`,
			http.StatusUnsupportedMediaType,
		),
		Entry(
			"malformed JSON",
			hypermedia.MediaType,
			"application/json",
			"{",
			http.StatusBadRequest,
		),
		Entry(
			"multiple values",
			hypermedia.MediaType,
			"application/json",
			`{"name":"Alex"}{}`,
			http.StatusBadRequest,
		),
		Entry(
			"wrong type",
			hypermedia.MediaType,
			"application/json",
			`{"name":12}`,
			http.StatusBadRequest,
		),
	)

	It("renders binding failures in the form", func(ctx SpecContext) {
		handler, _, view := newHandler()
		view.On("Form", viewmodels.Form{
			ActionHref: "/profiles", CancelHref: "/profiles",
			SummaryError: "The form could not be read. Try again.",
		}, true).Return(templ.Raw("invalid body")).Once()
		e := echo.New()
		e.POST("/profiles", handler.Create)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPost,
			"/profiles",
			strings.NewReader("{"),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		e.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(response.Body.String()).To(Equal("invalid body"))
	})

	It(
		"does not mutate state for HTML requests without a renderer",
		func(ctx SpecContext) {
			handler := profiles.NewHandler(
				&profiles.HandlerDeps{Service: &MockService{}},
				&profiles.HandlerConfig{UIEnabled: true},
			)
			e := echo.New()
			e.POST("/profiles", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/profiles",
				strings.NewReader(`{"name":"Alex"}`),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		},
	)
})
