package schedules_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/schedules"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/schedules"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Schedule edit form", func() {
	var service *MockService
	var view *MockView
	var app *echo.Echo
	BeforeEach(func() {
		service = NewMockService(GinkgoT())
		view = NewMockView(GinkgoT())
		handler := schedules.NewHandler(
			&schedules.HandlerDeps{Service: service, View: view},
		)
		app = echo.New()
		app.GET("/schedules/:id/edit", handler.EditForm)
	})

	It("rejects an invalid schedule ID without loading", func(ctx SpecContext) {
		view.EXPECT().
			Error(viewmodels.Error{Message: "invalid schedule id", Links: api.Relations{{Rel: "collection", Href: "/schedules"}}}, true).
			Return(templ.Raw("invalid ID")).
			Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/schedules/invalid/edit",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
	})

	DescribeTable(
		"loads editable values",
		func(ctx SpecContext, hx string, fullPage bool) {
			service.EXPECT().
				Get(ctx, 42).
				Return(&models.Schedule{Id: 42, Name: "Kitchen", IntervalDays: 7}, nil).
				Once()
			view.EXPECT().EditForm(viewmodels.EditForm{
				ActionHref:   "/schedules/42",
				CancelHref:   "/schedules/42",
				Name:         "Kitchen",
				IntervalDays: "7",
			}, fullPage).Return(templ.Raw("edit form")).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/schedules/42/edit",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hx)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("edit form"))
		},
		Entry("full page", "", true),
		Entry("fragment", "true", false),
	)

	It("rejects deactivated schedules", func(ctx SpecContext) {
		now := time.Now()
		service.EXPECT().
			Get(ctx, 42).
			Return(&models.Schedule{Id: 42, ChoreTemplateId: 7, DeactivatedAt: &now}, nil).
			Once()
		view.EXPECT().
			Error(viewmodels.Error{Message: models.ErrInactive.Error(), Links: api.Relations{
				{
					Rel:  "self",
					Href: "/schedules/42",
				},
				{Rel: "collection", Href: "/schedules"},
				{Rel: "chore_template", Href: "/chore-templates/7"},
			}}, true).
			Return(templ.Raw("inactive")).
			Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/schedules/42/edit",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.String()).To(Equal("inactive"))
	})

	DescribeTable(
		"renders load failures",
		func(ctx SpecContext, err error, status int, message string) {
			service.EXPECT().Get(ctx, 42).Return(nil, err).Once()
			view.EXPECT().
				Error(viewmodels.Error{Message: message, Links: api.Relations{{Rel: "collection", Href: "/schedules"}}}, true).
				Return(templ.Raw("load error")).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/schedules/42/edit",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry("missing", models.ErrNotFound, 404, models.ErrNotFound.Error()),
		Entry("storage failure", errors.New("storage"), 500, "something went wrong"),
	)

	DescribeTable("rejects non-HTML requests", func(ctx SpecContext, accept string) {
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/schedules/42/edit",
			nil,
		)
		request.Header.Set("Accept", accept)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
	}, Entry("JSON", hypermedia.MediaType), Entry("unsupported", "image/png"))
})
