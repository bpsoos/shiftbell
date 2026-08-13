package choretemplates_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/a-h/templ"
	choretemplatesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choretemplateapimodels "github.com/bpsoos/shiftbell/internal/models/api/choretemplates"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edit chore template", func() {
	It("preserves the vendor JSON behavior", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Edit(ctx, &choretemplatemodels.EditChoreTemplateParams{
			Id:          8,
			Name:        "Laundry edited",
			Description: "Wash, dry, and fold",
		}).Return(&choretemplatemodels.ChoreTemplate{
			Id:          8,
			Name:        "Laundry edited",
			Description: "Wash, dry, and fold",
		}, nil).Once()
		handler := choretemplatesendpoint.NewHandler(
			&choretemplatesendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.PATCH("/chore-templates/:id", handler.Edit)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPatch,
			"/chore-templates/8",
			strings.NewReader(`{
				"name": "Laundry edited",
				"description": "Wash, dry, and fold"
			}`),
		)
		request.Header.Set("Accept", "application/vnd.shiftbell+json")
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"id": 8,
			"name": "Laundry edited",
			"description": "Wash, dry, and fold",
			"deactivated_at": null,
			"_links": [
				{"rel": "self", "href": "/chore-templates/8"},
				{"rel": "collection", "href": "/chore-templates"}
			],
			"_actions": [
				{"rel": "edit", "href": "/chore-templates/8"},
				{"rel": "deactivate", "href": "/chore-templates/8/deactivation"}
			]
		}`))
	})

	It("edits a template from an HTMX form", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Edit(ctx, &choretemplatemodels.EditChoreTemplateParams{
			Id:          8,
			Name:        "Laundry edited",
			Description: "Wash, dry, and fold",
		}).Return(&choretemplatemodels.ChoreTemplate{
			Id:          8,
			Name:        "Laundry edited",
			Description: "Wash, dry, and fold",
		}, nil).Once()
		representation := choretemplateapimodels.Representation{
			Response: choretemplateapimodels.Response{
				Id:          8,
				Name:        "Laundry edited",
				Description: "Wash, dry, and fold",
				Links: api.Relations{
					{Rel: "self", Href: "/chore-templates/8"},
					{Rel: "collection", Href: "/chore-templates"},
				},
			},
			Actions: api.Relations{
				{Rel: "edit", Href: "/chore-templates/8"},
				{Rel: "deactivate", Href: "/chore-templates/8/deactivation"},
			},
		}
		view := NewMockView(GinkgoT())
		view.EXPECT().Detail(viewmodels.Detail{
			ChoreTemplate:  representation,
			CollectionHref: "/chore-templates",
			EditHref:       "/chore-templates/8/edit",
			DeactivateHref: "/chore-templates/8/deactivation",
			Notice:         "Template updated.",
		}, false).Return(templ.Raw("updated detail sentinel")).Once()
		handler := choretemplatesendpoint.NewHandler(
			&choretemplatesendpoint.HandlerDeps{Service: service, View: view},
		)
		e := echo.New()
		e.PATCH("/chore-templates/:id", handler.Edit)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPatch,
			"/chore-templates/8",
			strings.NewReader(
				"name=Laundry+edited&description=Wash%2C+dry%2C+and+fold",
			),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("HX-Replace-Url")).To(Equal("/chore-templates/8"))
		Expect(response.Body.String()).To(Equal("updated detail sentinel"))
	})

	It("preserves values and shows field validation feedback", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Edit(ctx, &choretemplatemodels.EditChoreTemplateParams{
			Id:          8,
			Name:        "",
			Description: "Submitted description",
		}).Return(nil, errors.Join(
			errors.New("edit chore template"),
			validationerrors.ErrInvalidName,
		)).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().EditForm(viewmodels.EditForm{
			ActionHref: "/chore-templates/8",
			CancelHref: "/chore-templates/8",
			Name: viewmodels.Field{
				Error: "invalid name",
			},
			Description: viewmodels.Field{
				Value: "Submitted description",
			},
			SummaryError: "invalid name",
		}, false).Return(templ.Raw("validation sentinel")).Once()
		handler := choretemplatesendpoint.NewHandler(
			&choretemplatesendpoint.HandlerDeps{Service: service, View: view},
		)
		e := echo.New()
		e.PATCH("/chore-templates/:id", handler.Edit)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPatch,
			"/chore-templates/8",
			strings.NewReader("name=&description=Submitted+description"),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(Equal("validation sentinel"))
	})

	It("puts duplicate-name conflicts against the name field", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Edit(ctx, &choretemplatemodels.EditChoreTemplateParams{
			Id:          8,
			Name:        "Kitchen",
			Description: "Submitted description",
		}).Return(nil, errors.Join(
			errors.New("edit chore template"),
			choretemplatemodels.ErrNameConflict,
		)).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().EditForm(viewmodels.EditForm{
			ActionHref: "/chore-templates/8",
			CancelHref: "/chore-templates/8",
			Name: viewmodels.Field{
				Value: "Kitchen",
				Error: "chore template name conflicts with an active chore template",
			},
			Description: viewmodels.Field{
				Value: "Submitted description",
			},
			SummaryError: "chore template name conflicts with an active chore template",
		}, false).Return(templ.Raw("conflict sentinel")).Once()
		handler := choretemplatesendpoint.NewHandler(
			&choretemplatesendpoint.HandlerDeps{Service: service, View: view},
		)
		e := echo.New()
		e.PATCH("/chore-templates/:id", handler.Edit)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPatch,
			"/chore-templates/8",
			strings.NewReader("name=Kitchen&description=Submitted+description"),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusConflict))
		Expect(response.Body.String()).To(Equal("conflict sentinel"))
	})

	It("stops showing the form when the template became inactive", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Edit(ctx, &choretemplatemodels.EditChoreTemplateParams{
			Id:          8,
			Name:        "Laundry",
			Description: "Submitted description",
		}).Return(nil, errors.Join(
			errors.New("edit chore template"),
			choretemplatemodels.ErrInactive,
		)).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Error(viewmodels.Error{
			Message: "chore template inactive",
			Links: []viewmodels.Link{
				{Label: "Back to chore template", Href: "/chore-templates/8"},
				{Label: "Back to chore templates", Href: "/chore-templates"},
			},
		}, false).Return(templ.Raw("inactive sentinel")).Once()
		handler := choretemplatesendpoint.NewHandler(
			&choretemplatesendpoint.HandlerDeps{Service: service, View: view},
		)
		e := echo.New()
		e.PATCH("/chore-templates/:id", handler.Edit)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPatch,
			"/chore-templates/8",
			strings.NewReader("name=Laundry&description=Submitted+description"),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(Equal("inactive sentinel"))
	})
})
