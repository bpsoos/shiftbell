package schedules_test

import (
	"context"
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

type mockService struct {
	mock.Mock
}

func (m *mockService) Browse(
	ctx context.Context,
	params *models.BrowseSchedulesParams,
) (*models.SchedulePage, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*models.SchedulePage), args.Error(1)
}

var _ = Describe("Browse", func() {
	var (
		service *mockService
		app     *echo.Echo
	)
	BeforeEach(func() {
		service = &mockService{}
		service.Test(GinkgoT())
		DeferCleanup(func() { service.AssertExpectations(GinkgoT()) })
		handler := schedules.NewHandler(&schedules.HandlerDeps{Service: service})
		app = echo.New()
		app.GET("/schedules", handler.Browse)
	})

	It("returns an empty JSON array and a self link", func(ctx SpecContext) {
		service.On("Browse", mock.Anything, &models.BrowseSchedulesParams{Limit: 20}).
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
		service.On("Browse", mock.Anything, &models.BrowseSchedulesParams{Limit: 20, Offset: 20}).
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

	It("conceals persistence failures", func(ctx SpecContext) {
		service.On("Browse", mock.Anything, mock.Anything).
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
