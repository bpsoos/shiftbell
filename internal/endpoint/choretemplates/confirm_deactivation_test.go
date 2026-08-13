package choretemplates_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/a-h/templ"
	choretemplatesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	confirmationviewmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Confirm chore template deactivation", func() {
	It("rejects an ordinary HTML request", func(ctx SpecContext) {
		handler := choretemplatesendpoint.NewHandler(
			&choretemplatesendpoint.HandlerDeps{Service: NewMockService(GinkgoT())},
		)
		e := echo.New()
		e.GET("/chore-templates/:id/deactivation", handler.ConfirmDeactivation)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates/8/deactivation",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Header().Get("Vary")).To(Equal("Accept, HX-Request"))
	})

	It("loads an active template and renders its dialog", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 8).Return(&choretemplatemodels.ChoreTemplateDetails{
			ChoreTemplate: choretemplatemodels.ChoreTemplate{
				Id:   8,
				Name: "Laundry",
			},
		}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().ConfirmationDialog(confirmationviewmodels.Dialog{
			Heading: "Deactivate template?",
			Prompt:  "Deactivate",
			Name:    "Laundry",
			Supporting: []string{
				"It will no longer appear in template selectors.",
				"Existing chores are not changed.",
				"This cannot be reversed.",
			},
			ActionHref:   "/chore-templates/8/deactivation",
			ActionMethod: "put",
			ActionLabel:  "Deactivate permanently",
			Icon:         "archive",
		}).Return(templ.Raw("deactivation dialog sentinel")).Once()
		handler := choretemplatesendpoint.NewHandler(
			&choretemplatesendpoint.HandlerDeps{Service: service, View: view},
		)
		e := echo.New()
		e.GET("/chore-templates/:id/deactivation", handler.ConfirmDeactivation)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates/8/deactivation",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("deactivation dialog sentinel"))
	})

	It("renders a non-actionable state for an inactive template", func(ctx SpecContext) {
		deactivatedAt := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
		service := NewMockService(GinkgoT())
		service.EXPECT().Get(ctx, 8).Return(&choretemplatemodels.ChoreTemplateDetails{
			ChoreTemplate: choretemplatemodels.ChoreTemplate{
				Id:            8,
				Name:          "Laundry",
				DeactivatedAt: &deactivatedAt,
			},
		}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().ConfirmationDialog(confirmationviewmodels.Dialog{
			Heading:    "Template already deactivated",
			Supporting: []string{"This template is no longer available for active use."},
			Error:      "chore template inactive",
		}).Return(templ.Raw("inactive dialog sentinel")).Once()
		handler := choretemplatesendpoint.NewHandler(
			&choretemplatesendpoint.HandlerDeps{Service: service, View: view},
		)
		e := echo.New()
		e.GET("/chore-templates/:id/deactivation", handler.ConfirmDeactivation)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chore-templates/8/deactivation",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(Equal("inactive dialog sentinel"))
	})
})
