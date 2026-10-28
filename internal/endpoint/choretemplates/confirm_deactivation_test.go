package choretemplates_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/a-h/templ"
	choretemplatesendpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
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
			Id:   8,
			Name: "Laundry",
		}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().ConfirmationDialog(confirmationviewmodels.Dialog{
			Heading:      "Deactivate template?",
			Prompt:       "Deactivate",
			Name:         "Laundry",
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
			Id:            8,
			Name:          "Laundry",
			DeactivatedAt: &deactivatedAt,
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

	DescribeTable(
		"renders a non-actionable dialog when the template cannot be loaded",
		func(ctx SpecContext, loadError error, status int, message string) {
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 8).Return(nil, loadError).Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().ConfirmationDialog(confirmationviewmodels.Dialog{
				Heading: "Unable to deactivate template", Error: message,
			}).Return(templ.Raw("unavailable dialog sentinel")).Once()
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

			Expect(response.Code).To(Equal(status))
			Expect(response.Body.String()).To(Equal("unavailable dialog sentinel"))
		},
		Entry(
			"missing template",
			choretemplatemodels.ErrNotFound,
			http.StatusNotFound,
			"This template could not be found. Close this dialog and refresh the templates.",
		),
		Entry(
			"unavailable storage",
			errors.New("storage"),
			http.StatusInternalServerError,
			"The template could not be loaded. Close this dialog and try again.",
		),
	)

	DescribeTable(
		"reveals active references instead of opening confirmation",
		func(ctx SpecContext, count int, viewAllHref string) {
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 8).Return(&choretemplatemodels.ChoreTemplateDetails{
				Id: 8, Name: "Laundry", ActiveScheduleCount: count,
			}, nil).Once()
			schedules := make([]choretemplatemodels.ScheduleReference, count)
			links := make([]viewmodels.ScheduleLink, min(count, 10))
			for i := range schedules {
				schedules[i] = choretemplatemodels.ScheduleReference{
					Id:   i + 1,
					Name: "Laundry " + strconv.Itoa(i),
				}
			}
			for i := range links {
				links[i] = viewmodels.ScheduleLink{
					Name: schedules[i].Name,
					Href: "/schedules/" + strconv.Itoa(i+1),
				}
			}
			service.EXPECT().
				ListSchedules(ctx, 8, schedulemodels.ScheduleFilterActive).
				Return(&choretemplatemodels.TemplateSchedules{TemplateName: "Laundry", Schedules: schedules}, nil).
				Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().ScheduleList(viewmodels.ScheduleList{
				Heading:      "Active schedules",
				EmptyMessage: "No active schedules.",
				TemplateHref: "/chore-templates/8",
				State:        "active",
				Total:        count,
				Inline:       true,
				Notice:       "Deactivate these active schedules before deactivating the template.",
				ViewAllHref:  viewAllHref,
				Schedules:    links,
			}, false).Return(templ.Raw("blocked schedules")).Once()
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
			Expect(response.Code).To(Equal(http.StatusConflict))
			Expect(
				response.Header().Get("HX-Retarget"),
			).To(Equal("#template-schedule-list"))
			Expect(response.Header().Get("HX-Reswap")).To(Equal("innerHTML"))
			Expect(
				response.Header().Get("HX-Trigger-After-Swap"),
			).To(Equal("templateDeactivationBlocked"))
			Expect(response.Body.String()).To(Equal("blocked schedules"))
		},
		Entry("a few schedules", 2, ""),
		Entry("more than ten", 12, "/chore-templates/8/schedules?state=active"),
	)
})
