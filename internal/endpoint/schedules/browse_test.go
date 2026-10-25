package schedules_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/schedules"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
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

	It("returns an empty JSON array with state navigation links", func(ctx SpecContext) {
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
		).To(MatchJSON(`{"items":[],"more":false,"_links":[{"rel":"self","href":"/schedules"},{"rel":"active","href":"/schedules?state=active"},{"rel":"deactivated","href":"/schedules?state=deactivated"}]}`))
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
		).To(MatchJSON(`{"items":[],"more":true,"_links":[{"rel":"self","href":"/schedules?offset=20"},{"rel":"active","href":"/schedules?state=active"},{"rel":"deactivated","href":"/schedules?state=deactivated"},{"rel":"next","href":"/schedules?offset=40"},{"rel":"previous","href":"/schedules?offset=0"}]}`))
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
			"more":false,"_links":[{"rel":"self","href":"/schedules"},{"rel":"active","href":"/schedules?state=active"},{"rel":"deactivated","href":"/schedules?state=deactivated"}]
		}`))
	})

	DescribeTable(
		"preserves state in pagination and resets offset when switching states",
		func(ctx SpecContext, state models.ScheduleFilter) {
			service.EXPECT().
				Browse(ctx, &models.BrowseSchedulesParams{Filter: state, Offset: 20, Limit: 20}).
				Return(&models.SchedulePage{More: true}, nil).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/schedules?state="+string(state)+"&offset=20",
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.Bytes()).To(MatchJSON(fmt.Sprintf(`{
			"items":[],"more":true,"_links":[
				{"rel":"self","href":"/schedules?state=%s&offset=20"},
				{"rel":"active","href":"/schedules?state=active"},
				{"rel":"deactivated","href":"/schedules?state=deactivated"},
				{"rel":"next","href":"/schedules?offset=40&state=%s"},
				{"rel":"previous","href":"/schedules?offset=0&state=%s"}
			]
		}`, state, state, state)))
		},
		Entry("active schedules", models.ScheduleFilterActive),
		Entry("deactivated schedules", models.ScheduleFilterDeactivated),
	)

	DescribeTable(
		"returns validation feedback for invalid states",
		func(ctx SpecContext, state string) {
			service.EXPECT().
				Browse(ctx, &models.BrowseSchedulesParams{Filter: models.ScheduleFilter(state), Limit: 20}).
				Return(nil, fmt.Errorf("browse schedules: %w", validationerrors.ErrInvalidFilter)).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/schedules?state="+state,
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
			Expect(response.Body.Bytes()).To(MatchJSON(`{"error":"invalid filter"}`))
		},
		Entry("combined state", "all"),
		Entry("unknown", "invalid"),
		Entry("wrong case", "Active"),
	)

	DescribeTable(
		"preserves search in pagination and state navigation",
		func(ctx SpecContext, state models.ScheduleFilter) {
			search := "Laundry & towels + folding"
			service.EXPECT().
				Browse(ctx, &models.BrowseSchedulesParams{Filter: state, Search: search, Offset: 20, Limit: 20}).
				Return(&models.SchedulePage{More: true}, nil).
				Once()
			encodedSearch := url.QueryEscape(search)
			request := httptest.NewRequestWithContext(ctx, http.MethodGet,
				"/schedules?state="+string(state)+"&offset=20&search="+encodedSearch, nil)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.Bytes()).To(MatchJSON(fmt.Sprintf(`{
			"items":[],"more":true,"_links":[
				{"rel":"self","href":"/schedules?state=%s&offset=20&search=%s"},
				{"rel":"active","href":"/schedules?search=%s&state=active"},
				{"rel":"deactivated","href":"/schedules?search=%s&state=deactivated"},
				{"rel":"next","href":"/schedules?offset=40&search=%s&state=%s"},
				{"rel":"previous","href":"/schedules?offset=0&search=%s&state=%s"}
			]
		}`, state, encodedSearch, encodedSearch, encodedSearch, encodedSearch, state, encodedSearch, state)))
		},
		Entry("active schedules", models.ScheduleFilterActive),
		Entry("deactivated schedules", models.ScheduleFilterDeactivated),
	)

	It("returns validation feedback for an invalid search", func(ctx SpecContext) {
		service.EXPECT().
			Browse(ctx, &models.BrowseSchedulesParams{Search: "invalid search", Limit: 20}).
			Return(nil, fmt.Errorf("browse schedules: %w", validationerrors.ErrInvalidSearch)).
			Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/schedules?search=invalid+search",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.Bytes()).To(MatchJSON(`{"error":"invalid search"}`))
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
