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

var _ = Describe("Get chore", func() {
	It(
		"returns an active manual one-off with its mutation controls",
		func(ctx SpecContext) {
			service := NewMockService(GinkgoT())
			service.EXPECT().Get(ctx, 42).Return(&choremodels.ChoreDetails{
				Id:          42,
				Status:      choremodels.ChoreStatusActive,
				Name:        "Kitchen",
				Description: "Wash and fold.",
				Deadline: time.Date(
					2020,
					time.February,
					3,
					0,
					0,
					0,
					0,
					time.UTC,
				),
			}, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			e := echo.New()
			e.GET("/chores/:id", handler.Get)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/chores/42",
				nil,
			)
			request.Header.Set("Accept", hypermedia.MediaType)
			response := httptest.NewRecorder()

			e.ServeHTTP(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Header().Get("Content-Type")).To(Equal(hypermedia.MediaType))
			Expect(response.Body.Bytes()).To(MatchJSON(`{
			"id": 42,
			"schedule_id": null,
			"status": "active",
			"name": "Kitchen",
			"description": "Wash and fold.",
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
