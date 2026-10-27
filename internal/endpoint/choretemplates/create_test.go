package choretemplates_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
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
		func(ctx SpecContext, hxRequest string, fullPage bool) {
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
				strings.NewReader(`{"name":"Kitchen","description":"Reusable steps."}`),
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hxRequest)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusCreated))
			Expect(
				response.Header().Get("Content-Type"),
			).To(Equal(echo.MIMETextHTMLCharsetUTF8))
			Expect(response.Header().Get("Location")).To(Equal("/chore-templates/7"))
			Expect(response.Body.String()).To(Equal("created template"))
		},
		Entry("full page", "", true),
		Entry("HTMX fragment", "true", false),
	)

	DescribeTable(
		"renders creation failures as HTML",
		func(ctx SpecContext, createErr error, status int, expected viewmodels.Error) {
			service := NewMockService(GinkgoT())
			service.EXPECT().
				Create(ctx, &models.CreateChoreTemplateParams{Name: "Kitchen"}).
				Return(nil, createErr).
				Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().Error(expected, true).Return(templ.Raw("template error")).Once()
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
			viewmodels.Error{Message: models.ErrNameConflict.Error()},
		),
		Entry(
			"validation failure",
			validationerrors.ErrInvalidName,
			http.StatusUnprocessableEntity,
			viewmodels.Error{
				Message: "invalid name",
				Links: []viewmodels.Link{
					{Label: "Back to chore templates", Href: "/chore-templates"},
				},
			},
		),
		Entry(
			"storage failure",
			errors.New("storage failed"),
			http.StatusInternalServerError,
			viewmodels.Error{
				Message: "something went wrong",
				Links: []viewmodels.Link{
					{Label: "Back to chore templates", Href: "/chore-templates"},
				},
			},
		),
	)
})
