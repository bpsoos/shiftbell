package choretemplates_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/a-h/templ"
	choretemplatesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edit chore template form", func() {
	It("rejects an ordinary HTML request", func(ctx SpecContext) {
		handler := choretemplatesendpoint.NewHandler(
			&choretemplatesendpoint.HandlerDeps{Service: NewMockService(GinkgoT())},
		)
		e := echo.New()
		e.GET("/chore-templates/:id/edit", handler.EditForm)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates/8/edit",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
	})

	It("renders an active template edit form", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 8).Return(&choretemplatemodels.ChoreTemplateDetails{
			ChoreTemplate: choretemplatemodels.ChoreTemplate{
				Id:          8,
				Name:        "Laundry",
				Description: "Wash and fold",
			},
		}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().EditForm(viewmodels.EditForm{
			ActionHref: "/chore-templates/8",
			CancelHref: "/chore-templates/8",
			Name: viewmodels.Field{
				Value: "Laundry",
			},
			Description: viewmodels.Field{
				Value: "Wash and fold",
			},
		}, false).Return(templ.Raw("edit form sentinel")).Once()
		handler := choretemplatesendpoint.NewHandler(
			&choretemplatesendpoint.HandlerDeps{Service: service, View: view},
		)
		e := echo.New()
		e.GET("/chore-templates/:id/edit", handler.EditForm)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates/8/edit",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("edit form sentinel"))
	})

	It("rejects a deactivated template", func(ctx SpecContext) {
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
		e.GET("/chore-templates/:id/edit", handler.EditForm)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates/8/edit",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(Equal("inactive sentinel"))
	})

	It("returns not found for a missing template", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 8).Return(nil, errors.Join(
			errors.New("get chore template"),
			choretemplatemodels.ErrNotFound,
		)).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Error(viewmodels.Error{
			Message: "chore template not found",
			Links: []viewmodels.Link{
				{Label: "Back to chore templates", Href: "/chore-templates"},
			},
		}, false).Return(templ.Raw("missing sentinel")).Once()
		handler := choretemplatesendpoint.NewHandler(
			&choretemplatesendpoint.HandlerDeps{Service: service, View: view},
		)
		e := echo.New()
		e.GET("/chore-templates/:id/edit", handler.EditForm)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates/8/edit",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotFound))
		Expect(response.Body.String()).To(Equal("missing sentinel"))
	})
})
