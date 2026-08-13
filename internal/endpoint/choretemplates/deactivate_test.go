package choretemplates_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	choretemplatesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	confirmationviewmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deactivate chore template", func() {
	It("rejects an HTML response", func(ctx SpecContext) {
		handler := choretemplatesendpoint.NewHandler(&choretemplatesendpoint.HandlerDeps{
			Service: NewMockService(GinkgoT()),
		})
		e := echo.New()
		e.PUT("/chore-templates/:id/deactivation", handler.Deactivate)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPut,
			"/chore-templates/8/deactivation",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
	})

	It("deactivates a template from an HTMX dialog and redirects", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Deactivate(ctx, 8).Return(&choretemplatemodels.ChoreTemplate{
			Id:   8,
			Name: "Laundry",
		}, nil).Once()
		handler := choretemplatesendpoint.NewHandler(
			&choretemplatesendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.PUT("/chore-templates/:id/deactivation", handler.Deactivate)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPut,
			"/chore-templates/8/deactivation",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("HX-Redirect")).To(Equal("/chore-templates"))
		Expect(response.Header().Get("HX-Trigger")).To(Equal("templateDeactivated"))
		cookies := response.Result().Cookies()
		Expect(cookies).To(HaveLen(1))
		Expect(cookies[0].Name).To(Equal("shiftbell_template_flash"))
		Expect(cookies[0].Value).To(Equal("template-deactivated"))
	})

	It(
		"keeps the dialog open when active schedules block deactivation",
		func(ctx SpecContext) {
			blocked := &choretemplatemodels.ActiveScheduleReferencesError{
				Schedules: []choretemplatemodels.ActiveScheduleReference{
					{Id: 7, Name: "Every two weeks"},
				},
			}
			service := NewMockService(GinkgoT())
			service.EXPECT().Deactivate(ctx, 8).Return(nil, errors.Join(
				errors.New("deactivate chore template"),
				blocked,
			)).Once()
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
				Error:        "This template cannot be deactivated while active schedules use it.",
				Icon:         "archive",
			}).Return(templ.Raw("blocked dialog sentinel")).Once()
			handler := choretemplatesendpoint.NewHandler(
				&choretemplatesendpoint.HandlerDeps{Service: service, View: view},
			)
			e := echo.New()
			e.PUT("/chore-templates/:id/deactivation", handler.Deactivate)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				"/chore-templates/8/deactivation",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusConflict))
			Expect(response.Header().Get("HX-Redirect")).To(BeEmpty())
			Expect(response.Body.String()).To(Equal("blocked dialog sentinel"))
		},
	)

	It(
		"renders a non-actionable state for a stale inactive request",
		func(ctx SpecContext) {
			service := NewMockService(GinkgoT())
			service.EXPECT().Deactivate(ctx, 8).Return(nil, errors.Join(
				errors.New("deactivate chore template"),
				choretemplatemodels.ErrInactive,
			)).Once()
			service.EXPECT().Get(ctx, 8).Return(&choretemplatemodels.ChoreTemplateDetails{
				ChoreTemplate: choretemplatemodels.ChoreTemplate{
					Id:   8,
					Name: "Laundry",
				},
			}, nil).Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().ConfirmationDialog(confirmationviewmodels.Dialog{
				Heading: "Template already deactivated",
				Supporting: []string{
					"This template is no longer available for active use.",
				},
				Error: "chore template inactive",
			}).Return(templ.Raw("inactive dialog sentinel")).Once()
			handler := choretemplatesendpoint.NewHandler(
				&choretemplatesendpoint.HandlerDeps{Service: service, View: view},
			)
			e := echo.New()
			e.PUT("/chore-templates/:id/deactivation", handler.Deactivate)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				"/chore-templates/8/deactivation",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
			Expect(response.Body.String()).To(Equal("inactive dialog sentinel"))
		},
	)
})
