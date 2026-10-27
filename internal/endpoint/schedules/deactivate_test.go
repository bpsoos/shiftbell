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

var _ = Describe("Deactivate", func() {
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
		app.PUT("/schedules/:id/deactivation", handler.Deactivate)
	})

	DescribeTable(
		"renders deactivated schedule details as HTML",
		func(ctx SpecContext, hxRequest string, fullPage bool) {
			deactivatedAt := time.Date(2020, 4, 3, 10, 0, 0, 0, time.UTC)
			service.EXPECT().Deactivate(ctx, 42).Return(&models.Schedule{
				Id:                42,
				Name:              "Weekly kitchen",
				ChoreTemplateId:   7,
				ChoreTemplateName: "Kitchen",
				IntervalDays:      7,
				DeactivatedAt:     &deactivatedAt,
			}, nil).Once()
			view.EXPECT().Detail(viewmodels.Detail{Schedule: apimodels.Representation{
				Response: apimodels.Response{
					Id:                42,
					Name:              "Weekly kitchen",
					ChoreTemplateId:   7,
					ChoreTemplateName: "Kitchen",
					IntervalDays:      7,
					DeactivatedAt:     &deactivatedAt,
					Links: api.Relations{
						{Rel: "self", Href: "/schedules/42"},
						{Rel: "collection", Href: "/schedules"},
						{Rel: "chore_template", Href: "/chore-templates/7"},
					},
				},
				Actions: api.Relations{},
			}, Notice: "Schedule permanently deactivated."}, fullPage).Return(templ.Raw("deactivated schedule")).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				"/schedules/42/deactivation",
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
			Expect(response.Body.String()).To(Equal("deactivated schedule"))
		},
		Entry("full page", "", true),
		Entry("HTMX fragment", "true", false),
	)

	It("renders deactivation failures as HTML", func(ctx SpecContext) {
		service.EXPECT().Deactivate(ctx, 42).Return(nil, models.ErrNotFound).Once()
		view.EXPECT().
			Error(viewmodels.Error{Message: models.ErrNotFound.Error(), Links: api.Relations{{Rel: "collection", Href: "/schedules"}}}, true).
			Return(templ.Raw("deactivation error")).
			Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPut,
			"/schedules/42/deactivation",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotFound))
		Expect(response.Body.String()).To(Equal("deactivation error"))
	})

	It(
		"returns read-only schedule details without an active chore link",
		func(ctx SpecContext) {
			deactivatedAt := time.Date(2020, 4, 3, 10, 0, 0, 0, time.UTC)
			service.EXPECT().Deactivate(ctx, 42).Return(&models.Schedule{
				Id:                42,
				Name:              "Weekly laundry",
				ChoreTemplateId:   7,
				ChoreTemplateName: "Wash clothes",
				IntervalDays:      7,
				DeactivatedAt:     &deactivatedAt,
			}, nil).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				"/schedules/42/deactivation",
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
			"interval_days":7,"deactivated_at":"2020-04-03T10:00:00Z","active_chore_id":null,"active_chore_deadline":null,
			"_links":[{"rel":"self","href":"/schedules/42"},{"rel":"collection","href":"/schedules"},
				{"rel":"chore_template","href":"/chore-templates/7"}],"_actions":[]
		}`))
		},
	)

	DescribeTable(
		"maps deactivation errors",
		func(ctx SpecContext, serviceErr error, status int, message string) {
			service.EXPECT().
				Deactivate(ctx, 42).
				Return(nil, fmt.Errorf("deactivate schedule: %w", serviceErr)).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPut,
				"/schedules/42/deactivation",
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(status))
			Expect(response.Body.Bytes()).To(MatchJSON(fmt.Sprintf(`{
			"error":%q,"_links":[{"rel":"collection","href":"/schedules"}]
		}`, message)))
		},
		Entry(
			"not found",
			models.ErrNotFound,
			http.StatusNotFound,
			models.ErrNotFound.Error(),
		),
		Entry(
			"internal failure",
			errors.New("database unavailable"),
			http.StatusInternalServerError,
			"something went wrong",
		),
	)

	DescribeTable("rejects invalid IDs", func(ctx SpecContext, id string) {
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPut,
			"/schedules/"+id+"/deactivation",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(
			response.Body.Bytes(),
		).To(MatchJSON(`{"error":"invalid schedule id","_links":[{"rel":"collection","href":"/schedules"}]}`))
	}, Entry("zero", "0"), Entry("negative", "-1"), Entry("non-numeric", "invalid"))
})
