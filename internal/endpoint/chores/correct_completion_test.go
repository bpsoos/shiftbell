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

var _ = Describe("Correct chore completion", func() {
	It("rejects an HTML representation", func(ctx SpecContext) {
		service := NewMockService(GinkgoT())
		handler := choresendpoint.NewHandler(
			&choresendpoint.HandlerDeps{Service: service},
		)
		e := echo.New()
		e.PATCH("/chores/:id/completion", handler.CorrectCompletion)
		request := httptest.NewRequestWithContext(
			ctx,
			http.MethodPatch,
			"/chores/42/completion",
			strings.NewReader(`{"completed_on":"2020-02-04"}`),
		)
		request.Header.Set("Accept", "text/html")
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		e.ServeHTTP(response, request)

		Expect(response.Code).To(Equal(http.StatusNotAcceptable))
		Expect(response.Body.String()).To(BeEmpty())
	})

	It(
		"corrects a one-off completion through its advertised action",
		func(ctx SpecContext) {
			completedOn := time.Date(2020, time.February, 4, 0, 0, 0, 0, time.UTC)
			deadline := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
			service := NewMockService(GinkgoT())
			service.EXPECT().CorrectCompletion(ctx, &choremodels.CorrectCompletionParams{
				Id:          42,
				CompletedOn: completedOn,
			}).Return(&choremodels.ChoreDetails{
				Id:          42,
				Status:      choremodels.ChoreStatusCompleted,
				Name:        "Kitchen",
				Description: "Clean counters",
				Deadline:    deadline,
				CompletedOn: completedOn,
			}, nil).Once()
			handler := choresendpoint.NewHandler(
				&choresendpoint.HandlerDeps{Service: service},
			)
			e := echo.New()
			e.PATCH("/chores/:id/completion", handler.CorrectCompletion)
			request := httptest.NewRequestWithContext(
				ctx,
				http.MethodPatch,
				"/chores/42/completion",
				strings.NewReader(`{"completed_on":"2020-02-04"}`),
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
			"status": "completed",
			"name": "Kitchen",
			"description": "Clean counters",
			"deadline": "2020-02-01",
			"completed_on": "2020-02-04",
			"_links": {
				"self": {"href": "/chores/42"},
				"collection": {"href": "/chores"}
			},
			"_actions": {
				"correct_completion": {
					"href": "/chores/42/completion",
					"method": "PATCH",
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
