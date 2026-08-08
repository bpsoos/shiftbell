package choretemplates_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/a-h/templ"
	choretemplatesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choretemplateapimodels "github.com/bpsoos/shiftbell/internal/models/api/choretemplates"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get chore template", func() {
	It("maps an active template to a read-only HTML detail", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 8).Return(&choretemplatemodels.ChoreTemplateDetails{
			ChoreTemplate: choretemplatemodels.ChoreTemplate{
				Id:          8,
				Name:        "Laundry",
				Description: "Wash and fold",
			},
		}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Detail(viewmodels.Detail{
			ChoreTemplate: choretemplateapimodels.Representation{
				Response: choretemplateapimodels.Response{
					Id:          8,
					Name:        "Laundry",
					Description: "Wash and fold",
					Links: map[string]api.Link{
						"self":       {Href: "/chore-templates/8"},
						"collection": {Href: "/chore-templates"},
					},
				},
				Actions: map[string]api.Action{
					"edit": {
						Href:        "/chore-templates/8",
						Method:      http.MethodPatch,
						ContentType: "application/json",
						Fields: []api.ActionField{
							{Name: "name", Type: "string", Required: true},
							{Name: "description", Type: "string"},
						},
					},
					"deactivate": {
						Href:   "/chore-templates/8/deactivation",
						Method: http.MethodPut,
					},
				},
			},
			CollectionHref: "/chore-templates",
		}, true).Return(templ.Raw("detail sentinel")).Once()
		handler := choretemplatesendpoint.NewHandler(&choretemplatesendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chore-templates/:id", handler.Get)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates/8",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("detail sentinel"))
	})

	It(
		"maps a deactivated template to a read-only HTML detail",
		func(ctx SpecContext) {
			deactivatedAt := time.Date(2026, time.August, 8, 12, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 8).Return(&choretemplatemodels.ChoreTemplateDetails{
				ChoreTemplate: choretemplatemodels.ChoreTemplate{
					Id:            8,
					Name:          "Laundry",
					DeactivatedAt: &deactivatedAt,
				},
			}, nil).Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().Detail(viewmodels.Detail{
				ChoreTemplate: choretemplateapimodels.Representation{
					Response: choretemplateapimodels.Response{
						Id:            8,
						Name:          "Laundry",
						DeactivatedAt: &deactivatedAt,
						Links: map[string]api.Link{
							"self":       {Href: "/chore-templates/8"},
							"collection": {Href: "/chore-templates"},
						},
					},
					Actions: map[string]api.Action{},
				},
				CollectionHref: "/chore-templates",
			}, true).Return(templ.Raw("read-only detail sentinel")).Once()
			handler := choretemplatesendpoint.NewHandler(
				&choretemplatesendpoint.HandlerDeps{
					Service: service,
					View:    view,
				},
			)
			e := echo.New()
			e.GET("/chore-templates/:id", handler.Get)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chore-templates/8",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("read-only detail sentinel"))
		},
	)
})
