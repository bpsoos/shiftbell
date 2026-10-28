package chores_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/a-h/templ"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	models "github.com/bpsoos/shiftbell/internal/models/chores"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Completion correction form", func() {
	var service *MockService
	var view *MockView
	var app *echo.Echo
	BeforeEach(func() {
		service = NewMockService(GinkgoT())
		view = NewMockView(GinkgoT())
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service, View: view},
		)
		app = echo.New()
		app.GET("/chores/:id/completion/edit", handler.CompletionCorrectionForm)
	})

	DescribeTable(
		"prefills the existing completion date",
		func(ctx SpecContext, hx string, fullPage bool) {
			service.EXPECT().
				Get(ctx, 42).
				Return(&models.ChoreDetails{Id: 42, Status: models.ChoreStatusCompleted, CompletedOn: time.Date(2020, 2, 4, 0, 0, 0, 0, time.UTC)}, nil).
				Once()
			view.EXPECT().CompletionCorrectionForm(viewmodels.CompletionCorrectionForm{
				ActionHref:  "/chores/42/completion",
				CancelHref:  "/chores/42",
				CompletedOn: viewmodels.Field{Value: "2020-02-04"},
			}, fullPage).Return(templ.Raw("correction form")).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/42/completion/edit",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", hx)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("correction form"))
		},
		Entry("full page", "", true),
		Entry("HTMX", "true", false),
	)

	It("rejects an active chore", func(ctx SpecContext) {
		service.EXPECT().
			Get(ctx, 42).
			Return(&models.ChoreDetails{Id: 42, Status: models.ChoreStatusActive}, nil).
			Once()
		view.EXPECT().
			Error(viewmodels.Error{Message: "Only completed chores have a completion date to correct."}, true).
			Return(templ.Raw("active chore")).
			Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42/completion/edit",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
	})

	DescribeTable(
		"renders load errors",
		func(ctx SpecContext, err error, status int, expected viewmodels.Error) {
			service.EXPECT().Get(ctx, 42).Return(nil, err).Once()
			view.EXPECT().Error(expected, true).Return(templ.Raw("load error")).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/42/completion/edit",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry(
			"missing",
			models.ErrNotFound,
			404,
			viewmodels.Error{
				Message: models.ErrNotFound.Error(),
				Links:   []viewmodels.Link{{Label: "Back to chores", Href: "/chores"}},
			},
		),
		Entry(
			"storage",
			errors.New("storage"),
			500,
			viewmodels.Error{Message: "something went wrong"},
		),
	)

	It("rejects an invalid ID without loading", func(ctx SpecContext) {
		view.EXPECT().
			Error(viewmodels.Error{Message: "invalid chore id"}, true).
			Return(templ.Raw("invalid ID")).
			Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/invalid/completion/edit",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
	})

	DescribeTable("rejects non-HTML requests", func(ctx SpecContext, accept string) {
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/42/completion/edit",
			nil,
		)
		request.Header.Set("Accept", accept)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
	}, Entry("JSON", hypermedia.MediaType), Entry("unsupported", "image/png"))
})
