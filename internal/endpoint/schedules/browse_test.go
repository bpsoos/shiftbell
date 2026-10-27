package schedules_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/endpoint/schedules"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	apimodels "github.com/bpsoos/shiftbell/internal/models/api/schedules"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/schedules"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Browse", func() {
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
		app.GET("/schedules", handler.Browse)
	})

	DescribeTable(
		"renders active schedule rows and pagination as HTML",
		func(ctx SpecContext, hxRequest string, fullPage bool) {
			choreId := 11
			deadline := time.Date(2020, 4, 2, 0, 0, 0, 0, time.UTC)
			service.EXPECT().
				Browse(ctx, &models.BrowseSchedulesParams{Offset: 20, Limit: 20}).
				Return(&models.SchedulePage{Schedules: []models.Schedule{
					{
						Id:                  42,
						Name:                "Weekly kitchen",
						ChoreTemplateId:     7,
						ChoreTemplateName:   "Kitchen",
						IntervalDays:        7,
						ActiveChoreId:       &choreId,
						ActiveChoreDeadline: &deadline,
					},
				}, More: true}, nil).
				Once()
			view.EXPECT().Collection(viewmodels.Collection{
				Collection: apimodels.CollectionResponse{
					Items: []apimodels.Response{
						{
							Id:                  42,
							Name:                "Weekly kitchen",
							ChoreTemplateId:     7,
							ChoreTemplateName:   "Kitchen",
							IntervalDays:        7,
							ActiveChoreId:       &choreId,
							ActiveChoreDeadline: &deadline,
							Links: api.Relations{
								{Rel: "self", Href: "/schedules/42"},
								{Rel: "collection", Href: "/schedules"},
								{Rel: "chore_template", Href: "/chore-templates/7"},
								{Rel: "active_chore", Href: "/chores/11"},
							},
						},
					}, More: true,
					Links: api.Relations{
						{Rel: "self", Href: "/schedules?offset=20"},
						{Rel: "active", Href: "/schedules?state=active"},
						{Rel: "deactivated", Href: "/schedules?state=deactivated"},
						{Rel: "next", Href: "/schedules?offset=40"},
						{Rel: "previous", Href: "/schedules?offset=0"},
					},
				}, Filter: models.ScheduleFilterActive,
			}, fullPage).Return(templ.Raw("active schedule page")).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/schedules?offset=20",
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
			Expect(response.Body.String()).To(Equal("active schedule page"))
		},
		Entry("full page", "", true),
		Entry("HTMX fragment", "true", false),
	)

	DescribeTable(
		"renders the HTML collection with state and search",
		func(ctx SpecContext, hxRequest string, fullPage bool) {
			service.EXPECT().
				Browse(ctx, &models.BrowseSchedulesParams{Filter: models.ScheduleFilterDeactivated, Search: "kitchen", Limit: 20}).
				Return(&models.SchedulePage{}, nil).
				Once()
			view.EXPECT().Collection(viewmodels.Collection{
				Collection: apimodels.CollectionResponse{
					Items: []apimodels.Response{},
					Links: api.Relations{
						{
							Rel:  "self",
							Href: "/schedules?state=deactivated&search=kitchen",
						},
						{Rel: "active", Href: "/schedules?search=kitchen&state=active"},
						{
							Rel:  "deactivated",
							Href: "/schedules?search=kitchen&state=deactivated",
						},
					},
				},
				Filter:     models.ScheduleFilterDeactivated,
				Search:     "kitchen",
				SearchOpen: true,
			}, fullPage).Return(templ.Raw("schedule collection")).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/schedules?state=deactivated&search=kitchen",
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
			Expect(
				strings.Split(response.Header().Get("Vary"), ", "),
			).To(ConsistOf("Accept", "HX-Request", "HX-Current-URL"))
			Expect(response.Body.String()).To(Equal("schedule collection"))
		},
		Entry("full page", "", true),
		Entry("HTMX fragment", "true", false),
	)

	DescribeTable(
		"focuses HTML search only when its value or visibility changed",
		func(ctx SpecContext, currentURL string, autofocus bool) {
			service.EXPECT().
				Browse(ctx, &models.BrowseSchedulesParams{Search: "kitchen", Limit: 20}).
				Return(&models.SchedulePage{}, nil).
				Once()
			view.EXPECT().Collection(viewmodels.Collection{
				Collection: apimodels.CollectionResponse{
					Items: []apimodels.Response{},
					Links: api.Relations{
						{Rel: "self", Href: "/schedules?search=kitchen"},
						{Rel: "active", Href: "/schedules?search=kitchen&state=active"},
						{
							Rel:  "deactivated",
							Href: "/schedules?search=kitchen&state=deactivated",
						},
					},
				},
				Filter:          models.ScheduleFilterActive,
				Search:          "kitchen",
				SearchOpen:      true,
				AutofocusSearch: autofocus,
			}, false).Return(templ.Raw("search results")).Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/schedules?search=kitchen",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			request.Header.Set("HX-Current-URL", currentURL)
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("search results"))
		},
		Entry("no current URL", "", false),
		Entry("opening search", "http://example.com/schedules", true),
		Entry("typing", "http://example.com/schedules?search=kit", true),
		Entry(
			"pagination",
			"http://example.com/schedules?search=kitchen&offset=20",
			false,
		),
		Entry(
			"state switching",
			"http://example.com/schedules?state=deactivated&search=kitchen",
			false,
		),
		Entry(
			"another collection",
			"http://example.com/chore-templates?search=kit",
			false,
		),
		Entry("malformed current URL", "%zz", false),
	)

	It("opens an empty HTML search field", func(ctx SpecContext) {
		service.EXPECT().
			Browse(ctx, &models.BrowseSchedulesParams{Limit: 20}).
			Return(&models.SchedulePage{}, nil).
			Once()
		view.EXPECT().Collection(viewmodels.Collection{
			Collection: apimodels.CollectionResponse{
				Items: []apimodels.Response{},
				Links: api.Relations{
					{Rel: "self", Href: "/schedules?search="},
					{Rel: "active", Href: "/schedules?search=&state=active"},
					{Rel: "deactivated", Href: "/schedules?search=&state=deactivated"},
				},
			},
			Filter:          models.ScheduleFilterActive,
			SearchOpen:      true,
			AutofocusSearch: true,
		}, false).Return(templ.Raw("empty search")).Once()
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/schedules?search=",
			nil,
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("HX-Request", "true")
		request.Header.Set("HX-Current-URL", "http://example.com/schedules")
		response := httptest.NewRecorder()

		app.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("empty search"))
	})

	DescribeTable(
		"renders HTML collection errors",
		func(ctx SpecContext, browseErr error, status int, message string) {
			service.EXPECT().
				Browse(ctx, &models.BrowseSchedulesParams{Limit: 20}).
				Return(nil, browseErr).
				Once()
			view.EXPECT().
				Error(viewmodels.Error{Message: message, Links: api.Relations{{Rel: "collection", Href: "/schedules"}}}, true).
				Return(templ.Raw("collection error")).
				Once()
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/schedules",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			response := httptest.NewRecorder()

			app.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(status))
			Expect(response.Body.String()).To(Equal("collection error"))
		},
		Entry(
			"invalid filter",
			validationerrors.ErrInvalidFilter,
			http.StatusUnprocessableEntity,
			"invalid filter",
		),
		Entry(
			"invalid search",
			validationerrors.ErrInvalidSearch,
			http.StatusUnprocessableEntity,
			"invalid search",
		),
		Entry(
			"storage failure",
			errors.New("storage failed"),
			http.StatusInternalServerError,
			"something went wrong",
		),
	)

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
})
