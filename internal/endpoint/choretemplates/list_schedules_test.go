package choretemplates_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/a-h/templ"
	endpoint "github.com/bpsoos/shiftbell/internal/endpoint/choretemplates"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/choretemplates"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List template schedules", func() {
	var service *MockService
	var view *MockView
	var app *echo.Echo
	var request func(context.Context, string, string, string) *httptest.ResponseRecorder
	BeforeEach(func() {
		service = NewMockService(GinkgoT())
		view = NewMockView(GinkgoT())
		handler := endpoint.NewHandler(
			&endpoint.HandlerDeps{Service: service, View: view},
		)
		app = echo.New()
		app.GET("/chore-templates/:id/schedules", handler.ListSchedules)
		request = func(ctx context.Context, path, hx, target string) *httptest.ResponseRecorder {
			req := httptest.NewRequestWithContext(ctx, http.MethodGet, path, nil)
			req.Header.Set("Accept", "text/html")
			req.Header.Set("HX-Request", hx)
			req.Header.Set("HX-Target", target)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, req)
			return response
		}
	})
	DescribeTable(
		"renders the selected inline list",
		func(ctx SpecContext, query string, state schedulemodels.ScheduleFilter, heading, emptyMessage string) {
			service.EXPECT().
				ListSchedules(ctx, 8, state).
				Return(&models.TemplateSchedules{
					TemplateName: "Laundry",
					Schedules: []models.ScheduleReference{
						{Id: 42, Name: "Weekly laundry"},
					},
				}, nil).
				Once()
			view.EXPECT().ScheduleList(viewmodels.ScheduleList{
				Heading:      heading,
				EmptyMessage: emptyMessage,
				TemplateName: "Laundry",
				TemplateHref: "/chore-templates/8",
				State:        string(state),
				Total:        1,
				Inline:       true,
				Schedules: []viewmodels.ScheduleLink{
					{Name: "Weekly laundry", Href: "/schedules/42"},
				},
			}, false).Return(templ.Raw("schedule links")).Once()
			response := request(
				ctx,
				"/chore-templates/8/schedules"+query,
				"true",
				"template-schedule-list",
			)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("schedule links"))
			Expect(
				response.Header().Get("Vary"),
			).To(Equal("Accept, HX-Request, HX-Target"))
		},
		Entry(
			"default active",
			"",
			schedulemodels.ScheduleFilterActive,
			"Active schedules",
			"No active schedules.",
		),
		Entry(
			"active",
			"?state=active",
			schedulemodels.ScheduleFilterActive,
			"Active schedules",
			"No active schedules.",
		),
		Entry(
			"deactivated",
			"?state=deactivated",
			schedulemodels.ScheduleFilterDeactivated,
			"Deactivated schedules",
			"No deactivated schedules.",
		),
	)

	DescribeTable(
		"caps inline links without losing the total",
		func(ctx SpecContext, count int, viewAllHref string) {
			schedules := make([]models.ScheduleReference, 0, count)
			links := make([]viewmodels.ScheduleLink, 0, min(count, 10))
			for i := range count {
				schedules = append(
					schedules,
					models.ScheduleReference{
						Id:   i + 1,
						Name: "Laundry " + strconv.Itoa(i),
					},
				)
			}
			for _, schedule := range schedules[:min(count, 10)] {
				links = append(
					links,
					viewmodels.ScheduleLink{
						Name: schedule.Name,
						Href: "/schedules/" + strconv.Itoa(schedule.Id),
					},
				)
			}
			service.EXPECT().
				ListSchedules(ctx, 8, schedulemodels.ScheduleFilterActive).
				Return(&models.TemplateSchedules{TemplateName: "Laundry", Schedules: schedules}, nil).
				Once()
			view.EXPECT().ScheduleList(viewmodels.ScheduleList{
				Heading:      "Active schedules",
				EmptyMessage: "No active schedules.",
				TemplateName: "Laundry",
				TemplateHref: "/chore-templates/8",
				State:        "active",
				Total:        count,
				Inline:       true,
				ViewAllHref:  viewAllHref,
				Schedules:    links,
			}, false).Return(templ.Raw("inline schedules")).Once()
			response := request(
				ctx,
				"/chore-templates/8/schedules",
				"true",
				"template-schedule-list",
			)
			Expect(response.Code).To(Equal(http.StatusOK))
		},
		Entry("empty", 0, ""),
		Entry("exactly ten", 10, ""),
		Entry("overflow", 12, "/chore-templates/8/schedules?state=active"),
	)

	DescribeTable(
		"renders all links on the fallback page",
		func(ctx SpecContext, hx, target string, fullPage bool) {
			schedules := make([]models.ScheduleReference, 12)
			links := make([]viewmodels.ScheduleLink, 12)
			for i := range schedules {
				schedules[i] = models.ScheduleReference{
					Id:   i + 1,
					Name: "Laundry " + strconv.Itoa(i),
				}
				links[i] = viewmodels.ScheduleLink{
					Name: schedules[i].Name,
					Href: "/schedules/" + strconv.Itoa(i+1),
				}
			}
			service.EXPECT().
				ListSchedules(ctx, 8, schedulemodels.ScheduleFilterDeactivated).
				Return(&models.TemplateSchedules{TemplateName: "Laundry", Schedules: schedules}, nil).
				Once()
			view.EXPECT().ScheduleList(viewmodels.ScheduleList{
				Heading:      "Deactivated schedules",
				EmptyMessage: "No deactivated schedules.",
				TemplateName: "Laundry",
				TemplateHref: "/chore-templates/8",
				State:        "deactivated",
				Total:        12,
				Schedules:    links,
			}, fullPage).Return(templ.Raw("all schedules")).Once()
			response := request(
				ctx,
				"/chore-templates/8/schedules?state=deactivated",
				hx,
				target,
			)
			Expect(response.Code).To(Equal(http.StatusOK))
		},
		Entry("direct navigation", "", "", true),
		Entry("HTMX navigation", "true", "main", false),
	)

	DescribeTable(
		"renders inline load errors",
		func(ctx SpecContext, serviceErr error, status int, message string) {
			service.EXPECT().
				ListSchedules(ctx, 8, schedulemodels.ScheduleFilterActive).
				Return(nil, serviceErr).
				Once()
			view.EXPECT().
				ScheduleList(viewmodels.ScheduleList{Heading: "Unable to load schedules", Error: message, Inline: true}, false).
				Return(templ.Raw("list error")).
				Once()
			response := request(
				ctx,
				"/chore-templates/8/schedules",
				"true",
				"template-schedule-list",
			)
			Expect(response.Code).To(Equal(status))
		},
		Entry(
			"missing template",
			models.ErrNotFound,
			404,
			"This template could not be found.",
		),
		Entry(
			"invalid state",
			validationerrors.ErrInvalidFilter,
			422,
			"Invalid schedule state.",
		),
		Entry(
			"storage",
			errors.New("storage"),
			500,
			"The schedules could not be loaded. Try again.",
		),
	)

	DescribeTable(
		"rejects unsupported representations",
		func(ctx SpecContext, accept string) {
			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chore-templates/8/schedules",
				nil,
			)
			req.Header.Set("Accept", accept)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, req)
			Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		},
		Entry("JSON", hypermedia.MediaType),
		Entry("unsupported", "image/png"),
	)

	It("rejects invalid IDs without loading", func(ctx SpecContext) {
		view.EXPECT().
			ScheduleList(viewmodels.ScheduleList{Heading: "Unable to load schedules", Error: "invalid chore template id", Inline: true}, false).
			Return(templ.Raw("invalid ID")).
			Once()
		response := request(
			ctx,
			"/chore-templates/invalid/schedules",
			"true",
			"template-schedule-list",
		)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
	})
})
