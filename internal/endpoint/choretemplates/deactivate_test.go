package choretemplates_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	choretemplatesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
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
		"explains when a template disappears before confirmation is submitted",
		func(ctx SpecContext) {
			service := NewMockService(GinkgoT())
			service.EXPECT().
				Deactivate(ctx, 8).
				Return(nil, choretemplatemodels.ErrNotFound).
				Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().ConfirmationDialog(confirmationviewmodels.Dialog{
				Heading: "Unable to deactivate template",
				Error:   "This template could not be found. Close this dialog and refresh the templates.",
			}).Return(templ.Raw("missing dialog sentinel")).Once()
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

			Expect(response.Code).To(Equal(http.StatusNotFound))
			Expect(response.Body.String()).To(Equal("missing dialog sentinel"))
		},
	)

	It(
		"replaces a stale confirmation with the authoritative active inline list",
		func(ctx SpecContext) {
			blocked := &choretemplatemodels.ActiveScheduleReferencesError{
				Schedules: []choretemplatemodels.ActiveScheduleReference{
					{Id: 7, Name: "Every two weeks"},
					{Id: 9, Name: "Weekly laundry"},
				},
			}
			service := NewMockService(GinkgoT())
			service.EXPECT().Deactivate(ctx, 8).Return(nil, errors.Join(
				errors.New("deactivate chore template"),
				blocked,
			)).Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().ScheduleList(viewmodels.ScheduleList{
				Heading:      "Active schedules",
				EmptyMessage: "No active schedules.",
				TemplateHref: "/chore-templates/8",
				State:        "active",
				Total:        2,
				Inline:       true,
				Notice:       "Deactivate these active schedules before deactivating the template.",
				Schedules: []viewmodels.ScheduleLink{
					{Name: "Every two weeks", Href: "/schedules/7"},
					{Name: "Weekly laundry", Href: "/schedules/9"},
				},
			}, false).Return(templ.Raw("blocked list sentinel")).Once()
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
			Expect(
				response.Header().Get("HX-Retarget"),
			).To(Equal("#template-schedule-list"))
			Expect(response.Header().Get("HX-Reswap")).To(Equal("innerHTML"))
			Expect(
				response.Header().Get("HX-Trigger-After-Swap"),
			).To(Equal("templateDeactivationBlocked"))
			Expect(response.Body.String()).To(Equal("blocked list sentinel"))
		},
	)

	It(
		"returns a JSON conflict when active schedules block deactivation",
		func(ctx SpecContext) {
			blocked := &choretemplatemodels.ActiveScheduleReferencesError{
				Schedules: []choretemplatemodels.ActiveScheduleReference{
					{Id: 7, Name: "Weekly"},
				},
			}
			service := NewMockService(GinkgoT())
			service.EXPECT().Deactivate(ctx, 8).Return(nil, errors.Join(
				errors.New("deactivate chore template"), blocked,
			)).Once()
			handler := choretemplatesendpoint.NewHandler(
				&choretemplatesendpoint.HandlerDeps{Service: service},
			)
			app := echo.New()
			app.PUT("/chore-templates/:id/deactivation", handler.Deactivate)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				"/chore-templates/8/deactivation",
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusConflict))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"error":"chore template has 1 active schedule references",
			"_links":[{"rel":"self","href":"/chore-templates/8"},{"rel":"collection","href":"/chore-templates"}],
			"_actions":[]
		}`))
			Expect(response.Header().Get("Set-Cookie")).To(BeEmpty())
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
				Id:   8,
				Name: "Laundry",
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
