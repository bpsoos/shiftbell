package chores_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	choresendpoint "github.com/bpsoos/shiftbell/internal/endpoint/chores"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Edit chore", func() {
	It("rejects an HTML representation", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.PATCH("/chores/:id", handler.Patch)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPatch,
			"/chores/42",
			strings.NewReader(`{
				"name": "Kitchen",
				"deadline": "2020-02-03"
			}`),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
	})

	It(
		"edits an active one-off chore through its advertised action",
		func(ctx SpecContext) {
			deadline := time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
				Id:     42,
				Status: choremodels.ChoreStatusActive,
			}, nil).Once()
			service.EXPECT().Edit(ctx, &choremodels.EditChoreParams{
				Id:          42,
				Name:        "  Kitchen edited  ",
				Description: "  Clean all counters.  ",
				Deadline:    deadline,
			}).Return(&choremodels.ChoreDetails{
				Id:          42,
				Status:      choremodels.ChoreStatusActive,
				Name:        "Kitchen edited",
				Description: "Clean all counters.",
				Deadline:    deadline,
			}, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			e := echo.New()
			e.PATCH("/chores/:id", handler.Patch)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42",
				strings.NewReader(`{
				"name": "  Kitchen edited  ",
				"description": "  Clean all counters.  ",
				"deadline": "2020-02-03"
			}`),
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"id": 42,
			"schedule_id": null,
			"status": "active",
			"name": "Kitchen edited",
			"description": "Clean all counters.",
			"deadline": "2020-02-03",
			"completed_on": null,
			"_links": {
				"self": {"href": "/chores/42"},
				"collection": {"href": "/chores"}
			},
			"_actions": {
				"edit": {
					"href": "/chores/42",
					"method": "PATCH",
					"content_type": "application/json",
					"fields": [
						{"name": "name", "type": "string", "required": true},
						{"name": "description", "type": "string", "required": false},
						{"name": "deadline", "type": "date", "required": true}
					]
				},
				"complete": {
					"href": "/chores/42/completion",
					"method": "PUT",
					"content_type": "application/json",
					"fields": [
						{"name": "completed_on", "type": "date", "required": true}
					]
				},
				"delete": {
					"href": "/chores/42",
					"method": "DELETE",
					"content_type": "",
					"fields": null
				}
			}
		}`))
		},
	)
})
