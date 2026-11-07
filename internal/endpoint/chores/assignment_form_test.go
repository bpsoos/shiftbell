package chores_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/a-h/templ"
	endpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	models "github.com/bpsoos/shiftbell/internal/models/chores"
	profilemodels "github.com/bpsoos/shiftbell/internal/models/profiles"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Assignment form", func() {
	var service *MockService
	var profiles *MockProfileService
	var view *MockView
	var app *echo.Echo

	BeforeEach(func() {
		service = NewMockService(GinkgoT())
		profiles = NewMockProfileService(GinkgoT())
		view = NewMockView(GinkgoT())
		handler := endpoint.NewHandler(&endpoint.HandlerDeps{
			Service:           service,
			ProfileService:    profiles,
			View:              view,
			ProfilesUIEnabled: true,
		})
		app = echo.New()
		app.GET("/chores/:id/assignment", handler.AssignmentForm)
	})

	DescribeTable(
		"renders only dropdown choices with the existing owner",
		func(ctx SpecContext, htmx string, status models.ChoreStatus) {
			available := []profilemodels.Profile{
				{Id: 7, Name: "Alex"},
				{Id: 8, Name: "Sam"},
			}
			service.EXPECT().
				Get(ctx, 42).
				Return(&models.Chore{Id: 42, Name: "Kitchen", Status: status, OwnerProfileId: new(7), OwnerProfileName: "Alex"}, nil).
				Once()
			profiles.EXPECT().List(ctx).Return(available, nil).Once()
			view.EXPECT().
				Assignment(mock.Anything).
				Run(func(model viewmodels.Assignment) {
					Expect(model.Profiles).To(Equal(available))
					Expect(model.Chore.Name).To(Equal("Kitchen"))
					Expect(model.Chore.OwnerProfileId).To(Equal(new(7)))
					Expect(model.Chore.OwnerProfileName).To(Equal("Alex"))
					Expect(
						model.Chore.Actions.Href("set_assignment"),
					).To(Equal("/chores/42/assignment"))
					Expect(
						model.Chore.Actions.Href("clear_assignment"),
					).To(Equal("/chores/42/assignment"))
				}).
				Return(templ.Raw("assignment form sentinel")).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/42/assignment",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", htmx)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("assignment form sentinel"))
			Expect(response.Header().Get("Cache-Control")).To(Equal("private, no-store"))
		},
		Entry("active HTML", "", models.ChoreStatusActive),
		Entry("completed HTMX", "true", models.ChoreStatusCompleted),
	)

	DescribeTable(
		"rejects unavailable form requests without loading data",
		func(ctx SpecContext, enabled bool, accept string, status int) {
			handler := endpoint.NewHandler(
				&endpoint.HandlerDeps{
					Service:           service,
					ProfileService:    profiles,
					View:              view,
					ProfilesUIEnabled: enabled,
				},
			)
			app := echo.New()
			app.GET("/chores/:id/assignment", handler.AssignmentForm)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/42/assignment",
				nil,
			)
			request.Header.Set("Accept", accept)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
		},
		Entry("flag disabled", false, "text/html", 404),
		Entry("JSON", true, hypermedia.MediaType, 406),
	)

	It("rejects invalid chore IDs before loading data", func(ctx SpecContext) {
		view.EXPECT().
			Assignment(viewmodels.Assignment{Error: "invalid chore id"}).
			Return(templ.Raw("invalid id")).
			Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores/no/assignment",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
	})

	DescribeTable(
		"handles chore lookup failures without loading profiles",
		func(ctx SpecContext, failure error, status int, message, retryHref string) {
			service.EXPECT().Get(ctx, 42).Return(nil, failure).Once()
			view.EXPECT().
				Assignment(viewmodels.Assignment{Error: message, RetryHref: retryHref}).
				Return(templ.Raw("lookup failed")).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/42/assignment",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(status))
			Expect(response.Body.String()).To(Equal("lookup failed"))
		},
		Entry("missing chore", models.ErrNotFound, 404, "Chore not found.", ""),
		Entry(
			"storage failure",
			errors.New("storage"),
			500,
			"Could not load assignment. Try again.",
			"/chores/42/assignment",
		),
	)

	It(
		"returns a recoverable error when profile choices cannot load",
		func(ctx SpecContext) {
			service.EXPECT().
				Get(ctx, 42).
				Return(&models.Chore{Id: 42, Status: models.ChoreStatusActive}, nil).
				Once()
			profiles.EXPECT().
				List(ctx).
				Return(nil, errors.New("private storage details")).
				Once()
			view.EXPECT().Assignment(viewmodels.Assignment{
				Error:     "Could not load profiles. Try again.",
				RetryHref: "/chores/42/assignment",
			}).Return(templ.Raw("profile lookup failed")).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/42/assignment",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusInternalServerError))
			Expect(response.Body.String()).To(Equal("profile lookup failed"))
		},
	)
})
