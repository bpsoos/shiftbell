package chores_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Browse chores", func() {
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
				Status: choremodels.ChoreStatusActive,
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
				"_links": {
					"self": {"href": "/chores/42"},
					"collection": {"href": "/chores"}
				}
			}],
			"more": false,
			"_links": {"self": {"href": "/chores"}},
			"_actions": {
				"create": {
					"href": "/chores/new",
					"method": "GET",
					"content_type": "",
					"fields": null
				}
			}
		}`))
		},
	)
})
