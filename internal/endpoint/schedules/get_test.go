package schedules_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/schedules"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	apimodels "github.com/bpsoos/shiftbell/internal/models/api/schedules"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/schedules"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get", func() {
	var (
		service *MockService
		view    *MockView
		app     *echo.Echo
	)
	BeforeEach(func() {
		service = NewMockService(GinkgoT())
		view = NewMockView(GinkgoT())
		handler := schedules.NewHandler(
			&schedules.HandlerDeps{Service: service, View: view},
		)
		app = echo.New()
		app.GET("/schedules/:id", handler.Get)
	})

	DescribeTable(
		"renders HTML schedule details",
		func(ctx SpecContext, hxRequest string, fullPage bool) {
			service.EXPECT().Get(ctx, 42).Return(&models.ScheduleDetails{
				Id:                42,
				Name:              "Weekly kitchen",
				ChoreTemplateId:   7,
				ChoreTemplateName: "Kitchen",
				IntervalDays:      7,
			}, nil).Once()
			view.EXPECT().Detail(viewmodels.Detail{Schedule: apimodels.Representation{
				Response: apimodels.Response{
					Id:                42,
					Name:              "Weekly kitchen",
					ChoreTemplateId:   7,
					ChoreTemplateName: "Kitchen",
					IntervalDays:      7,
					Links: api.Relations{
						{Rel: "self", Href: "/schedules/42"},
						{Rel: "collection", Href: "/schedules"},
						{Rel: "chore_template", Href: "/chore-templates/7"},
					},
				},
				Actions: api.Relations{
					{Rel: "edit", Href: "/schedules/42"},
					{Rel: "deactivate", Href: "/schedules/42/deactivation"},
				},
			}}, fullPage).Return(templ.Raw("schedule detail")).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/schedules/42",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hxRequest)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(
				response.Header().Get("Content-Type"),
			).To(Equal(echo.MIMETextHTMLCharsetUTF8))
			Expect(response.Body.String()).To(Equal("schedule detail"))
		},
		Entry("full page", "", true),
		Entry("HTMX fragment", "true", false),
	)

	DescribeTable(
		"renders HTML detail errors",
		func(ctx SpecContext, getErr error, status int, message string) {
			service.EXPECT().Get(ctx, 42).Return(nil, getErr).Once()
			view.EXPECT().
				Error(viewmodels.Error{Message: message, Links: api.Relations{{Rel: "collection", Href: "/schedules"}}}, true).
				Return(templ.Raw("detail error")).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/schedules/42",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(status))
			Expect(response.Body.String()).To(Equal("detail error"))
		},
		Entry(
			"missing schedule",
			models.ErrNotFound,
			http.StatusNotFound,
			"schedule not found",
		),
		Entry(
			"storage failure",
			errors.New("storage failed"),
			http.StatusInternalServerError,
			"something went wrong",
		),
	)

	It(
		"returns the schedule with navigable template and active chore links",
		func(ctx SpecContext) {
			choreID := 11
			deadline := time.Date(2020, 4, 2, 0, 0, 0, 0, time.UTC)
			service.EXPECT().Get(ctx, 42).Return(&models.ScheduleDetails{
				Id:                  42,
				Name:                "Weekly laundry",
				ChoreTemplateId:     7,
				ChoreTemplateName:   "Wash clothes",
				IntervalDays:        7,
				ActiveChoreId:       &choreID,
				ActiveChoreDeadline: &deadline,
			}, nil).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/schedules/42",
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Header().Get("Vary")).To(Equal("Accept"))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"id":42,"name":"Weekly laundry","chore_template_id":7,"chore_template_name":"Wash clothes",
			"interval_days":7,"deactivated_at":null,"active_chore_id":11,"active_chore_deadline":"2020-04-02T00:00:00Z",
			"_links":[{"rel":"self","href":"/schedules/42"},{"rel":"collection","href":"/schedules"},
				{"rel":"chore_template","href":"/chore-templates/7"},{"rel":"active_chore","href":"/chores/11"}],
			"_actions":[{"rel":"edit","href":"/schedules/42"},{"rel":"deactivate","href":"/schedules/42/deactivation"}]
		}`))
		},
	)

	It("omits the active chore link for a deactivated schedule", func(ctx SpecContext) {
		deactivatedAt := time.Date(2020, 4, 3, 10, 0, 0, 0, time.UTC)
		service.EXPECT().Get(ctx, 42).Return(&models.ScheduleDetails{
			Id:                42,
			Name:              "Weekly laundry",
			ChoreTemplateId:   7,
			ChoreTemplateName: "Wash clothes",
			IntervalDays:      7,
			DeactivatedAt:     &deactivatedAt,
		}, nil).Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/schedules/42",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"id":42,"name":"Weekly laundry","chore_template_id":7,"chore_template_name":"Wash clothes",
			"interval_days":7,"deactivated_at":"2020-04-03T10:00:00Z","active_chore_id":null,"active_chore_deadline":null,
			"_links":[{"rel":"self","href":"/schedules/42"},{"rel":"collection","href":"/schedules"},
				{"rel":"chore_template","href":"/chore-templates/7"}],"_actions":[]
		}`))
	})

	It("returns not found with a link back to the collection", func(ctx SpecContext) {
		service.EXPECT().Get(ctx, 42).
			Return((*models.ScheduleDetails)(nil), fmt.Errorf("get schedule: %w", models.ErrNotFound)).
			Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/schedules/42",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotFound))
		Expect(
			response.Body.Bytes(),
		).To(MatchJSON(`{"error":"schedule not found","_links":[{"rel":"collection","href":"/schedules"}]}`))
	})

	It("conceals service failures", func(ctx SpecContext) {
		service.EXPECT().Get(ctx, 42).
			Return((*models.ScheduleDetails)(nil), errors.New("database unavailable")).
			Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/schedules/42",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusInternalServerError))
		Expect(
			response.Body.Bytes(),
		).To(MatchJSON(`{"error":"something went wrong","_links":[{"rel":"collection","href":"/schedules"}]}`))
	})

	DescribeTable("rejects invalid schedule IDs", func(ctx SpecContext, id string) {
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/schedules/"+id,
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(
			response.Body.Bytes(),
		).To(MatchJSON(`{"error":"invalid schedule id","_links":[{"rel":"collection","href":"/schedules"}]}`))
	},
		Entry("zero", "0"),
		Entry("negative", "-1"),
		Entry("non-numeric", "invalid"),
		Entry("overflow", "999999999999999999999999999999"),
	)
})
