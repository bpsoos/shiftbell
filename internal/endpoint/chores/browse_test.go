package chores_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/a-h/templ"
	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choreapimodels "github.com/bpsoos/shiftbell/internal/models/api/chores"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Browse chores", func() {
	It("rejects a non-numeric offset", func(ctx SpecContext) {
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: NewMockService(GinkgoT())},
		)
		e := echo.New()
		e.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores?offset=invalid",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"error": "invalid offset",
			"_links": [],
			"_actions": []
		}`))
	})

	It("rejects a non-numeric limit", func(ctx SpecContext) {
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: NewMockService(GinkgoT())},
		)
		e := echo.New()
		e.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores?limit=invalid",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		Expect(response.Body.Bytes()).To(MatchJSON(`{
			"error": "invalid limit",
			"_links": [],
			"_actions": []
		}`))
	})

	It("passes filters and pagination to the service", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
			Status: choremodels.ChoreStatusCompleted,
			Search: "Kitchen",
			Offset: 4,
			Limit:  7,
		}).Return(&choremodels.ChorePage{}, nil).Once()
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/chores?status=completed&search=Kitchen&offset=4&limit=7",
			nil,
		)
		request.Header.Set("Accept", hypermedia.MediaType)
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
	})

	It(
		"returns the active chore collection with its creation affordance",
		func(ctx SpecContext) {
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
				Offset: 0,
				Limit:  20,
			}).Return(&choremodels.ChorePage{
				Chores: []choremodels.Chore{
					{
						Id:          42,
						Status:      choremodels.ChoreStatusActive,
						Name:        "Kitchen",
						Description: "Clean counters",
						Deadline:    deadline,
					},
				},
			}, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			e := echo.New()
			e.GET("/chores", handler.GetBatch)
			request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"items": [{
				"id": 42,
				"schedule_id": null,
				"status": "active",
				"name": "Kitchen",
				"description": "Clean counters",
				"deadline": "2020-02-03",
				"completed_on": null,
				"_links": [
					{"rel": "self", "href": "/chores/42"},
					{"rel": "collection", "href": "/chores"}
				]
			}],
			"more": false,
			"_links": [{"rel": "self", "href": "/chores"}],
			"_actions": [{"rel": "create", "href": "/chores/new"}]
		}`))
		},
	)

	It(
		"passes search and status filters to the HTML collection",
		func(ctx SpecContext) {
			deadline := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
				Status: choremodels.ChoreStatusCompleted,
				Search: "Kitchen",
				Offset: 4,
				Limit:  7,
			}).Return(&choremodels.ChorePage{
				Chores: []choremodels.Chore{{
					Id:          42,
					Status:      choremodels.ChoreStatusCompleted,
					Name:        "Kitchen",
					Description: "Wash and fold",
					Deadline:    deadline,
				}},
			}, nil).Once()
			view := NewMockView(GinkgoT())
			view.EXPECT().Collection(choreviewmodels.Collection{
				Items: []choreviewmodels.CollectionItem{{
					Chore: choreapimodels.Response{
						Id:          42,
						Status:      choremodels.ChoreStatusCompleted,
						Name:        "Kitchen",
						Description: "Wash and fold",
						Deadline:    "2026-08-15",
						Links: api.Relations{
							{Rel: "self", Href: "/chores/42"},
							{Rel: "collection", Href: "/chores"},
						},
					},
				}},
				Links: api.Relations{
					{
						Rel:  "self",
						Href: "/chores?status=completed&search=Kitchen&offset=4&limit=7",
					},
					{
						Rel:  "previous",
						Href: "/chores?limit=7&offset=0&search=Kitchen&status=completed",
					},
				},
				Actions: api.Relations{
					{Rel: "create", Href: "/chores/new"},
				},
				Status:          choremodels.ChoreStatusCompleted,
				Search:          "Kitchen",
				SearchOpen:      true,
				AutofocusSearch: true,
			}, false).Return(templ.Raw("collection sentinel")).Once()
			handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
				Service: service,
				View:    view,
			})
			e := echo.New()
			e.GET("/chores", handler.GetBatch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores?status=completed&search=Kitchen&offset=4&limit=7",
				nil,
			)
			request.Header.Set("Accept", "text/html")
			request.Header.Set("HX-Request", "true")
			request.Header.Set(
				"HX-Current-URL",
				"http://example.com/chores?status=completed&search=Kit",
			)
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(Equal("collection sentinel"))
		},
	)

	It("consumes the chore-and-template success flash", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
			Offset: 0,
			Limit:  20,
		}).Return(&choremodels.ChorePage{}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Collection(choreviewmodels.Collection{
			Items: []choreviewmodels.CollectionItem{},
			Links: api.Relations{{Rel: "self", Href: "/chores"}},
			Actions: api.Relations{
				{Rel: "create", Href: "/chores/new"},
			},
			Status: choremodels.ChoreStatusActive,
			Notice: "Chore added and template saved.",
		}, true).Return(templ.Raw("collection sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
		request.Header.Set("Accept", "text/html")
		request.AddCookie(&http.Cookie{
			Name:  "shiftbell_flash",
			Value: "chore-and-template-created",
		})
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("collection sentinel"))
		cookies := response.Result().Cookies()
		Expect(cookies).To(HaveLen(1))
		Expect(cookies[0].Name).To(Equal("shiftbell_flash"))
		Expect(cookies[0].MaxAge).To(Equal(-1))
	})

	It("consumes the chore completion success flash", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
			Offset: 0,
			Limit:  20,
		}).Return(&choremodels.ChorePage{}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Collection(choreviewmodels.Collection{
			Items: []choreviewmodels.CollectionItem{},
			Links: api.Relations{{Rel: "self", Href: "/chores"}},
			Actions: api.Relations{
				{Rel: "create", Href: "/chores/new"},
			},
			Status: choremodels.ChoreStatusActive,
			Notice: "Chore completed.",
		}, true).Return(templ.Raw("collection sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
		request.Header.Set("Accept", "text/html")
		request.AddCookie(&http.Cookie{
			Name:  "shiftbell_flash",
			Value: "chore-completed",
		})
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("collection sentinel"))
		cookies := response.Result().Cookies()
		Expect(cookies).To(HaveLen(1))
		Expect(cookies[0].Name).To(Equal("shiftbell_flash"))
		Expect(cookies[0].MaxAge).To(Equal(-1))
	})

	It("consumes the chore deletion success flash", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		service.EXPECT().Browse(ctx, &choremodels.BrowseChoresParams{
			Offset: 0,
			Limit:  20,
		}).Return(&choremodels.ChorePage{}, nil).Once()
		view := NewMockView(GinkgoT())
		view.EXPECT().Collection(choreviewmodels.Collection{
			Items: []choreviewmodels.CollectionItem{},
			Links: api.Relations{{Rel: "self", Href: "/chores"}},
			Actions: api.Relations{
				{Rel: "create", Href: "/chores/new"},
			},
			Status: choremodels.ChoreStatusActive,
			Notice: "Chore deleted.",
		}, true).Return(templ.Raw("collection sentinel")).Once()
		handler := choresendpoint.NewHandler(&choresendpoint.HandlerDeps{
			Service: service,
			View:    view,
		})
		e := echo.New()
		e.GET("/chores", handler.GetBatch)
		request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/chores", nil)
		request.Header.Set("Accept", "text/html")
		request.AddCookie(&http.Cookie{
			Name:  "shiftbell_flash",
			Value: "chore-deleted",
		})
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("collection sentinel"))
	})
})
