package schedules_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/schedules"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Browse", func() {
	var (
		service *MockService
		app     *echo.Echo
	)
	BeforeEach(func() {
		service = NewMockService(GinkgoT())
		handler := schedules.NewHandler(&schedules.HandlerDeps{Service: service})
		app = echo.New()
		app.GET("/schedules", handler.Browse)
	})

	It("returns an empty JSON array and a self link", func(ctx SpecContext) {
		service.EXPECT().Browse(ctx, &models.BrowseSchedulesParams{Limit: 20}).
			Return(&models.SchedulePage{}, nil).
			Once()
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/schedules", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
		Expect(response.Header().Get("Vary")).To(Equal("Accept"))
		Expect(
			response.Body.Bytes(),
		).To(MatchJSON(`{"items":[],"more":false,"_links":[{"rel":"self","href":"/schedules"}]}`))
	})

	It("returns navigation for subsequent pages", func(ctx SpecContext) {
		service.EXPECT().
			Browse(ctx, &models.BrowseSchedulesParams{Limit: 20, Offset: 20}).
			Return(&models.SchedulePage{More: true}, nil).
			Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/schedules?offset=20",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(
			response.Body.Bytes(),
		).To(MatchJSON(`{"items":[],"more":true,"_links":[{"rel":"self","href":"/schedules?offset=20"},{"rel":"next","href":"/schedules?offset=40"},{"rel":"previous","href":"/schedules?offset=0"}]}`))
	})

	It("includes navigable links on each schedule item", func(ctx SpecContext) {
		choreID := 11
		service.EXPECT().Browse(ctx, &models.BrowseSchedulesParams{Limit: 20}).
			Return(&models.SchedulePage{Schedules: []models.Schedule{
				{
					Id:                42,
					Name:              "Weekly laundry",
					ChoreTemplateId:   7,
					ChoreTemplateName: "Wash clothes",
					IntervalDays:      7,
					ActiveChoreId:     &choreID,
				},
			}}, nil).Once()
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/schedules", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"items":[{"id":42,"name":"Weekly laundry","chore_template_id":7,"chore_template_name":"Wash clothes",
				"interval_days":7,"deactivated_at":null,"active_chore_id":11,"active_chore_deadline":null,
				"_links":[{"rel":"self","href":"/schedules/42"},{"rel":"collection","href":"/schedules"},
					{"rel":"chore_template","href":"/chore-templates/7"},{"rel":"active_chore","href":"/chores/11"}]}],
			"more":false,"_links":[{"rel":"self","href":"/schedules"}]
		}`))
	})

	It("conceals persistence failures", func(ctx SpecContext) {
		service.EXPECT().Browse(ctx, mock.Anything).
			Return((*models.SchedulePage)(nil), errors.New("database unavailable")).
			Once()
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/schedules", nil)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusInternalServerError))
		Expect(response.Body.Bytes()).To(MatchJSON(`{"error":"something went wrong"}`))
	})

	DescribeTable("rejects invalid offsets", func(ctx SpecContext, offset string) {
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/schedules?offset="+offset,
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		service.AssertNotCalled(GinkgoT(), "Browse", mock.Anything, mock.Anything)
	}, Entry("negative", "-1"), Entry("non-numeric", "invalid"))

	It("rejects unsupported representations", func(ctx SpecContext) {
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/schedules", nil)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		service.AssertNotCalled(GinkgoT(), "Browse", mock.Anything, mock.Anything)
	})
})
