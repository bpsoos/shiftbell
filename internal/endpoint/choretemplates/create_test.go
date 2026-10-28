package choretemplates_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	apimodels "github.com/bpsoos/shiftbell/internal/models/api/choretemplates"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create chore template", func() {
	DescribeTable(
		"renders the created template as HTML",
		func(ctx SpecContext, hxRequest string, fullPage bool, contentType, body string) {
			service := NewMockService(GinkgoT())
			service.EXPECT().
				Create(ctx, &models.CreateChoreTemplateParams{Name: "Kitchen", Description: "Reusable steps."}).
				Return(&models.ChoreTemplate{Id: 7, Name: "Kitchen", Description: "Reusable steps."}, nil).
				Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().Detail(viewmodels.Detail{
				ChoreTemplate: apimodels.Representation{Response: apimodels.Response{
					Id:          7,
					Name:        "Kitchen",
					Description: "Reusable steps.",
					Links: api.Relations{
						{Rel: "self", Href: "/chore-templates/7"},
						{Rel: "collection", Href: "/chore-templates"},
					},
				}, Actions: api.Relations{{Rel: "edit", Href: "/chore-templates/7"}, {Rel: "deactivate", Href: "/chore-templates/7/deactivation"}}},
				CollectionHref: "/chore-templates",
				EditHref:       "/chore-templates/7/edit",
				DeactivateHref: "/chore-templates/7/deactivation",
				Notice:         "Chore template created.",
			}, fullPage).Return(templ.Raw("created template")).Once()
			handler := choretemplates.NewHandler(
				&choretemplates.HandlerDeps{Service: service, View: view},
			)
			app := echo.New()
			app.POST("/chore-templates", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chore-templates",
				strings.NewReader(body),
			)
			request.Header.Set("Content-Type", contentType)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hxRequest)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusCreated))
			Expect(
				response.Header().Get("Content-Type"),
			).To(Equal(echo.MIMETextHTMLCharsetUTF8))
			Expect(response.Header().Get("Location")).To(Equal("/chore-templates/7"))
			Expect(
				response.Header().Get("HX-Replace-Url"),
			).To(Equal("/chore-templates/7"))
			Expect(response.Body.String()).To(Equal("created template"))
		},
		Entry(
			"full page JSON",
			"",
			true,
			"",
			`{"name":"Kitchen","description":"Reusable steps."}`,
		),
		Entry(
			"HTMX JSON",
			"true",
			false,
			"application/json",
			`{"name":"Kitchen","description":"Reusable steps."}`,
		),
		Entry(
			"HTMX form",
			"true",
			false,
			echo.MIMEApplicationForm,
			"name=Kitchen&description=Reusable+steps.",
		),
	)

	DescribeTable(
		"renders creation failures as HTML",
		func(ctx SpecContext, createErr error, status int, nameError, descriptionError, summary string) {
			service := NewMockService(GinkgoT())
			service.EXPECT().
				Create(ctx, &models.CreateChoreTemplateParams{Name: "Kitchen"}).
				Return(nil, createErr).
				Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().CreateForm(viewmodels.CreateForm{
				ActionHref: "/chore-templates",
				CancelHref: "/chore-templates",
				Name: viewmodels.Field{
					Value: "Kitchen",
					Error: nameError,
				},
				Description:  viewmodels.Field{Error: descriptionError},
				SummaryError: summary,
			}, true).Return(templ.Raw("template error")).Once()
			handler := choretemplates.NewHandler(
				&choretemplates.HandlerDeps{Service: service, View: view},
			)
			app := echo.New()
			app.POST("/chore-templates", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chore-templates",
				strings.NewReader(`{"name":"Kitchen"}`),
			)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(status))
			Expect(response.Body.String()).To(Equal("template error"))
		},
		Entry(
			"name conflict",
			models.ErrNameConflict,
			http.StatusConflict,
			"An active template already uses this name.", "", "",
		),
		Entry(
			"validation failure",
			validationerrors.ErrInvalidName,
			http.StatusUnprocessableEntity,
			"Enter a name between 1 and 200 characters.", "", "",
		),
		Entry(
			"description validation",
			validationerrors.ErrInvalidDescription,
			422,
			"",
			"Keep the description within 2000 characters.",
			"",
		),
		Entry(
			"storage failure",
			errors.New("storage failed"),
			http.StatusInternalServerError,
			"", "", "The template could not be created. Try again.",
		),
	)

	DescribeTable(
		"returns created templates as JSON independently of input encoding",
		func(ctx SpecContext, contentType, body string) {
			service := NewMockService(GinkgoT())
			service.EXPECT().
				Create(ctx, &models.CreateChoreTemplateParams{Name: "Kitchen", Description: "Steps."}).
				Return(&models.ChoreTemplate{Id: 7, Name: "Kitchen", Description: "Steps."}, nil).
				Once()
			handler := choretemplates.NewHandler(
				&choretemplates.HandlerDeps{Service: service},
			)
			app := echo.New()
			app.POST("/chore-templates", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chore-templates",
				strings.NewReader(body),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", contentType)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusCreated))
			Expect(response.Header().Get("Location")).To(Equal("/chore-templates/7"))
			Expect(response.Header().Get("HX-Replace-Url")).To(BeEmpty())
			Expect(response.Body.String()).To(ContainSubstring(`"name":"Kitchen"`))
		},
		Entry("JSON", "application/json", `{"name":"Kitchen","description":"Steps."}`),
		Entry("form", echo.MIMEApplicationForm, "name=Kitchen&description=Steps."),
	)

	DescribeTable(
		"preserves JSON error responses",
		func(ctx SpecContext, serviceErr error, status int, message string) {
			service := NewMockService(GinkgoT())
			service.EXPECT().
				Create(ctx, &models.CreateChoreTemplateParams{Name: "Kitchen"}).
				Return(nil, serviceErr).
				Once()
			handler := choretemplates.NewHandler(
				&choretemplates.HandlerDeps{Service: service},
			)
			app := echo.New()
			app.POST("/chore-templates", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chore-templates",
				strings.NewReader(`{"name":"Kitchen"}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
			Expect(
				response.Body.String(),
			).To(ContainSubstring(`"error":"` + message + `"`))
		},
		Entry("conflict", models.ErrNameConflict, 409, models.ErrNameConflict.Error()),
		Entry("name", validationerrors.ErrInvalidName, 422, "invalid name"),
		Entry(
			"description",
			validationerrors.ErrInvalidDescription,
			422,
			"invalid description",
		),
		Entry("storage", errors.New("storage"), 500, "something went wrong"),
	)

	DescribeTable(
		"rejects invalid requests before creation",
		func(ctx SpecContext, accept, contentType, body string, status int) {
			handler := choretemplates.NewHandler(
				&choretemplates.HandlerDeps{Service: NewMockService(GinkgoT())},
			)
			app := echo.New()
			app.POST("/chore-templates", handler.Create)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/chore-templates",
				strings.NewReader(body),
			)
			request.Header.Set("Accept", accept)
			request.Header.Set("Content-Type", contentType)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry(
			"unsupported representation",
			"image/png",
			"application/json",
			`{"name":"Kitchen"}`,
			406,
		),
		Entry("unsupported encoding", hypermedia.MediaType, "text/plain", "Kitchen", 415),
		Entry("malformed JSON", hypermedia.MediaType, "application/json", "{", 400),
		Entry("empty body", hypermedia.MediaType, "", "", 400),
	)
})
